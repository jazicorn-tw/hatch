package sqlite

// Tests covering error paths in sqlite.go that require special DB states.

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
)

// openRawDB opens a raw *sql.DB without running migrations.
func openRawDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "raw.db")
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// TestEnsureMigrationsTableError covers the error path in ensureMigrationsTable
// by using a closed DB.
func TestEnsureMigrationsTableError(t *testing.T) {
	db := openRawDB(t)
	db.Close() // close so Exec fails
	s := &Store{db: db}
	if err := s.ensureMigrationsTable(); err == nil {
		t.Error("expected error on closed DB")
	}
}

// TestMigrateEnsureTableError covers the migrate() error path (line 52-54).
func TestMigrateEnsureTableError(t *testing.T) {
	db := openRawDB(t)
	db.Close()
	s := &Store{db: db}
	if err := s.migrate(); err == nil {
		t.Error("expected error when ensureMigrationsTable fails")
	}
}

// TestApplyMigrationQueryError covers the QueryRow.Scan error in applyMigration
// by closing the DB after creating the schema_migrations table.
func TestApplyMigrationQueryError(t *testing.T) {
	db := openRawDB(t)
	s := &Store{db: db}
	// Create the migrations table so ensureMigrationsTable would succeed.
	if err := s.ensureMigrationsTable(); err != nil {
		t.Fatal(err)
	}
	db.Close() // now close so QueryRow fails
	if err := s.applyMigration("001_init.sql"); err == nil {
		t.Error("expected error when DB is closed during QueryRow")
	}
}

