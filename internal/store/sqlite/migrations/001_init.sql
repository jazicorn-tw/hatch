-- 001_init.sql: initial schema

CREATE TABLE IF NOT EXISTS schema_migrations (
    version     TEXT PRIMARY KEY,
    applied_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS chunks (
    id          TEXT PRIMARY KEY,
    source      TEXT NOT NULL,
    text        TEXT NOT NULL,
    embedding   BLOB,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
