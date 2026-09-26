// Package gamestats reads objective observations separately from the personal archive.
package gamestats

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

type Tracking struct {
	Provider string `json:"provider"`
	AppID    int64  `json:"app_id,omitempty"`
	GameID   int64  `json:"game_id,omitempty"`
}

type Record struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Tracking *Tracking `json:"tracking,omitempty"`
}

type Stat struct {
	Provider   string   `json:"provider"`
	Hours      *float64 `json:"hours,omitempty"`
	LastPlayed string   `json:"last_played,omitempty"`
}

type Cache map[string]Stat

func ReadRecords(path string) ([]Record, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var archive struct {
		Games []Record `json:"games"`
	}
	if err = json.Unmarshal(raw, &archive); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	mappings := map[string]bool{}
	for _, g := range archive.Games {
		if g.ID == "" || g.Title == "" || seen[g.ID] {
			return nil, fmt.Errorf("every game needs a unique stable id and title")
		}
		seen[g.ID] = true
		if g.Tracking != nil {
			t := g.Tracking
			if t.Provider != "steam" && t.Provider != "lutris" {
				return nil, fmt.Errorf("%s: unsupported provider %q", g.ID, t.Provider)
			}
			if t.AppID < 0 || t.GameID < 0 || (t.Provider == "steam" && t.GameID != 0) || (t.Provider == "lutris" && t.AppID != 0) {
				return nil, fmt.Errorf("%s: invalid tracking identifier", g.ID)
			}
			key := fmt.Sprintf("%s:%d:%d", t.Provider, t.AppID, t.GameID)
			if t.AppID != 0 || t.GameID != 0 {
				if mappings[key] {
					return nil, fmt.Errorf("%s: duplicate provider mapping", g.ID)
				}
				mappings[key] = true
			}
		}
	}
	return archive.Games, nil
}

func Read(path string) (Cache, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Cache{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cache Cache
	if err = json.Unmarshal(raw, &cache); err != nil {
		return nil, err
	}
	if cache == nil {
		return nil, fmt.Errorf("stats must be a JSON object")
	}
	for id, s := range cache {
		if s.Hours != nil && (*s.Hours < 0 || math.IsNaN(*s.Hours) || math.IsInf(*s.Hours, 0)) {
			return nil, fmt.Errorf("%s: invalid hours", id)
		}
		if s.LastPlayed != "" {
			if _, err = time.Parse(time.RFC3339, s.LastPlayed); err != nil {
				return nil, fmt.Errorf("%s: invalid last_played", id)
			}
		}
	}
	return cache, nil
}

// WriteAtomic only replaces the disposable cache after a complete write and flush.
func WriteAtomic(path string, cache Cache) error {
	raw, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".game-stats-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0644); err == nil {
		_, err = f.Write(append(raw, '\n'))
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, path)
}

func Hours(value float64) *float64 { value = math.Round(value*100) / 100; return &value }
func timestamp(unix int64) string {
	if unix <= 0 {
		return ""
	}
	return time.Unix(unix, 0).UTC().Format(time.RFC3339)
}
