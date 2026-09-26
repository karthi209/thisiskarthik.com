package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"for-later-when-i-forget/internal/gamestats"
)

type gameArchive struct {
	Games []map[string]any `json:"games"`
}
type providerGame struct {
	Provider   string
	ExternalID int64
	Name       string
	Hours      float64
	LastPlayed int64
}

func importSteamLibrary(gamesPath, backlogPath string, cache gamestats.Cache, library []gamestats.SteamGame, staleDays int, now time.Time, dry bool) (gamestats.Cache, int, int, error) {
	providers := make([]providerGame, 0, len(library))
	for _, g := range library {
		providers = append(providers, providerGame{"steam", g.AppID, g.Name, float64(g.Minutes) / 60, g.LastPlayed})
	}
	return reconcileLibraries(gamesPath, backlogPath, cache, providers, staleDays, now, dry)
}

func reconcileLibraries(gamesPath, backlogPath string, old gamestats.Cache, providers []providerGame, staleDays int, now time.Time, dry bool) (gamestats.Cache, int, int, error) {
	var archive gameArchive
	var oldBacklog []map[string]any
	if err := readJSON(gamesPath, &archive); err != nil {
		return nil, 0, 0, err
	}
	if err := readJSON(backlogPath, &oldBacklog); err != nil {
		return nil, 0, 0, err
	}

	used, priorByTitle, priorByProvider := map[string]bool{}, map[string]map[string]any{}, map[string]map[string]any{}
	for _, records := range [][]map[string]any{archive.Games, oldBacklog} {
		for _, r := range records {
			used[stringValue(r["id"])] = true
			priorByTitle[normalizeTitle(stringValue(r["title"]))] = r
			if p, id := trackingKey(r); p != "" {
				priorByProvider[fmt.Sprintf("%s:%d", p, id)] = r
			}
		}
	}

	// Steam is authoritative for duplicate titles. Lutris contributes only local
	// games that Steam did not return for this account.
	selected := map[string]providerGame{}
	for _, g := range providers {
		key := normalizeTitle(g.Name)
		previous, ok := selected[key]
		steamWins := g.Provider == "steam" && previous.Provider != "steam"
		sameProviderNewer := g.Provider == previous.Provider && (g.LastPlayed > previous.LastPlayed || (g.LastPlayed == previous.LastPlayed && g.Hours > previous.Hours))
		if !ok || steamWins || sameProviderNewer {
			selected[key] = g
		}
	}
	ordered := make([]providerGame, 0, len(selected))
	for _, g := range selected {
		ordered = append(ordered, g)
	}
	sort.Slice(ordered, func(i, j int) bool { return strings.ToLower(ordered[i].Name) < strings.ToLower(ordered[j].Name) })

	completed, completedByTitle, completedByProvider := []map[string]any{}, map[string]map[string]any{}, map[string]map[string]any{}
	next := gamestats.Cache{}
	for _, r := range archive.Games {
		if stringValue(r["status"]) != "completed" {
			continue
		}
		completed = append(completed, r)
		completedByTitle[normalizeTitle(stringValue(r["title"]))] = r
		if p, id := trackingKey(r); p != "" {
			completedByProvider[fmt.Sprintf("%s:%d", p, id)] = r
		}
		if stat, ok := old[stringValue(r["id"])]; ok {
			next[stringValue(r["id"])] = stat
		}
	}

	active, backlog, cutoff := append([]map[string]any{}, completed...), []map[string]any{}, now.AddDate(0, 0, -staleDays)
	for _, g := range ordered {
		key := fmt.Sprintf("%s:%d", g.Provider, g.ExternalID)
		r := completedByProvider[key]
		if r == nil {
			r = completedByTitle[normalizeTitle(g.Name)]
		}
		if r != nil {
			next[stringValue(r["id"])] = providerStat(g)
			continue
		}
		r = priorByProvider[key]
		if r == nil {
			r = priorByTitle[normalizeTitle(g.Name)]
		}
		id := ""
		if r != nil {
			id = stringValue(r["id"])
		}
		if id == "" {
			id = uniqueGameID(slug(g.Name), g.ExternalID, used)
		}
		if r == nil {
			r = map[string]any{"id": id, "title": g.Name}
		}
		// Keep personal fields such as notes, labels and artwork when a provider
		// moves a game between the active archive and backlog.
		if stringValue(r["title"]) == "" {
			r["title"] = g.Name
		}
		if stringValue(r["platform"]) == "" {
			r["platform"] = "PC"
		}
		r["source"] = strings.Title(g.Provider)
		r["imported"] = true
		r["tracking"] = providerTracking(g)
		next[id] = providerStat(g)
		if belongsInProviderBacklog(g, cutoff) {
			delete(r, "status")
			backlog = append(backlog, r)
		} else {
			r["status"] = providerStatus(g, cutoff)
			active = append(active, r)
		}
	}
	archive.Games = active
	if dry {
		return next, len(active) - len(completed), len(backlog), nil
	}
	if err := writeJSONAtomic(gamesPath, archive); err != nil {
		return nil, 0, 0, err
	}
	if err := writeJSONAtomic(backlogPath, backlog); err != nil {
		return nil, 0, 0, err
	}
	return next, len(active) - len(completed), len(backlog), nil
}

func providerTracking(g providerGame) map[string]any {
	if g.Provider == "lutris" {
		return map[string]any{"provider": "lutris", "game_id": g.ExternalID}
	}
	return map[string]any{"provider": "steam", "app_id": g.ExternalID}
}
func providerStat(g providerGame) gamestats.Stat {
	s := gamestats.Stat{Provider: g.Provider, Hours: gamestats.Hours(g.Hours)}
	if g.LastPlayed > 0 {
		s.LastPlayed = time.Unix(g.LastPlayed, 0).UTC().Format(time.RFC3339)
	}
	return s
}
func belongsInProviderBacklog(g providerGame, cutoff time.Time) bool {
	return g.Hours == 0 || (g.LastPlayed > 0 && time.Unix(g.LastPlayed, 0).Before(cutoff))
}
func providerStatus(g providerGame, cutoff time.Time) string {
	if g.LastPlayed > 0 && !time.Unix(g.LastPlayed, 0).Before(cutoff) {
		return "playing"
	}
	return "played"
}
func trackingKey(r map[string]any) (string, int64) {
	t, _ := r["tracking"].(map[string]any)
	p := stringValue(t["provider"])
	if p == "lutris" {
		return p, int64Value(t["game_id"])
	}
	return p, int64Value(t["app_id"])
}

func readJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
func writeJSONAtomic(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".games-import-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(append(raw, '\n')); err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, path)
}
func normalizeTitle(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
func stringValue(v any) string       { s, _ := v.(string); return s }
func int64Value(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slug(v string) string {
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(v), "-"), "-")
}
func uniqueGameID(base string, id int64, used map[string]bool) string {
	if base != "" && !used[base] {
		used[base] = true
		return base
	}
	v := fmt.Sprintf("%s-%d", base, id)
	used[v] = true
	return v
}
