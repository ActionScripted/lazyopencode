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
func loadSessions(dbPath string) ([]Session, error) {
	db, missing, err := openReadOnlyDB(dbPath)
	if err != nil {
		return nil, err
	}
	if missing {
		return []Session{}, nil
	}
	defer func() { _ = db.Close() }()
	home := resolveHome()
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

	var sessions []Session
	for rows.Next() {
		var s Session
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
func loadMessages(dbPath, sessionID string) ([]Message, error) {
	db, missing, err := openReadOnlyDB(dbPath)
	if err != nil {
		return nil, err
	}
	if missing {
		return []Message{}, nil
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

	var messages []Message
	for rows.Next() {
		var msg Message
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
// Returns a zero-value SessionStats (not an error) if no step-finish parts
// exist (pure chat sessions with no model calls).
func loadStats(dbPath, sessionID string) (SessionStats, error) {
	db, missing, err := openReadOnlyDB(dbPath)
	if err != nil {
		return SessionStats{}, err
	}
	if missing {
		return SessionStats{}, nil
	}
	defer func() { _ = db.Close() }()

	var stats SessionStats
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
		return SessionStats{}, fmt.Errorf("query stats: %w", err)
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
		return SessionStats{}, fmt.Errorf("query models: %w", err)
	}
	defer func() { _ = modelRows.Close() }()
	for modelRows.Next() {
		var name string
		if err := modelRows.Scan(&name); err != nil {
			return SessionStats{}, fmt.Errorf("scan model: %w", err)
		}
		stats.Models = append(stats.Models, name)
	}
	if err := modelRows.Err(); err != nil {
		return SessionStats{}, fmt.Errorf("iterate models: %w", err)
	}

	return stats, nil
}
