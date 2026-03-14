<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [adr, store, embedder, db]
description:  "ADR-012: Use sqlite-vec for vector search instead of a dedicated vector database"
-->
# ADR-012: Use sqlite-vec for Vector Search Instead of a Dedicated Vector Database

- **Status:** Accepted
- **Date:** 2026-03-14
- **Deciders:** Project maintainers
- **Scope:** Vector storage and similarity search for the ingestion pipeline and quiz retrieval

---

## Context

Hatch needs to store and query embedding vectors to support semantic search over indexed
source material. The quiz engine retrieves the top-K most relevant chunks for a given topic
using cosine/KNN similarity (Milestone 2 onward).

Dedicated vector databases (ChromaDB, Pinecone, Weaviate, Qdrant) exist to solve exactly
this problem, but they are designed for server-side or cloud deployments and impose runtime
dependencies that conflict with hatch's core design constraint: **distribute as a single
self-contained binary**.

---

## Decision

Use **sqlite-vec** (`vec0` virtual table) for KNN vector search, co-located in the same
SQLite database file used for all other hatch persistence.

---

## Alternatives Considered

### ChromaDB

- Python-based — incompatible with a pure Go binary
- Requires a running HTTP server (`chroma run`) or in-process Python embedding via `chromadb`
  package
- Adds Python, pip, and a background process as hard runtime dependencies
- Data lives in a separate store from hatch's relational tables, requiring cross-store joins

### Pinecone / Weaviate / Qdrant

- Cloud or self-hosted server processes — all require network configuration
- Overkill for a local CLI tool with one user and tens of thousands of chunks at most
- Introduce API keys, authentication, and availability dependencies for an offline-first tool

### pgvector (PostgreSQL extension)

- Requires PostgreSQL — already ruled out by ADR-001 (no database server)

### In-process brute-force cosine (initial M1 implementation)

- Implemented in M1 as a placeholder: scan all rows, compute cosine similarity in Go
- O(n) per query — acceptable for small stores but degrades linearly as the corpus grows
- Replaced by sqlite-vec KNN in M2

---

## Consequences

### Positive

- No additional runtime dependencies — sqlite-vec ships as part of the Go binary via CGO
- Vector data and relational data (chunks, sessions, users) live in the same file and
  transaction boundary
- KNN query is a single SQL statement; results join directly to the `chunks` table
- Offline-first: works without network access or any running service
- Trivial test isolation — in-memory SQLite per test, same as all other store tests

### Trade-offs

- Requires CGO: `mattn/go-sqlite3` replaced `modernc.org/sqlite` (pure Go) in M2 to gain
  access to the sqlite-vec extension. Cross-compilation is slightly more complex.
- sqlite-vec is an extension maintained outside the SQLite core — must track upstream for
  bug fixes and new SQLite version compatibility
- Not suitable if hatch ever needs to serve thousands of concurrent vector queries; at that
  scale a dedicated ANN index (HNSW, IVF) would outperform sqlite-vec's brute-force KNN

---

## Related ADRs

- ADR-001: Use SQLite + sqlite-vec for embedded persistence
- ADR-006: Local developer experience (no external services required)
