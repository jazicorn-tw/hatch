package fake_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jazicorn/hatch/internal/source"
	"github.com/jazicorn/hatch/internal/source/fake"
)

func TestFetchReturnsDocs(t *testing.T) {
	docs := []source.Document{
		{ID: "doc1", Source: "test", Content: "hello"},
	}
	f := &fake.Fetcher{Docs: docs}

	got, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(got) != 1 || got[0].ID != "doc1" {
		t.Errorf("unexpected docs: %v", got)
	}
}

func TestFetchReturnsError(t *testing.T) {
	f := &fake.Fetcher{Err: errors.New("fetch failed")}

	_, err := f.Fetch(context.Background())
	if err == nil {
		t.Error("expected error from Fetch")
	}
}

func TestFetchEmpty(t *testing.T) {
	f := &fake.Fetcher{}

	got, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty docs, got %d", len(got))
	}
}
