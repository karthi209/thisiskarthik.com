package gamestats

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// Python's standard-library SQLite avoids a CGO dependency in the Go site.
//
//go:embed lutris_read.py
var lutrisReader string

type LutrisGame struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	Hours      *float64 `json:"playtime"`
	LastPlayed *int64   `json:"lastplayed"`
}

func ReadLutris(ctx context.Context, path string) ([]LutrisGame, error) {
	output, err := exec.CommandContext(ctx, "python3", "-c", lutrisReader, path).Output()
	if err != nil {
		return nil, fmt.Errorf("cannot read Lutris database %s (requires Python 3, readable SQLite, and games id/name/slug/playtime/lastplayed columns)", path)
	}
	var games []LutrisGame
	if err = json.Unmarshal(output, &games); err != nil {
		return nil, fmt.Errorf("invalid Lutris query result")
	}
	return games, nil
}

// DiscoverLutris validates candidates rather than choosing the first database.
func DiscoverLutris(ctx context.Context, home, xdg string) (string, error) {
	roots := []string{filepath.Join(home, ".var/app/net.lutris.Lutris/data/lutris"), filepath.Join(home, ".local/share/lutris")}
	if xdg != "" {
		roots = append(roots, filepath.Join(xdg, "lutris"))
	}
	seen := map[string]bool{}
	var valid []string
	for _, root := range roots {
		paths, _ := filepath.Glob(filepath.Join(root, "*.db"))
		for _, path := range paths {
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil || seen[resolved] {
				continue
			}
			seen[resolved] = true
			if _, err = ReadLutris(ctx, resolved); err == nil {
				valid = append(valid, resolved)
			}
		}
	}
	sort.Strings(valid)
	if len(valid) == 1 {
		return valid[0], nil
	}
	if len(valid) > 1 {
		return "", fmt.Errorf("multiple Lutris databases found; select one with --lutris-db")
	}
	return "", fmt.Errorf("no compatible Lutris database found; set --lutris-db or LUTRIS_DB (Python 3 required)")
}

func LutrisPath(ctx context.Context, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return DiscoverLutris(ctx, home, os.Getenv("XDG_DATA_HOME"))
}
