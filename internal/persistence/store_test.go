package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeleteSaveRemovesFileAndIndex(t *testing.T) {
	dir := t.TempDir()
	store := NewLocalStore(dir)
	now := time.Unix(100, 0).UTC()
	save, err := store.CreateSave(context.Background(), SaveFile{
		SaveIdentity:   SaveIdentity{CompanyName: "Acme", CompanyType: "SaaS", CreatedAt: now, LastPlayedAt: now, Version: 1},
		SimulationSeed: 1,
		TickCounter:    0,
		Version:        1,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := store.DeleteSave(context.Background(), save.SaveIdentity.SaveID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "saves", save.SaveIdentity.SaveID+".json")); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, err=%v", err)
	}
	entries, err := store.LoadIndex(context.Background())
	if err != nil {
		t.Fatalf("load index: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty index, got %+v", entries)
	}
}

func TestDeleteSaveMissingFileStillCleansIndex(t *testing.T) {
	dir := t.TempDir()
	store := NewLocalStore(dir)
	now := time.Unix(100, 0).UTC()
	save, err := store.CreateSave(context.Background(), SaveFile{
		SaveIdentity:   SaveIdentity{CompanyName: "Acme", CompanyType: "SaaS", CreatedAt: now, LastPlayedAt: now, Version: 1},
		SimulationSeed: 1,
		TickCounter:    0,
		Version:        1,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_ = os.Remove(filepath.Join(dir, "saves", save.SaveIdentity.SaveID+".json"))
	if err := store.DeleteSave(context.Background(), save.SaveIdentity.SaveID); err != nil {
		t.Fatalf("delete missing file: %v", err)
	}
	entries, err := store.LoadIndex(context.Background())
	if err != nil {
		t.Fatalf("load index: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty index, got %+v", entries)
	}
}
