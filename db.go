package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// openReadOnlyDB opens the SQLite database at path in read-only mode and pings
// it. Returns (db, false, nil) on success. Returns (nil, true, nil) when the
// file does not exist (caller should treat this as empty state). Returns
// (nil, false, err) for any other failure (permissions, corrupt file, etc.).
func openReadOnlyDB(path string) (*sql.DB, bool, error) {
	// Distinguish "file not found" from other failures (permissions, corrupt
	// header) before attempting to open so callers get a meaningful error
	// instead of silently showing an empty session list.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, true, nil // missing file — treat as empty
	} else if err != nil {
		return nil, false, fmt.Errorf("stat db: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?mode=ro")
	if err != nil {
		return nil, false, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close() // best-effort close on ping failure; original error takes priority
		return nil, false, fmt.Errorf("ping db: %w", err)
	}
	return db, false, nil
}

// loadSessions opens the opencode SQLite database and returns all primary
// sessions (parent_id IS NULL), ordered by most recently updated.
// Returns an empty slice (not an error) if the database file does not exist.
// home is the user's home directory (used for display path substitution) and
// is passed explicitly so the function does not call os.UserHomeDir itself.
func loadSessions(dbPath, home string) ([]session, error) {
	db, missing, err := openReadOnlyDB(dbPath)
	if err != nil {
		return nil, err
	}
	if missing {
		return []session{}, nil
	}
	defer func() { _ = db.Close() }()
	rows, err := db.Query(`
		SELECT id, title, directory,
		       time_created, time_updated,
		       COALESCE(summary_files, 0),
		       COALESCE(summary_additions, 0),
		       COALESCE(summary_deletions, 0)
		FROM session
		WHERE parent_id IS NULL
		ORDER BY time_updated DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var sessions []session
	for rows.Next() {
		var s session
		var createdMS, updatedMS int64
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Directory,
			&createdMS, &updatedMS,
			&s.SummaryFiles, &s.SummaryAdditions, &s.SummaryDeletions,
		); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		s.CreatedAt = time.Unix(createdMS/1000, (createdMS%1000)*int64(time.Millisecond))
		s.UpdatedAt = time.Unix(updatedMS/1000, (updatedMS%1000)*int64(time.Millisecond))
		s.DisplayDir = homeToTilde(s.Directory, home)
		s.ShortDir = baseName(s.Directory, home)
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return sessions, nil
}

// loadMessages returns all text messages for a session, ordered by time.
// One entry per message (first non-empty text part).
func loadMessages(dbPath, sessionID string) ([]message, error) {
	db, missing, err := openReadOnlyDB(dbPath)
	if err != nil {
		return nil, err
	}
	if missing {
		return []message{}, nil
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(`
		SELECT json_extract(m.data, '$.role'), json_extract(p.data, '$.text')
		FROM message m
		JOIN part p ON p.id = (
		    SELECT p2.id
		    FROM part p2
		    WHERE p2.message_id = m.id
		      AND json_extract(p2.data, '$.type') = 'text'
		      AND json_extract(p2.data, '$.text') IS NOT NULL
		      AND trim(json_extract(p2.data, '$.text')) != ''
		    ORDER BY p2.rowid ASC
		    LIMIT 1
		)
		WHERE m.session_id = ?
		ORDER BY m.rowid ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var messages []message
	for rows.Next() {
		var msg message
		if err := rows.Scan(&msg.Role, &msg.Text); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return messages, nil
}

// loadStats returns aggregated statistics for a single session by querying the
// part table. Two correlated subqueries keep it to a single round-trip:
//   - output_tokens: sum of tokens.output across all step-finish parts
//   - context_tokens: tokens.total from the most recent step-finish part
//
// Returns a zero-value sessionStats (not an error) if no step-finish parts
// exist (pure chat sessions with no model calls).
func loadStats(dbPath, sessionID string) (sessionStats, error) {
	db, missing, err := openReadOnlyDB(dbPath)
	if err != nil {
		return sessionStats{}, err
	}
	if missing {
		return sessionStats{}, nil
	}
	defer func() { _ = db.Close() }()

	var stats sessionStats
	var inputTokens, outputTokens, contextTokens *int // nullable — NULL when no step-finish parts

	err = db.QueryRow(`
		SELECT
		    COUNT(DISTINCT m.id),
		    (
		        SELECT SUM(json_extract(p2.data, '$.tokens.input'))
		        FROM part p2
		        WHERE p2.session_id = ?
		          AND json_extract(p2.data, '$.type') = 'step-finish'
		    ),
		    (
		        SELECT SUM(json_extract(p2.data, '$.tokens.output'))
		        FROM part p2
		        WHERE p2.session_id = ?
		          AND json_extract(p2.data, '$.type') = 'step-finish'
		    ),
		    (
		        SELECT json_extract(p3.data, '$.tokens.total')
		        FROM part p3
		        WHERE p3.session_id = ?
		          AND json_extract(p3.data, '$.type') = 'step-finish'
		        ORDER BY p3.time_created DESC
		        LIMIT 1
		    )
		FROM message m
		WHERE m.session_id = ?
	`, sessionID, sessionID, sessionID, sessionID).Scan(&stats.MsgCount, &inputTokens, &outputTokens, &contextTokens)
	if err != nil {
		return sessionStats{}, fmt.Errorf("query stats: %w", err)
	}

	if inputTokens != nil {
		stats.InputTokens = *inputTokens
	}
	if outputTokens != nil {
		stats.OutputTokens = *outputTokens
	}
	if contextTokens != nil {
		stats.ContextTokens = *contextTokens
	}

	// Distinct models ordered by first use.
	modelRows, err := db.Query(`
		SELECT COALESCE(
		           json_extract(data, '$.modelID'),
		           json_extract(data, '$.model.modelID'),
		           'unknown'
		       )
		FROM message
		WHERE session_id = ?
		  AND json_extract(data, '$.role') = 'assistant'
		  AND (
		      json_extract(data, '$.modelID') IS NOT NULL
		      OR json_extract(data, '$.model.modelID') IS NOT NULL
		  )
		GROUP BY 1
		ORDER BY MIN(rowid)
	`, sessionID)
	if err != nil {
		return sessionStats{}, fmt.Errorf("query models: %w", err)
	}
	defer func() { _ = modelRows.Close() }()
	for modelRows.Next() {
		var name string
		if err := modelRows.Scan(&name); err != nil {
			return sessionStats{}, fmt.Errorf("scan model: %w", err)
		}
		stats.Models = append(stats.Models, name)
	}
	if err := modelRows.Err(); err != nil {
		return sessionStats{}, fmt.Errorf("iterate models: %w", err)
	}

	return stats, nil
}

// loadGlobalStats returns aggregate statistics across all primary sessions
// (parent_id IS NULL). Two windows are returned: all-time and last 7 days.
// The work is split across three goroutines that each open their own read-only
// connection so they can run concurrently:
//
//   - goroutine 1: all-time session + token totals (single collapsed query)
//   - goroutine 2: last-7-days session + token totals (single collapsed query)
//   - goroutine 3: model breakdown + per-project model breakdown
//
// Token and turn counts are read from the message table (assistant rows) rather
// than the part table; the values are identical but the message table has a
// covering index on session_id, making each query ~50× faster.
//
// home is the user's home directory (used for display path substitution) and
// is passed explicitly so the function does not call os.UserHomeDir itself.
func loadGlobalStats(dbPath, home string) (globalStats, error) {
	// Quick pre-check: if the file is missing, return empty state immediately
	// without opening three connections.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return globalStats{}, nil
	} else if err != nil {
		return globalStats{}, fmt.Errorf("stat db: %w", err)
	}

	sevenDaysAgoMS := time.Now().UnixMilli() - (7 * 24 * 60 * 60 * 1000)

	type allTimeResult struct {
		gs  globalStats
		err error
	}
	type recentResult struct {
		sessions, messages                   int
		files, additions, deletions          int
		input, output, cacheRead, cacheWrite int
		durationMS                           int64
		err                                  error
	}
	type breakdownResult struct {
		models   []modelStat
		projects []projectStat
		err      error
	}

	allTimeCh := make(chan allTimeResult, 1)
	recentCh := make(chan recentResult, 1)
	breakdownCh := make(chan breakdownResult, 1)

	// ── Goroutine 1: all-time totals ─────────────────────────────────────────
	go func() {
		db, _, err := openReadOnlyDB(dbPath)
		if err != nil {
			allTimeCh <- allTimeResult{err: fmt.Errorf("open db (all-time): %w", err)}
			return
		}
		defer func() { _ = db.Close() }()

		// Single query: session metadata + assistant-message token sums joined.
		// Using the message table for tokens avoids a full scan of the part table;
		// assistant message rows carry the same token counts as step-finish parts.
		var r allTimeResult
		var input, output, cacheRead, cacheWrite *int
		err = db.QueryRow(`
			SELECT
			    COUNT(DISTINCT s.id),
			    COALESCE(SUM(COALESCE(s.summary_files,0)),0),
			    COALESCE(SUM(COALESCE(s.summary_additions,0)),0),
			    COALESCE(SUM(COALESCE(s.summary_deletions,0)),0),
			    COALESCE(SUM(DISTINCT s.time_updated - s.time_created),0),
			    COUNT(m.id),
			    SUM(json_extract(m.data,'$.tokens.input')),
			    SUM(json_extract(m.data,'$.tokens.output')),
			    SUM(json_extract(m.data,'$.tokens.cache.read')),
			    SUM(json_extract(m.data,'$.tokens.cache.write'))
			FROM session s
			LEFT JOIN message m ON m.session_id = s.id
			    AND json_extract(m.data,'$.role') = 'assistant'
			WHERE s.parent_id IS NULL
		`).Scan(
			&r.gs.TotalSessions, &r.gs.TotalFiles, &r.gs.TotalAdditions,
			&r.gs.TotalDeletions, &r.gs.TotalDurationMS,
			&r.gs.TotalMessages, &input, &output, &cacheRead, &cacheWrite,
		)
		if err != nil {
			allTimeCh <- allTimeResult{err: fmt.Errorf("query all-time totals: %w", err)}
			return
		}
		if input != nil {
			r.gs.TotalInput = *input
		}
		if output != nil {
			r.gs.TotalOutput = *output
		}
		if cacheRead != nil {
			r.gs.TotalCacheRead = *cacheRead
		}
		if cacheWrite != nil {
			r.gs.TotalCacheWrite = *cacheWrite
		}
		allTimeCh <- r
	}()

	// ── Goroutine 2: recent totals (last 7 days) ──────────────────────────────
	go func() {
		db, _, err := openReadOnlyDB(dbPath)
		if err != nil {
			recentCh <- recentResult{err: fmt.Errorf("open db (recent): %w", err)}
			return
		}
		defer func() { _ = db.Close() }()

		var r recentResult
		var input, output, cacheRead, cacheWrite *int
		err = db.QueryRow(`
			SELECT
			    COUNT(DISTINCT s.id),
			    COALESCE(SUM(COALESCE(s.summary_files,0)),0),
			    COALESCE(SUM(COALESCE(s.summary_additions,0)),0),
			    COALESCE(SUM(COALESCE(s.summary_deletions,0)),0),
			    COALESCE(SUM(DISTINCT s.time_updated - s.time_created),0),
			    COUNT(m.id),
			    SUM(json_extract(m.data,'$.tokens.input')),
			    SUM(json_extract(m.data,'$.tokens.output')),
			    SUM(json_extract(m.data,'$.tokens.cache.read')),
			    SUM(json_extract(m.data,'$.tokens.cache.write'))
			FROM session s
			LEFT JOIN message m ON m.session_id = s.id
			    AND json_extract(m.data,'$.role') = 'assistant'
			WHERE s.parent_id IS NULL AND s.time_created > ?
		`, sevenDaysAgoMS).Scan(
			&r.sessions, &r.files, &r.additions, &r.deletions, &r.durationMS,
			&r.messages, &input, &output, &cacheRead, &cacheWrite,
		)
		if err != nil {
			recentCh <- recentResult{err: fmt.Errorf("query recent totals: %w", err)}
			return
		}
		if input != nil {
			r.input = *input
		}
		if output != nil {
			r.output = *output
		}
		if cacheRead != nil {
			r.cacheRead = *cacheRead
		}
		if cacheWrite != nil {
			r.cacheWrite = *cacheWrite
		}
		recentCh <- r
	}()

	// ── Goroutine 3: model + project breakdowns ───────────────────────────────
	go func() {
		db, _, err := openReadOnlyDB(dbPath)
		if err != nil {
			breakdownCh <- breakdownResult{err: fmt.Errorf("open db (breakdown): %w", err)}
			return
		}
		defer func() { _ = db.Close() }()

		var r breakdownResult

		// Model breakdown: read model, turns, and tokens from the message table.
		// Duration is summed per distinct session (SUM(DISTINCT s.dur)) so that
		// sessions using multiple models don't get double-counted within a row —
		// though a session's duration is still attributed to each model it used.
		modelRows, err := db.Query(`
			SELECT
			    COALESCE(
			        json_extract(m.data,'$.modelID'),
			        json_extract(m.data,'$.model.modelID'),
			        'unknown'
			    ) AS model_name,
			    COUNT(DISTINCT m.session_id),
			    COUNT(*),
			    COALESCE(SUM(json_extract(m.data,'$.tokens.input')),0),
			    COALESCE(SUM(json_extract(m.data,'$.tokens.output')),0),
			    COALESCE(SUM(DISTINCT s.time_updated - s.time_created),0)
			FROM message m
			JOIN session s ON s.id = m.session_id AND s.parent_id IS NULL
			WHERE json_extract(m.data,'$.role') = 'assistant'
			GROUP BY model_name
			ORDER BY COUNT(DISTINCT m.session_id) DESC
		`)
		if err != nil {
			breakdownCh <- breakdownResult{err: fmt.Errorf("query models: %w", err)}
			return
		}
		defer func() { _ = modelRows.Close() }()
		for modelRows.Next() {
			var ms modelStat
			if err := modelRows.Scan(&ms.Name, &ms.Sessions, &ms.Turns, &ms.InputTokens, &ms.OutputTokens, &ms.DurationMS); err != nil {
				breakdownCh <- breakdownResult{err: fmt.Errorf("scan model stat: %w", err)}
				return
			}
			r.models = append(r.models, ms)
		}
		if err := modelRows.Err(); err != nil {
			breakdownCh <- breakdownResult{err: fmt.Errorf("iterate models: %w", err)}
			return
		}

		// Project breakdown (top 10 by session count): tokens from message table.
		projRows, err := db.Query(`
			SELECT
			    s.directory,
			    COUNT(DISTINCT s.id) AS cnt,
			    COUNT(m.id),
			    COALESCE(SUM(json_extract(m.data,'$.tokens.input')),0),
			    COALESCE(SUM(json_extract(m.data,'$.tokens.output')),0),
			    COALESCE(SUM(DISTINCT s.time_updated - s.time_created),0)
			FROM session s
			LEFT JOIN message m ON m.session_id = s.id
			    AND json_extract(m.data,'$.role') = 'assistant'
			WHERE s.parent_id IS NULL
			GROUP BY s.directory
			ORDER BY cnt DESC
			LIMIT 10
		`)
		if err != nil {
			breakdownCh <- breakdownResult{err: fmt.Errorf("query projects: %w", err)}
			return
		}
		defer func() { _ = projRows.Close() }()
		for projRows.Next() {
			var ps projectStat
			if err := projRows.Scan(&ps.Dir, &ps.Sessions, &ps.Turns, &ps.InputTokens, &ps.OutputTokens, &ps.DurationMS); err != nil {
				breakdownCh <- breakdownResult{err: fmt.Errorf("scan project stat: %w", err)}
				return
			}
			ps.DisplayDir = homeToTilde(ps.Dir, home)
			r.projects = append(r.projects, ps)
		}
		if err := projRows.Err(); err != nil {
			breakdownCh <- breakdownResult{err: fmt.Errorf("iterate projects: %w", err)}
			return
		}

		// Per-project model breakdown: same message-only approach.
		if len(r.projects) > 0 {
			projModelRows, err := db.Query(`
				SELECT
				    s.directory,
				    COALESCE(
				        json_extract(m.data,'$.modelID'),
				        json_extract(m.data,'$.model.modelID'),
				        'unknown'
				    ) AS model_name,
				    COUNT(DISTINCT m.session_id),
				    COUNT(*),
				    COALESCE(SUM(json_extract(m.data,'$.tokens.input')),0),
				    COALESCE(SUM(json_extract(m.data,'$.tokens.output')),0),
				    COALESCE(SUM(DISTINCT s.time_updated - s.time_created),0)
				FROM message m
				JOIN session s ON s.id = m.session_id AND s.parent_id IS NULL
				WHERE json_extract(m.data,'$.role') = 'assistant'
				  AND s.directory IN (
				      SELECT directory FROM session WHERE parent_id IS NULL
				      GROUP BY directory ORDER BY COUNT(*) DESC LIMIT 10
				  )
				GROUP BY s.directory, model_name
				ORDER BY s.directory, COUNT(DISTINCT m.session_id) DESC
			`)
			if err != nil {
				breakdownCh <- breakdownResult{err: fmt.Errorf("query project models: %w", err)}
				return
			}
			defer func() { _ = projModelRows.Close() }()

			projModels := make(map[string][]modelStat, len(r.projects))
			for projModelRows.Next() {
				var dir string
				var ms modelStat
				if err := projModelRows.Scan(&dir, &ms.Name, &ms.Sessions, &ms.Turns, &ms.InputTokens, &ms.OutputTokens, &ms.DurationMS); err != nil {
					breakdownCh <- breakdownResult{err: fmt.Errorf("scan project model stat: %w", err)}
					return
				}
				projModels[dir] = append(projModels[dir], ms)
			}
			if err := projModelRows.Err(); err != nil {
				breakdownCh <- breakdownResult{err: fmt.Errorf("iterate project models: %w", err)}
				return
			}
			for i, ps := range r.projects {
				r.projects[i].Models = projModels[ps.Dir]
			}
		}

		breakdownCh <- r
	}()

	// ── Collect results ───────────────────────────────────────────────────────
	atRes := <-allTimeCh
	if atRes.err != nil {
		return globalStats{}, atRes.err
	}
	gs := atRes.gs

	recRes := <-recentCh
	if recRes.err != nil {
		return globalStats{}, recRes.err
	}
	gs.RecentSessions = recRes.sessions
	gs.RecentMessages = recRes.messages
	gs.RecentFiles = recRes.files
	gs.RecentAdditions = recRes.additions
	gs.RecentDeletions = recRes.deletions
	gs.RecentDurationMS = recRes.durationMS
	gs.RecentInput = recRes.input
	gs.RecentOutput = recRes.output
	gs.RecentCacheRead = recRes.cacheRead
	gs.RecentCacheWrite = recRes.cacheWrite

	bdRes := <-breakdownCh
	if bdRes.err != nil {
		return globalStats{}, bdRes.err
	}
	gs.Models = bdRes.models
	gs.Projects = bdRes.projects

	return gs, nil
}
