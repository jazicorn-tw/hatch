CREATE TABLE IF NOT EXISTS question_bank (
    id            TEXT    PRIMARY KEY,
    topic         TEXT    NOT NULL,
    question_text TEXT    NOT NULL,
    options       TEXT    NOT NULL,  -- JSON array of 4 strings
    correct_index INTEGER NOT NULL,
    explanation   TEXT    NOT NULL DEFAULT '',
    created_at    DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_question_bank_topic ON question_bank(topic);

CREATE TABLE IF NOT EXISTS kata_bank (
    id           TEXT    PRIMARY KEY,
    topic        TEXT    NOT NULL,
    title        TEXT    NOT NULL,
    description  TEXT    NOT NULL DEFAULT '',
    starter_code TEXT    NOT NULL DEFAULT '',
    tests        TEXT    NOT NULL DEFAULT '',
    language     TEXT    NOT NULL DEFAULT '',
    created_at   DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_kata_bank_topic ON kata_bank(topic);
