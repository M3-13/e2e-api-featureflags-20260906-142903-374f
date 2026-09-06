package store

import (
	"errors"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "feature", Enabled: true, Description: "desc", RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.Get("feature")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != f {
		t.Fatalf("got %+v, want %+v", got, f)
	}
}

func TestCreateDuplicate(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "feature"}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Create(f); !errors.Is(err, ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
}

func TestGetNotFound(t *testing.T) {
	s := NewStore()
	if _, err := s.Get("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestListEmpty(t *testing.T) {
	s := NewStore()
	if got := s.List(); len(got) != 0 {
		t.Fatalf("want empty list, got %d entries", len(got))
	}
}

func TestListPopulated(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a"})
	_ = s.Create(Flag{Key: "b"})
	if got := s.List(); len(got) != 2 {
		t.Fatalf("want 2 entries, got %d", len(got))
	}
}

func TestUpdateKeepsKey(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature", Enabled: false, RolloutPercent: 10})
	updated, err := s.Update("feature", Flag{Key: "ignored", Enabled: true, RolloutPercent: 20})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Key != "feature" {
		t.Fatalf("key must not change, got %q", updated.Key)
	}
	if !updated.Enabled || updated.RolloutPercent != 20 {
		t.Fatalf("unexpected update result %+v", updated)
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := NewStore()
	if _, err := s.Update("missing", Flag{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature"})
	if err := s.Delete("feature"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get("feature"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := NewStore()
	if err := s.Delete("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
