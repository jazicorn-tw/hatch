CREATE TABLE IF NOT EXISTS kata_sessions (
    id         TEXT    PRIMARY KEY,
    topic      TEXT    NOT NULL,
    kata_id    TEXT    NOT NULL,
    language   TEXT    NOT NULL DEFAULT '',
    passed     INTEGER NOT NULL DEFAULT 0,
    attempts   INTEGER NOT NULL DEFAULT 0,
    started_at DATETIME NOT NULL DEFAULT (datetime('now')),
    ended_at   DATETIME
);
