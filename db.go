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
// Non-fatal: returns a zero-value globalStats on any query error rather than
// surfacing DB errors to the stats view.
// home is the user's home directory (used for display path substitution) and
// is passed explicitly so the function does not call os.UserHomeDir itself.
func loadGlobalStats(dbPath, home string) (globalStats, error) {
	db, missing, err := openReadOnlyDB(dbPath)
	if err != nil {
		return globalStats{}, err
	}
	if missing {
		return globalStats{}, nil
	}
	defer func() { _ = db.Close() }()

	sevenDaysAgoMS := (time.Now().UnixMilli()) - (7 * 24 * 60 * 60 * 1000)

	var gs globalStats

	// ── All-time session totals (session table only) ──────────────────────────
	if err := db.QueryRow(`
		SELECT COUNT(*),
		       COALESCE(SUM(COALESCE(summary_files,0)),0),
		       COALESCE(SUM(COALESCE(summary_additions,0)),0),
		       COALESCE(SUM(COALESCE(summary_deletions,0)),0),
		       COALESCE(SUM(time_updated - time_created),0)
		FROM session WHERE parent_id IS NULL
	`).Scan(&gs.TotalSessions, &gs.TotalFiles, &gs.TotalAdditions, &gs.TotalDeletions, &gs.TotalDurationMS); err != nil {
		return globalStats{}, fmt.Errorf("query total sessions: %w", err)
	}

	// ── All-time token + message totals (part step-finish) ───────────────────
	var totalInput, totalOutput, totalCacheRead, totalCacheWrite *int
	if err := db.QueryRow(`
		SELECT COUNT(*),
		       SUM(json_extract(p.data,'$.tokens.input')),
		       SUM(json_extract(p.data,'$.tokens.output')),
		       SUM(json_extract(p.data,'$.tokens.cache.read')),
		       SUM(json_extract(p.data,'$.tokens.cache.write'))
		FROM part p
		WHERE json_extract(p.data,'$.type') = 'step-finish'
		  AND p.session_id IN (SELECT id FROM session WHERE parent_id IS NULL)
	`).Scan(&gs.TotalMessages, &totalInput, &totalOutput, &totalCacheRead, &totalCacheWrite); err != nil {
		return globalStats{}, fmt.Errorf("query total tokens: %w", err)
	}
	if totalInput != nil {
		gs.TotalInput = *totalInput
	}
	if totalOutput != nil {
		gs.TotalOutput = *totalOutput
	}
	if totalCacheRead != nil {
		gs.TotalCacheRead = *totalCacheRead
	}
	if totalCacheWrite != nil {
		gs.TotalCacheWrite = *totalCacheWrite
	}

	// ── Recent session totals (last 7 days) ───────────────────────────────────
	if err := db.QueryRow(`
		SELECT COUNT(*),
		       COALESCE(SUM(COALESCE(summary_files,0)),0),
		       COALESCE(SUM(COALESCE(summary_additions,0)),0),
		       COALESCE(SUM(COALESCE(summary_deletions,0)),0),
		       COALESCE(SUM(time_updated - time_created),0)
		FROM session WHERE parent_id IS NULL AND time_created > ?
	`, sevenDaysAgoMS).Scan(&gs.RecentSessions, &gs.RecentFiles, &gs.RecentAdditions, &gs.RecentDeletions, &gs.RecentDurationMS); err != nil {
		return globalStats{}, fmt.Errorf("query recent sessions: %w", err)
	}

	var recentInput, recentOutput, recentCacheRead, recentCacheWrite *int
	if err := db.QueryRow(`
		SELECT COUNT(*),
		       SUM(json_extract(p.data,'$.tokens.input')),
		       SUM(json_extract(p.data,'$.tokens.output')),
		       SUM(json_extract(p.data,'$.tokens.cache.read')),
		       SUM(json_extract(p.data,'$.tokens.cache.write'))
		FROM part p
		WHERE json_extract(p.data,'$.type') = 'step-finish'
		  AND p.session_id IN (SELECT id FROM session WHERE parent_id IS NULL AND time_created > ?)
	`, sevenDaysAgoMS).Scan(&gs.RecentMessages, &recentInput, &recentOutput, &recentCacheRead, &recentCacheWrite); err != nil {
		return globalStats{}, fmt.Errorf("query recent tokens: %w", err)
	}
	if recentInput != nil {
		gs.RecentInput = *recentInput
	}
	if recentOutput != nil {
		gs.RecentOutput = *recentOutput
	}
	if recentCacheRead != nil {
		gs.RecentCacheRead = *recentCacheRead
	}
	if recentCacheWrite != nil {
		gs.RecentCacheWrite = *recentCacheWrite
	}

	// ── Model breakdown ───────────────────────────────────────────────────────
	modelRows, err := db.Query(`
		SELECT COALESCE(
		           json_extract(m.data,'$.modelID'),
		           json_extract(m.data,'$.model.modelID'),
		           'unknown'
		       ) AS model_name,
		       COUNT(DISTINCT m.session_id),
		       COUNT(p.id),
		       COALESCE(SUM(json_extract(p.data,'$.tokens.input')),0),
		       COALESCE(SUM(json_extract(p.data,'$.tokens.output')),0),
		       COALESCE((
		           SELECT SUM(s.time_updated - s.time_created)
		           FROM session s
		           WHERE s.parent_id IS NULL
		             AND s.id IN (
		                 SELECT DISTINCT m2.session_id
		                 FROM message m2
		                 WHERE json_extract(m2.data,'$.role') = 'assistant'
		                   AND COALESCE(
		                           json_extract(m2.data,'$.modelID'),
		                           json_extract(m2.data,'$.model.modelID'),
		                           'unknown'
		                       ) = model_name
		                   AND m2.session_id IN (SELECT id FROM session WHERE parent_id IS NULL)
		             )
		       ), 0)
		FROM message m
		JOIN part p ON p.session_id = m.session_id
		          AND json_extract(p.data,'$.type') = 'step-finish'
		WHERE json_extract(m.data,'$.role') = 'assistant'
		  AND m.session_id IN (SELECT id FROM session WHERE parent_id IS NULL)
		GROUP BY 1
		ORDER BY 2 DESC
	`)
	if err != nil {
		return globalStats{}, fmt.Errorf("query models: %w", err)
	}
	defer func() { _ = modelRows.Close() }()
	for modelRows.Next() {
		var ms modelStat
		if err := modelRows.Scan(&ms.Name, &ms.Sessions, &ms.Turns, &ms.InputTokens, &ms.OutputTokens, &ms.DurationMS); err != nil {
			return globalStats{}, fmt.Errorf("scan model stat: %w", err)
		}
		gs.Models = append(gs.Models, ms)
	}
	if err := modelRows.Err(); err != nil {
		return globalStats{}, fmt.Errorf("iterate models: %w", err)
	}

	// ── Project breakdown (top 10 by session count) ───────────────────────────
	projRows, err := db.Query(`
		SELECT s.directory,
		       COUNT(DISTINCT s.id) AS cnt,
		       COUNT(p.id),
		       COALESCE(SUM(json_extract(p.data,'$.tokens.input')),0),
		       COALESCE(SUM(json_extract(p.data,'$.tokens.output')),0),
		       COALESCE(SUM(s.time_updated - s.time_created),0)
		FROM session s
		LEFT JOIN part p ON p.session_id = s.id
		    AND json_extract(p.data,'$.type') = 'step-finish'
		WHERE s.parent_id IS NULL
		GROUP BY s.directory
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err != nil {
		return globalStats{}, fmt.Errorf("query projects: %w", err)
	}
	defer func() { _ = projRows.Close() }()
	for projRows.Next() {
		var ps projectStat
		if err := projRows.Scan(&ps.Dir, &ps.Sessions, &ps.Turns, &ps.InputTokens, &ps.OutputTokens, &ps.DurationMS); err != nil {
			return globalStats{}, fmt.Errorf("scan project stat: %w", err)
		}
		ps.DisplayDir = homeToTilde(ps.Dir, home)
		gs.Projects = append(gs.Projects, ps)
	}
	if err := projRows.Err(); err != nil {
		return globalStats{}, fmt.Errorf("iterate projects: %w", err)
	}

	// ── Per-project model breakdown ───────────────────────────────────────────
	// Build a set of the top-10 project directories so we can filter the query.
	if len(gs.Projects) > 0 {
		projModelRows, err := db.Query(`
			SELECT s.directory,
			       COALESCE(
			           json_extract(m.data,'$.modelID'),
			           json_extract(m.data,'$.model.modelID'),
			           'unknown'
			       ) AS model_name,
			       COUNT(DISTINCT s.id),
			       COUNT(p.id),
			       COALESCE(SUM(json_extract(p.data,'$.tokens.input')),0),
			       COALESCE(SUM(json_extract(p.data,'$.tokens.output')),0),
			       COALESCE((
			           SELECT SUM(s2.time_updated - s2.time_created)
			           FROM session s2
			           WHERE s2.parent_id IS NULL
			             AND s2.directory = s.directory
			             AND s2.id IN (
			                 SELECT DISTINCT m2.session_id
			                 FROM message m2
			                 WHERE json_extract(m2.data,'$.role') = 'assistant'
			                   AND COALESCE(
			                           json_extract(m2.data,'$.modelID'),
			                           json_extract(m2.data,'$.model.modelID'),
			                           'unknown'
			                       ) = model_name
			                   AND m2.session_id IN (SELECT id FROM session WHERE parent_id IS NULL)
			             )
			       ), 0)
			FROM session s
			JOIN message m ON m.session_id = s.id
			    AND json_extract(m.data,'$.role') = 'assistant'
			JOIN part p ON p.session_id = s.id
			    AND json_extract(p.data,'$.type') = 'step-finish'
			WHERE s.parent_id IS NULL
			  AND s.directory IN (
			      SELECT directory FROM session WHERE parent_id IS NULL
			      GROUP BY directory ORDER BY COUNT(*) DESC LIMIT 10
			  )
			GROUP BY s.directory, model_name
			ORDER BY s.directory, 3 DESC
		`)
		if err != nil {
			return globalStats{}, fmt.Errorf("query project models: %w", err)
		}
		defer func() { _ = projModelRows.Close() }()

		// Build map[dir][]modelStat then attach to each projectStat.
		projModels := make(map[string][]modelStat, len(gs.Projects))
		for projModelRows.Next() {
			var dir string
			var ms modelStat
			if err := projModelRows.Scan(&dir, &ms.Name, &ms.Sessions, &ms.Turns, &ms.InputTokens, &ms.OutputTokens, &ms.DurationMS); err != nil {
				return globalStats{}, fmt.Errorf("scan project model stat: %w", err)
			}
			projModels[dir] = append(projModels[dir], ms)
		}
		if err := projModelRows.Err(); err != nil {
			return globalStats{}, fmt.Errorf("iterate project models: %w", err)
		}
		for i, ps := range gs.Projects {
			gs.Projects[i].Models = projModels[ps.Dir]
		}
	}

	return gs, nil
}
