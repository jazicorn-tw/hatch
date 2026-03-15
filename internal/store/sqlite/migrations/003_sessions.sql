CREATE TABLE IF NOT EXISTS quiz_sessions (
    id         TEXT PRIMARY KEY,
    topic      TEXT    NOT NULL,
    score      INTEGER NOT NULL DEFAULT 0,
    total      INTEGER NOT NULL DEFAULT 0,
    started_at DATETIME NOT NULL DEFAULT (datetime('now')),
    ended_at   DATETIME
);

CREATE TABLE IF NOT EXISTS quiz_questions (
    id            TEXT    PRIMARY KEY,
    session_id    TEXT    NOT NULL REFERENCES quiz_sessions(id) ON DELETE CASCADE,
    question_text TEXT    NOT NULL,
    options       TEXT    NOT NULL,  -- JSON array of 4 strings
    correct_index INTEGER NOT NULL,
    explanation   TEXT    NOT NULL DEFAULT '',
    user_answer   INTEGER,           -- NULL if unanswered, 0-3
    created_at    DATETIME NOT NULL DEFAULT (datetime('now'))
);
