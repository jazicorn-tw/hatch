package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// float32Size is the byte width of a single float32 value.
const float32Size = 4

// Store is a SQLite-backed implementation of store.Store.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at path with WAL mode enabled,
// then runs any pending migrations.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %s: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// migrate runs pending schema migrations.
func (s *Store) migrate() error {
	if err := s.ensureMigrationsTable(); err != nil {
		return err
	}
	return s.runMigrations()
}

// ensureMigrationsTable creates the schema_migrations tracking table if absent.
func (s *Store) ensureMigrationsTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("sqlite: ensure migrations table: %w", err)
	}
	return nil
}

// runMigrations applies all *.sql files in migrations/ not yet recorded.
func (s *Store) runMigrations() error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("sqlite: read migrations dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if err := s.applyMigration(e.Name()); err != nil {
			return err
		}
	}
	return nil
}

// applyMigration applies a single migration file if not already recorded.
func (s *Store) applyMigration(version string) error {
	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version,
	).Scan(&count); err != nil {
		return fmt.Errorf("sqlite: check migration %s: %w", version, err)
	}
	if count > 0 {
		return nil
	}
	sql, err := migrationsFS.ReadFile("migrations/" + version)
	if err != nil {
		return fmt.Errorf("sqlite: read migration %s: %w", version, err)
	}
	if _, err := s.db.Exec(string(sql)); err != nil {
		return fmt.Errorf("sqlite: apply migration %s: %w", version, err)
	}
	if _, err := s.db.Exec(
		`INSERT INTO schema_migrations (version) VALUES (?)`, version,
	); err != nil {
		return fmt.Errorf("sqlite: record migration %s: %w", version, err)
	}
	return nil
}

// Add inserts records into the chunks table.
func (s *Store) Add(ctx context.Context, records []store.Record) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO chunks (id, source, text, embedding) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("sqlite: prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, r := range records {
		blob := encodeVec(r.Embedding)
		if _, err := stmt.ExecContext(ctx, r.Chunk.ID, r.Chunk.Source, r.Chunk.Text, blob); err != nil {
			return fmt.Errorf("sqlite: insert chunk %s: %w", r.Chunk.ID, err)
		}
	}
	return tx.Commit()
}

// Search returns the k nearest records using min-heap top-k selection
// with cosine similarity — O(n log k) vs a naive O(n log n) full sort.
func (s *Store) Search(ctx context.Context, vec []float32, k int) ([]store.Record, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, source, text, embedding FROM chunks`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: search query: %w", err)
	}
	defer rows.Close()

	var records []store.Record
	for rows.Next() {
		var id, src, text string
		var blob []byte
		if err := rows.Scan(&id, &src, &text, &blob); err != nil {
			return nil, fmt.Errorf("sqlite: scan row: %w", err)
		}
		records = append(records, store.Record{
			Chunk:     chunker.Chunk{ID: id, Source: src, Text: text},
			Embedding: decodeVec(blob),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: rows: %w", err)
	}

	return store.TopK(records, vec, k), nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error { return s.db.Close() }

// encodeVec serialises a float32 slice as little-endian bytes.
func encodeVec(v []float32) []byte {
	b := make([]byte, len(v)*float32Size)
	for i, f := range v {
		binary.LittleEndian.PutUint32(b[i*float32Size:], math.Float32bits(f))
	}
	return b
}

// decodeVec deserialises little-endian bytes back to a float32 slice.
// Returns nil if the byte slice length is not a multiple of float32Size.
func decodeVec(b []byte) []float32 {
	if len(b)%float32Size != 0 {
		return nil
	}
	v := make([]float32, len(b)/float32Size)
	for i := range v {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*float32Size:]))
	}
	return v
}
