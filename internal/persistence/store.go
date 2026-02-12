package persistence

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const currentVersion = 1

type Store interface {
	LoadIndex(context.Context) ([]SaveIndexEntry, error)
	CreateSave(context.Context, SaveFile) (SaveFile, error)
	LoadSave(context.Context, string) (SaveFile, error)
	TouchSave(context.Context, string, time.Time) error
}

type LocalStore struct {
	baseDir string
}

func NewLocalStore(baseDir string) *LocalStore {
	return &LocalStore{baseDir: baseDir}
}

func DefaultBaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "terminal-business"), nil
	case "windows":
		appData := os.Getenv("AppData")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "terminal-business"), nil
	default:
		xdg := os.Getenv("XDG_DATA_HOME")
		if xdg == "" {
			xdg = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(xdg, "terminal-business"), nil
	}
}

func (s *LocalStore) LoadIndex(_ context.Context) ([]SaveIndexEntry, error) {
	if err := s.ensureDirs(); err != nil {
		return nil, err
	}
	idx, err := s.readIndex()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []SaveIndexEntry{}, nil
		}
		rebuilt, rebuildErr := s.rebuildIndex()
		if rebuildErr != nil {
			return nil, fmt.Errorf("read index: %w", err)
		}
		return rebuilt.Entries, nil
	}
	sortEntries(idx.Entries)
	return idx.Entries, nil
}

func (s *LocalStore) CreateSave(_ context.Context, save SaveFile) (SaveFile, error) {
	if err := s.ensureDirs(); err != nil {
		return SaveFile{}, err
	}
	if save.SaveIdentity.SaveID == "" {
		save.SaveIdentity.SaveID = newID()
	}
	if save.SaveIdentity.CompanyID == "" {
		save.SaveIdentity.CompanyID = newID()
	}
	if save.Version == 0 {
		save.Version = currentVersion
	}
	if save.SaveIdentity.Version == 0 {
		save.SaveIdentity.Version = currentVersion
	}
	if save.SaveIdentity.CreatedAt.IsZero() {
		save.SaveIdentity.CreatedAt = time.Now().UTC()
	}
	if save.SaveIdentity.LastPlayedAt.IsZero() {
		save.SaveIdentity.LastPlayedAt = save.SaveIdentity.CreatedAt
	}
	path := s.savePath(save.SaveIdentity.SaveID)
	if err := writeAtomicJSON(path, save); err != nil {
		return SaveFile{}, err
	}
	if err := s.upsertIndex(save, path); err != nil {
		return SaveFile{}, err
	}
	return save, nil
}

func (s *LocalStore) LoadSave(_ context.Context, saveID string) (SaveFile, error) {
	var save SaveFile
	path := s.savePath(saveID)
	data, err := os.ReadFile(path)
	if err != nil {
		return SaveFile{}, err
	}
	if err := json.Unmarshal(data, &save); err != nil {
		return SaveFile{}, err
	}
	return save, nil
}

func (s *LocalStore) TouchSave(ctx context.Context, saveID string, t time.Time) error {
	save, err := s.LoadSave(ctx, saveID)
	if err != nil {
		return err
	}
	save.SaveIdentity.LastPlayedAt = t.UTC()
	path := s.savePath(saveID)
	if err := writeAtomicJSON(path, save); err != nil {
		return err
	}
	return s.upsertIndex(save, path)
}

func (s *LocalStore) upsertIndex(save SaveFile, path string) error {
	idx, err := s.readIndex()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		idx = SaveIndex{Version: currentVersion, Entries: []SaveIndexEntry{}}
	}
	if idx.Version == 0 {
		idx.Version = currentVersion
	}
	entry := SaveIndexEntry{
		SaveID:       save.SaveIdentity.SaveID,
		CompanyName:  save.SaveIdentity.CompanyName,
		CompanyType:  save.SaveIdentity.CompanyType,
		LastPlayedAt: save.SaveIdentity.LastPlayedAt,
		SaveFilePath: path,
		Version:      currentVersion,
	}
	found := false
	for i := range idx.Entries {
		if idx.Entries[i].SaveID == entry.SaveID {
			idx.Entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		idx.Entries = append(idx.Entries, entry)
	}
	sortEntries(idx.Entries)
	return writeAtomicJSON(s.indexPath(), idx)
}

func (s *LocalStore) readIndex() (SaveIndex, error) {
	data, err := os.ReadFile(s.indexPath())
	if err != nil {
		return SaveIndex{}, err
	}
	var idx SaveIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return SaveIndex{}, err
	}
	if idx.Entries == nil {
		idx.Entries = []SaveIndexEntry{}
	}
	return idx, nil
}

func (s *LocalStore) rebuildIndex() (SaveIndex, error) {
	dir := s.savesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			idx := SaveIndex{Version: currentVersion, Entries: []SaveIndexEntry{}}
			_ = writeAtomicJSON(s.indexPath(), idx)
			return idx, nil
		}
		return SaveIndex{}, err
	}
	idx := SaveIndex{Version: currentVersion, Entries: []SaveIndexEntry{}}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var save SaveFile
		if err := json.Unmarshal(data, &save); err != nil {
			continue
		}
		idx.Entries = append(idx.Entries, SaveIndexEntry{
			SaveID:       save.SaveIdentity.SaveID,
			CompanyName:  save.SaveIdentity.CompanyName,
			CompanyType:  save.SaveIdentity.CompanyType,
			LastPlayedAt: save.SaveIdentity.LastPlayedAt,
			SaveFilePath: path,
			Version:      currentVersion,
		})
	}
	sortEntries(idx.Entries)
	if err := writeAtomicJSON(s.indexPath(), idx); err != nil {
		return SaveIndex{}, err
	}
	return idx, nil
}

func sortEntries(entries []SaveIndexEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].LastPlayedAt.Equal(entries[j].LastPlayedAt) {
			return entries[i].SaveID < entries[j].SaveID
		}
		return entries[i].LastPlayedAt.After(entries[j].LastPlayedAt)
	})
}

func (s *LocalStore) ensureDirs() error {
	for _, p := range []string{s.savesDir(), s.indexDir(), s.logsDir()} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalStore) savePath(saveID string) string {
	return filepath.Join(s.savesDir(), saveID+".json")
}
func (s *LocalStore) savesDir() string  { return filepath.Join(s.baseDir, "saves") }
func (s *LocalStore) indexDir() string  { return filepath.Join(s.baseDir, "index") }
func (s *LocalStore) logsDir() string   { return filepath.Join(s.baseDir, "logs") }
func (s *LocalStore) indexPath() string { return filepath.Join(s.indexDir(), "saves-index.json") }

func writeAtomicJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