// TestPrepareUpsertStmtsSecondError covers line 138-141 in prepareUpsertStmts
// (second Prepare fails). We use a context already cancelled after the
// first PrepareContext call by doing two sequential prepare attempts:
// first succeeds, second is the vec DELETE which fails with cancelled ctx.
func TestPrepareUpsertStmtsSecondPrepareError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Prepare the first statement manually so it succeeds.
	_, err = tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO chunks (id, source, text, embedding) VALUES (?, ?, ?, ?)`)
	if err != nil {
		t.Fatal(err)
	}

	// Now cancel the context and try prepareUpsertStmts which will fail on
	// the first (chunk) prepare with the cancelled context.
	ctx2, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = prepareUpsertStmts(ctx2, tx)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

// TestExecUpsertRecordVecDeleteError covers the vecDelete error path (line 161-163)
// by setting up a record with a non-empty embedding and then rolling back the tx.
func TestExecUpsertRecordVecDeleteError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	stmts, err := prepareUpsertStmts(ctx, tx)
	if err != nil {
		tx.Rollback() //nolint:errcheck
		t.Fatal(err)
	}

	// Execute the chunk INSERT so it succeeds (pre-rollback).
	r := store.Record{
		Chunk:     chunker.Chunk{ID: "x", Source: "s", Text: "t"},
		Embedding: unitVec(0),
	}
	blob := encodeVec(r.Embedding)
	if _, err := stmts.chunk.ExecContext(ctx, r.Chunk.ID, r.Chunk.Source, r.Chunk.Text, blob); err != nil {
		tx.Rollback() //nolint:errcheck
		t.Fatal(err)
	}

	// Rollback so the vecDelete ExecContext fails next.
	tx.Rollback() //nolint:errcheck

	_, vecDeleteErr := stmts.vecDelete.ExecContext(ctx, r.Chunk.ID)
	stmts.close()
	if vecDeleteErr == nil {
		t.Error("expected error on vecDelete after rollback")
	}
}

// TestExecUpsertRecordVecInsertError covers the vecInsert error path (line 164-166).
func TestExecUpsertRecordVecInsertError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	stmts, err := prepareUpsertStmts(ctx, tx)
	if err != nil {
		tx.Rollback() //nolint:errcheck
		t.Fatal(err)
	}

	r := store.Record{
		Chunk:     chunker.Chunk{ID: "y", Source: "s", Text: "t"},
		Embedding: unitVec(0),
	}
	blob := encodeVec(r.Embedding)

	// chunk INSERT succeeds
	if _, err := stmts.chunk.ExecContext(ctx, r.Chunk.ID, r.Chunk.Source, r.Chunk.Text, blob); err != nil {
		tx.Rollback() //nolint:errcheck
		t.Fatal(err)
	}
	// vecDelete succeeds
	if _, err := stmts.vecDelete.ExecContext(ctx, r.Chunk.ID); err != nil {
		tx.Rollback() //nolint:errcheck
		t.Fatal(err)
	}

	// Rollback so vecInsert fails.
	tx.Rollback() //nolint:errcheck
	_, vecInsertErr := stmts.vecInsert.ExecContext(ctx, r.Chunk.ID, blob)
	stmts.close()
	if vecInsertErr == nil {
		t.Error("expected error on vecInsert after rollback")
	}
}

// TestExecUpsertRecordVecDeleteClosedStmt covers the vecDelete error path
// (line 161-163) by closing the vecDelete stmt before calling execUpsertRecord.
func TestExecUpsertRecordVecDeleteClosedStmt(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmts, err := prepareUpsertStmts(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	defer stmts.chunk.Close()
	defer stmts.vecInsert.Close()

	// Closing vecDelete makes its ExecContext fail.
	stmts.vecDelete.Close()

	err = execUpsertRecord(ctx, stmts, store.Record{
		Chunk:     chunker.Chunk{ID: "v1", Source: "s", Text: "t"},
		Embedding: unitVec(0),
	})
	if err == nil {
		t.Error("expected error when vecDelete stmt is closed")
	}
}

// TestExecUpsertRecordVecInsertClosedStmt covers the vecInsert error path
// (line 164-166) by closing the vecInsert stmt before calling execUpsertRecord.
func TestExecUpsertRecordVecInsertClosedStmt(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmts, err := prepareUpsertStmts(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	defer stmts.chunk.Close()
	defer stmts.vecDelete.Close()

	// Closing vecInsert makes its ExecContext fail.
	stmts.vecInsert.Close()

	err = execUpsertRecord(ctx, stmts, store.Record{
		Chunk:     chunker.Chunk{ID: "v2", Source: "s", Text: "t"},
		Embedding: unitVec(0),
	})
	if err == nil {
		t.Error("expected error when vecInsert stmt is closed")
	}
}

// TestRunMigrationsApplyError covers the applyMigration error path in
// runMigrations (line 80-82) by closing the DB after ensureMigrationsTable.
func TestRunMigrationsApplyError(t *testing.T) {
	db := openRawDB(t)
	s := &Store{db: db}
	if err := s.ensureMigrationsTable(); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if err := s.runMigrations(); err == nil {
		t.Error("expected error when applyMigration fails within runMigrations")
	}
}

// TestUpsertBeginTxError covers the BeginTx error in Upsert (line 185-187)
// by using a closed DB.
func TestUpsertBeginTxError(t *testing.T) {
	db := openRawDB(t)
	db.Close()
	s := &Store{db: db}
	err := s.Upsert(context.Background(), []store.Record{
		{Chunk: chunker.Chunk{ID: "x", Source: "s", Text: "t"}, Embedding: unitVec(0)},
	})
	if err == nil {
		t.Error("expected error with closed DB")
	}
}

// TestUpsertExecUpsertRecordError covers the execUpsertRecord error path in
// Upsert (line 191-193). We need the prepareUpsertStmts to succeed but the
// execUpsertRecord to fail. We use query_only to allow BEGIN but block INSERT.
func TestUpsertExecUpsertRecordError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	// Make the DB read-only so INSERT fails.
	if _, err := s.db.Exec("PRAGMA query_only = ON"); err != nil {
		t.Fatal(err)
	}

	err := s.Upsert(ctx, []store.Record{
		{Chunk: chunker.Chunk{ID: "z", Source: "s", Text: "t"}, Embedding: unitVec(0)},
	})
	if err == nil {
		t.Error("expected error with read-only DB")
	}
}

// TestDeleteBySourceBeginTxError covers the BeginTx error in DeleteBySource
// (line 209-211).
func TestDeleteBySourceBeginTxError(t *testing.T) {
	db := openRawDB(t)
	db.Close()
	s := &Store{db: db}
	if err := s.DeleteBySource(context.Background(), "src"); err == nil {
		t.Error("expected error with closed DB")
	}
}

// TestDeleteBySourceFirstExecError covers the first ExecContext error in
// DeleteBySource (line 214-216) by making the DB read-only.
func TestDeleteBySourceFirstExecError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	if _, err := s.db.Exec("PRAGMA query_only = ON"); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteBySource(ctx, "src"); err == nil {
		t.Error("expected error with read-only DB")
	}
}

// TestSearchRowsErrError covers rows.Err() in Search (line 248-250).
// A cancelled context mid-iteration can cause rows.Err() to be non-nil.
func TestSearchRowsErrError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	// Add records so the search has something to iterate.
	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "r1", Source: "s", Text: "t1"}, Embedding: unitVec(0)},
	}
	if err := s.Upsert(ctx, records); err != nil {
		t.Fatal(err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // cancel before search
	_, err := s.Search(cancelCtx, unitVec(0), 1)
	if err == nil {
		t.Error("expected error with cancelled context during Search")
	}
}

// fakeDirEntry satisfies fs.DirEntry for testing readMigrationsDir overrides.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.isDir }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// TestRunMigrationsReadDirError covers line 73-75 (readMigrationsDir error).
func TestRunMigrationsReadDirError(t *testing.T) {
	orig := readMigrationsDir
	readMigrationsDir = func() ([]fs.DirEntry, error) {
		return nil, fmt.Errorf("forced readdir error")
	}
	defer func() { readMigrationsDir = orig }()

	db := openRawDB(t)
	s := &Store{db: db}
	if err := s.ensureMigrationsTable(); err != nil {
		t.Fatal(err)
	}
	if err := s.runMigrations(); err == nil {
		t.Error("expected error when ReadDir fails")
	}
}

// TestRunMigrationsIsDirSkip covers lines 77-78 (IsDir continue branch).
func TestRunMigrationsIsDirSkip(t *testing.T) {
	orig := readMigrationsDir
	readMigrationsDir = func() ([]fs.DirEntry, error) {
		return []fs.DirEntry{fakeDirEntry{name: "subdir", isDir: true}}, nil
	}
	defer func() { readMigrationsDir = orig }()

	db := openRawDB(t)
	s := &Store{db: db}
	if err := s.ensureMigrationsTable(); err != nil {
		t.Fatal(err)
	}
	if err := s.runMigrations(); err != nil {
		t.Errorf("expected no error when all entries are dirs: %v", err)
	}
}

// TestApplyMigrationReadFileError covers line 99-101 (readMigrationFileFn error).
func TestApplyMigrationReadFileError(t *testing.T) {
	orig := readMigrationFileFn
	readMigrationFileFn = func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("forced read file error")
	}
	defer func() { readMigrationFileFn = orig }()

	db := openRawDB(t)
	s := &Store{db: db}
	if err := s.ensureMigrationsTable(); err != nil {
		t.Fatal(err)
	}
	if err := s.applyMigration("fake_version.sql"); err == nil {
		t.Error("expected error when ReadFile fails")
	}
}

// TestApplyMigrationExecError covers line 102-104 (Exec migration SQL error).
func TestApplyMigrationExecError(t *testing.T) {
	orig := readMigrationFileFn
	readMigrationFileFn = func(_ string) ([]byte, error) {
		return []byte("THIS IS NOT VALID SQL!!!"), nil
	}
	defer func() { readMigrationFileFn = orig }()

	db := openRawDB(t)
	s := &Store{db: db}
	if err := s.ensureMigrationsTable(); err != nil {
		t.Fatal(err)
	}
	if err := s.applyMigration("fake_version.sql"); err == nil {
		t.Error("expected error when migration SQL is invalid")
	}
}

// TestApplyMigrationRecordError covers line 107-109 (INSERT schema_migrations error).
func TestApplyMigrationRecordError(t *testing.T) {
	orig := readMigrationFileFn
	readMigrationFileFn = func(_ string) ([]byte, error) {
		return []byte("SELECT 1"), nil
	}
	defer func() { readMigrationFileFn = orig }()

	db := openRawDB(t)
	s := &Store{db: db}
	if err := s.ensureMigrationsTable(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TRIGGER fail_schema_migrations_insert
		BEFORE INSERT ON schema_migrations
		BEGIN
			SELECT RAISE(FAIL, 'forced failure');
		END
	`); err != nil {
		t.Fatal(err)
	}
	if err := s.applyMigration("fake_version.sql"); err == nil {
		t.Error("expected error when recording migration fails due to trigger")
	}
}

