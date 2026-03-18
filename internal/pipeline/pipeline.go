// Package pipeline orchestrates the ingestion of documents into the vector store.
package pipeline

import (
	"context"
	"fmt"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/embedder"
	"github.com/jazicorn/hatch/internal/source"
	"github.com/jazicorn/hatch/internal/store"
)

// Progress is sent on the progress channel after each batch is upserted.
type Progress struct {
	// Done is the number of chunks processed so far.
	Done int
	// Total is the total number of chunks to process.
	Total int
	// Fields carries optional display labels (e.g. current source name).
	Fields map[string]string
}

type docChunks struct {
	doc    source.Document
	chunks []chunker.Chunk
}

// Run executes the ingestion pipeline end-to-end:
//  1. Fetch all documents from src.
//  2. Chunk each document with chk.
//  3. Embed all chunks in one call (the Embedder batches internally).
//  4. Upsert records per-document batch into st.
//  5. Send Progress updates on progressCh after each upsert.
//
// progressCh may be nil — progress updates are silently skipped.
// The pipeline does not call Close on any of its dependencies.
func Run(
	ctx context.Context,
	src source.Fetcher,
	chk chunker.Chunker,
	emb embedder.Embedder,
	st store.VecStore,
	progressCh chan<- Progress,
) error {
	docs, err := src.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("pipeline: fetch: %w", err)
	}

	all, total, err := chunkAll(docs, chk)
	if err != nil {
		return err
	}

	return embedAndUpsert(ctx, all, total, emb, st, progressCh)
}

// chunkAll chunks all documents and returns the results with total chunk count.
func chunkAll(docs []source.Document, chk chunker.Chunker) ([]docChunks, int, error) {
	var all []docChunks
	for _, doc := range docs {
		chunks, err := chk.Chunk(doc)
		if err != nil {
			return nil, 0, fmt.Errorf("pipeline: chunk %s: %w", doc.ID, err)
		}
		if len(chunks) == 0 {
			continue
		}
		all = append(all, docChunks{doc: doc, chunks: chunks})
	}
	total := 0
	for _, dc := range all {
		total += len(dc.chunks)
	}
	return all, total, nil
}

// embedAndUpsert embeds and upserts each document's chunks, sending progress updates.
func embedAndUpsert(ctx context.Context, all []docChunks, total int, emb embedder.Embedder, st store.VecStore, progressCh chan<- Progress) error {
	done := 0
	for _, dc := range all {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("pipeline: cancelled: %w", err)
		}

		texts := make([]string, len(dc.chunks))
		for i, c := range dc.chunks {
			texts[i] = c.Text
		}

		vecs, err := emb.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("pipeline: embed %s: %w", dc.doc.ID, err)
		}
		if len(vecs) != len(dc.chunks) {
			return fmt.Errorf("pipeline: embed %s: got %d vectors for %d chunks",
				dc.doc.ID, len(vecs), len(dc.chunks))
		}

		records := make([]store.Record, len(dc.chunks))
		for i, c := range dc.chunks {
			records[i] = store.Record{Chunk: c, Embedding: vecs[i]}
		}

		if err := st.Upsert(ctx, records); err != nil {
			return fmt.Errorf("pipeline: upsert %s: %w", dc.doc.ID, err)
		}

		done += len(records)
		if progressCh != nil {
			progressCh <- Progress{
				Done:  done,
				Total: total,
				Fields: map[string]string{
					"source": dc.doc.Source,
					"doc":    dc.doc.ID,
				},
			}
		}
	}
	return nil
}
