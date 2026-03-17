package sqlite

import (
	"context"
	"testing"

	"github.com/jazicorn/hatch/internal/kata"
)

func TestSaveAndListKatas(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	k := &kata.Kata{
		ID:          "k1",
		Topic:       "go",
		Title:       "Hello World",
		Description: "Write a hello world program",
		StarterCode: "package main\n",
		Tests:       "func TestHello(t *testing.T) {}",
		Language:    kata.Go,
	}

	if err := s.SaveKata(ctx, k); err != nil {
		t.Fatalf("SaveKata: %v", err)
	}

	got, err := s.ListKatas(ctx, "go")
	if err != nil {
		t.Fatalf("ListKatas: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 kata, got %d", len(got))
	}
	if got[0].ID != "k1" {
		t.Errorf("expected id k1, got %s", got[0].ID)
	}
	if got[0].Title != "Hello World" {
		t.Errorf("expected title 'Hello World', got %s", got[0].Title)
	}
	if got[0].Topic != "go" {
		t.Errorf("expected topic go, got %s", got[0].Topic)
	}
	if got[0].Language != kata.Go {
		t.Errorf("expected language go, got %s", got[0].Language)
	}
	if got[0].StarterCode != k.StarterCode {
		t.Errorf("starter code mismatch: want %q, got %q", k.StarterCode, got[0].StarterCode)
	}
	if got[0].Tests != k.Tests {
		t.Errorf("tests mismatch: want %q, got %q", k.Tests, got[0].Tests)
	}
	if got[0].Description != k.Description {
		t.Errorf("description mismatch: want %q, got %q", k.Description, got[0].Description)
	}
}

func TestListKatasEmpty(t *testing.T) {
	s := openTestStore(t)

	got, err := s.ListKatas(context.Background(), "go")
	if err != nil {
		t.Fatalf("ListKatas on empty store: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 katas, got %d", len(got))
	}
}

func TestSaveKataReplaces(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	orig := &kata.Kata{
		ID:       "k1",
		Topic:    "go",
		Title:    "original title",
		Language: kata.Go,
	}
	if err := s.SaveKata(ctx, orig); err != nil {
		t.Fatalf("SaveKata original: %v", err)
	}

	updated := &kata.Kata{
		ID:       "k1",
		Topic:    "go",
		Title:    "updated title",
		Language: kata.Go,
	}
	if err := s.SaveKata(ctx, updated); err != nil {
		t.Fatalf("SaveKata updated: %v", err)
	}

	got, err := s.ListKatas(ctx, "go")
	if err != nil {
		t.Fatalf("ListKatas: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 kata, got %d", len(got))
	}
	if got[0].Title != "updated title" {
		t.Errorf("expected title 'updated title', got %s", got[0].Title)
	}
}

func TestSaveKataCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	k := &kata.Kata{ID: "x", Topic: "go", Language: kata.Go}
	err := s.SaveKata(ctx, k)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestListKatasCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.ListKatas(ctx, "go")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestListKatasFiltersByTopic(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	katas := []*kata.Kata{
		{ID: "k1", Topic: "go", Title: "Go kata", Language: kata.Go},
		{ID: "k2", Topic: "python", Title: "Python kata", Language: kata.Python},
	}
	for _, k := range katas {
		if err := s.SaveKata(ctx, k); err != nil {
			t.Fatalf("SaveKata %s: %v", k.ID, err)
		}
	}

	got, err := s.ListKatas(ctx, "go")
	if err != nil {
		t.Fatalf("ListKatas go: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 kata for topic go, got %d", len(got))
	}
	if got[0].ID != "k1" {
		t.Errorf("expected id k1, got %s", got[0].ID)
	}
}