// TestOpenMigrateError covers lines 44-46 (migrate error in Open).
func TestOpenMigrateError(t *testing.T) {
	orig := migrateStoreFn
	migrateStoreFn = func(_ *Store) error { return fmt.Errorf("forced migrate error") }
	defer func() { migrateStoreFn = orig }()

	path := filepath.Join(t.TempDir(), "test.db")
	_, err := Open(path)
	if err == nil {
		t.Error("expected error when migrate fails")
	}
}

// TestPrepareVecDeleteFnError covers lines 138-141 (second PrepareContext error).
func TestPrepareVecDeleteFnError(t *testing.T) {
	orig := prepareVecDeleteFn
	prepareVecDeleteFn = func(_ context.Context, _ *sql.Tx) (*sql.Stmt, error) {
		return nil, fmt.Errorf("forced vec delete prepare error")
	}
	defer func() { prepareVecDeleteFn = orig }()

	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = prepareUpsertStmts(ctx, tx)
	if err == nil {
		t.Error("expected error when vecDelete prepare fails")
	}
}

// TestPrepareVecInsertFnError covers lines 144-148 (third PrepareContext error).
func TestPrepareVecInsertFnError(t *testing.T) {
	orig := prepareVecInsertFn
	prepareVecInsertFn = func(_ context.Context, _ *sql.Tx) (*sql.Stmt, error) {
		return nil, fmt.Errorf("forced vec insert prepare error")
	}
	defer func() { prepareVecInsertFn = orig }()

	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = prepareUpsertStmts(ctx, tx)
	if err == nil {
		t.Error("expected error when vecInsert prepare fails")
	}
}

