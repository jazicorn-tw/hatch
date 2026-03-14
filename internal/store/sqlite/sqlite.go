package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"encoding/binary"
	"fmt"
	"math"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo" // registers sqlite-vec with go-sqlite3
	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
	_ "github.com/mattn/go-sqlite3" // registers the "sqlite3" driver with database/sql
)

func init() {
	// Auto-load the sqlite-vec extension into every new go-sqlite3 connection.
	sqlite_vec.Auto()
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// float32Size is the byte width of a single float32 value.
const float32Size = 4

// Store is a SQLite-backed implementation of store.VecStore.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at path with WAL mode enabled,
// then runs any pending migrations.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
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
	sqlBytes, err := migrationsFS.ReadFile("migrations/" + version)
	if err != nil {
		return fmt.Errorf("sqlite: read migration %s: %w", version, err)
	}
	if _, err := s.db.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("sqlite: apply migration %s: %w", version, err)
	}
	if _, err := s.db.Exec(
		`INSERT INTO schema_migrations (version) VALUES (?)`, version,
	); err != nil {
		return fmt.Errorf("sqlite: record migration %s: %w", version, err)
	}
	return nil
}

// upsertStmts holds the prepared statements needed for a single Upsert transaction.
type upsertStmts struct {
	chunk     *sql.Stmt
	vecDelete *sql.Stmt
	vecInsert *sql.Stmt
}

// close releases all prepared statements.
func (u *upsertStmts) close() {
	u.chunk.Close()
	u.vecDelete.Close()
	u.vecInsert.Close()
}

// prepareUpsertStmts prepares the three statements required by Upsert within tx.
// The caller must call close() on the returned value when done.
func prepareUpsertStmts(ctx context.Context, tx *sql.Tx) (*upsertStmts, error) {
	chunk, err := tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO chunks (id, source, text, embedding) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: prepare chunk insert: %w", err)
	}
	// vec0 virtual tables do not support INSERT OR REPLACE; use DELETE then INSERT.
	vecDelete, err := tx.PrepareContext(ctx,
		`DELETE FROM vec_chunks WHERE chunk_id = ?`)
	if err != nil {
		chunk.Close()
		return nil, fmt.Errorf("sqlite: prepare vec delete: %w", err)
	}
	vecInsert, err := tx.PrepareContext(ctx,
		`INSERT INTO vec_chunks (chunk_id, embedding) VALUES (?, ?)`)
	if err != nil {
		chunk.Close()
		vecDelete.Close()
		return nil, fmt.Errorf("sqlite: prepare vec insert: %w", err)
	}
	return &upsertStmts{chunk: chunk, vecDelete: vecDelete, vecInsert: vecInsert}, nil
}

// execUpsertRecord writes a single record using pre-prepared statements.
func execUpsertRecord(ctx context.Context, stmts *upsertStmts, r store.Record) error {
	blob := encodeVec(r.Embedding)
	if _, err := stmts.chunk.ExecContext(ctx, r.Chunk.ID, r.Chunk.Source, r.Chunk.Text, blob); err != nil {
		return fmt.Errorf("sqlite: insert chunk %s: %w", r.Chunk.ID, err)
	}
	if len(r.Embedding) == 0 {
		return nil
	}
	if _, err := stmts.vecDelete.ExecContext(ctx, r.Chunk.ID); err != nil {
		return fmt.Errorf("sqlite: delete vec %s: %w", r.Chunk.ID, err)
	}
	if _, err := stmts.vecInsert.ExecContext(ctx, r.Chunk.ID, blob); err != nil {
		return fmt.Errorf("sqlite: insert vec %s: %w", r.Chunk.ID, err)
	}
	return nil
}

// Add inserts records into the store. Delegates to Upsert for store.Store compatibility.
func (s *Store) Add(ctx context.Context, records []store.Record) error {
	return s.Upsert(ctx, records)
}

// Upsert inserts or replaces records in both the chunks and vec_chunks tables.
// Records are keyed by Chunk.ID; existing rows are replaced atomically.
func (s *Store) Upsert(ctx context.Context, records []store.Record) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmts, err := prepareUpsertStmts(ctx, tx)
	if err != nil {
		return err
	}
	defer stmts.close()

	for _, r := range records {
		if err := execUpsertRecord(ctx, stmts, r); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DeleteBySource removes all records whose source matches the given value.
func (s *Store) DeleteBySource(ctx context.Context, source string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM vec_chunks WHERE chunk_id IN (SELECT id FROM chunks WHERE source = ?)`,
		source,
	); err != nil {
		return fmt.Errorf("sqlite: delete vec_chunks for source %q: %w", source, err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM chunks WHERE source = ?`, source,
	); err != nil {
		return fmt.Errorf("sqlite: delete chunks for source %q: %w", source, err)
	}
	return tx.Commit()
}

// Search returns the k nearest records using sqlite-vec KNN search.
func (s *Store) Search(ctx context.Context, vec []float32, k int) ([]store.Record, error) {
	blob := encodeVec(vec)
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.source, c.text, c.embedding
		 FROM vec_chunks v
		 JOIN chunks c ON c.id = v.chunk_id
		 WHERE v.embedding MATCH ? AND k = ?
		 ORDER BY v.distance`,
		blob, k,
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite: knn search: %w", err)
	}
	defer rows.Close()

	var records []store.Record
	for rows.Next() {
		var id, src, text string
		var embBlob []byte
		if err := rows.Scan(&id, &src, &text, &embBlob); err != nil {
			return nil, fmt.Errorf("sqlite: scan row: %w", err)
		}
		records = append(records, store.Record{
			Chunk:     chunker.Chunk{ID: id, Source: src, Text: text},
			Embedding: decodeVec(embBlob),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: rows: %w", err)
	}
	return records, nil
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
