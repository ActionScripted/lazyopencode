package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// loadSessions opens the opencode SQLite database and returns all primary
// sessions (parent_id IS NULL), ordered by most recently updated.
// Returns an empty slice (not an error) if the database file does not exist.
func loadSessions(dbPath string) ([]Session, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// Ping to detect missing file early — sqlite returns an error here if the
	// file doesn't exist when opened read-only.
	if err := db.Ping(); err != nil {
		// Treat missing DB as empty state, not a fatal error.
		return []Session{}, nil
	}

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
	defer rows.Close()

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
		s.CreatedAt = time.Unix(createdMS/1000, 0)
		s.UpdatedAt = time.Unix(updatedMS/1000, 0)
		s.DisplayDir = displayDir(s.Directory)
		s.ShortDir = shortDir(s.Directory)
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
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return []Message{}, nil
	}

	rows, err := db.Query(`
		SELECT json_extract(m.data, '$.role'), json_extract(p.data, '$.text')
		FROM message m
		JOIN part p ON p.message_id = m.id
		WHERE m.session_id = ?
		  AND json_extract(p.data, '$.type') = 'text'
		  AND json_extract(p.data, '$.text') IS NOT NULL
		  AND trim(json_extract(p.data, '$.text')) != ''
		GROUP BY m.id
		ORDER BY p.time_created
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

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
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return SessionStats{}, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return SessionStats{}, nil
	}

	var stats SessionStats
	var outputTokens, contextTokens *int // nullable — NULL when no step-finish parts

	err = db.QueryRow(`
		SELECT
		    COUNT(DISTINCT m.id),
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
	`, sessionID, sessionID, sessionID).Scan(&stats.MsgCount, &outputTokens, &contextTokens)
	if err != nil {
		return SessionStats{}, fmt.Errorf("query stats: %w", err)
	}

	if outputTokens != nil {
		stats.OutputTokens = *outputTokens
	}
	if contextTokens != nil {
		stats.ContextTokens = *contextTokens
	}

	return stats, nil
}