// TestUpsertPrepareStmtsFnError covers lines 185-187 (prepareUpsertStmts error in Upsert).
func TestUpsertPrepareStmtsFnError(t *testing.T) {
	orig := prepareUpsertStmtsFn
	prepareUpsertStmtsFn = func(_ context.Context, _ *sql.Tx) (*upsertStmts, error) {
		return nil, fmt.Errorf("forced prepare error")
	}
	defer func() { prepareUpsertStmtsFn = orig }()

	s := openTestStore(t)
	err := s.Upsert(context.Background(), []store.Record{
		{Chunk: chunker.Chunk{ID: "x", Source: "s", Text: "t"}, Embedding: unitVec(0)},
	})
	if err == nil {
		t.Error("expected error when prepareUpsertStmts fails")
	}
}

// TestDeleteBySourceSecondExecError covers lines 214-216 (second ExecContext in DeleteBySource).
// A BEFORE DELETE trigger on the chunks table makes the second DELETE fail while the first succeeds.
func TestDeleteBySourceSecondExecError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	// Insert a chunk row so the second DELETE has rows to delete (trigger fires only on actual rows).
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO chunks (id, source, text, embedding) VALUES ('c1', 'src', 'text', '')`,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := s.db.ExecContext(ctx, `
		CREATE TRIGGER fail_chunks_delete
		BEFORE DELETE ON chunks
		BEGIN
			SELECT RAISE(FAIL, 'forced failure');
		END
	`); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteBySource(ctx, "src"); err == nil {
		t.Error("expected error when chunks DELETE fails due to trigger")
	}
}
