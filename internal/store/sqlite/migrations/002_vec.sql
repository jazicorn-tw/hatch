-- 002_vec.sql: add sqlite-vec KNN virtual table for efficient vector search.
-- Requires the sqlite-vec extension to be loaded before this migration runs.
-- The extension is registered via sqlite_vec.Auto() in the sqlite package init.

CREATE VIRTUAL TABLE IF NOT EXISTS vec_chunks USING vec0(
    chunk_id TEXT PRIMARY KEY,
    embedding float[768]
);
