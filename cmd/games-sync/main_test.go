package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"for-later-when-i-forget/internal/gamestats"
)

func TestDryRunAndHumanOwnership(t *testing.T) {
	dir := t.TempDir()
	human := filepath.Join(dir, "games.json")
	stats := filepath.Join(dir, "game-stats.json")
	db := filepath.Join(dir, "lutris.db")
	source := []byte(`{"games":[{"id":"personal","title":"My name for it","status":"paused","notes":"Irreplaceable memory","tracking":{"provider":"lutris","game_id":2}}]}`)
	old := []byte(`{"personal":{"provider":"lutris","hours":1}}`)
	if err := os.WriteFile(human, source, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stats, old, 0600); err != nil {
		t.Fatal(err)
	}
	script := `import sqlite3,sys
with sqlite3.connect(sys.argv[1]) as c:
 c.execute('create table games (id integer,name text,slug text,playtime real,lastplayed integer)')
 c.execute("insert into games values (2,'Provider title','slug',2.125,1700000000)")`
	if output, err := exec.Command("python3", "-c", script, db).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
	args := []string{"--games", human, "--stats", stats, "--provider", "lutris", "--lutris-db", db}
	if code := run(append(args, "--dry-run")); code != 0 {
		t.Fatalf("dry-run: %d", code)
	}
	after, _ := os.ReadFile(stats)
	if !bytes.Equal(after, old) {
		t.Fatal("dry run wrote cache")
	}
	if code := run(args); code != 0 {
		t.Fatalf("sync: %d", code)
	}
	after, _ = os.ReadFile(human)
	if !bytes.Equal(after, source) {
		t.Fatal("human archive changed")
	}
	after, _ = os.ReadFile(stats)
	if !bytes.Contains(after, []byte("2.13")) {
		t.Fatal("stats not updated")
	}
	if code := run([]string{"--games", human, "--stats", human}); code == 0 {
		t.Fatal("human output alias accepted")
	}
	alias := filepath.Join(dir, "alias.json")
	if err := os.Symlink(human, alias); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"--games", human, "--stats", alias}); code == 0 {
		t.Fatal("human symlink alias accepted")
	}
	if err := os.WriteFile(stats, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if code := run(args); code == 0 {
		t.Fatal("corrupt cache overwritten")
	}
	after, _ = os.ReadFile(stats)
	if string(after) != "broken" {
		t.Fatal("corrupt cache lost")
	}
}

func TestImportSteamLibraryClassifiesOnlyNewGames(t *testing.T) {
	dir := t.TempDir()
	gamesPath := filepath.Join(dir, "games.json")
	backlogPath := filepath.Join(dir, "games-backlog.json")
	if err := os.WriteFile(gamesPath, []byte(`{"games":[{"id":"manual","title":"My Custom Title","status":"completed","finished":"2020","tracking":{"provider":"steam","app_id":1}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backlogPath, []byte(`[{"id":"waiting","title":"Waiting Game","label":"rainy day game","resume_note":"At the old bridge."}]`), 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	library := []gamestats.SteamGame{
		{AppID: 1, Name: "Provider Renamed It", Minutes: 120, LastPlayed: now.Unix()},
		{AppID: 2, Name: "Recent Game", Minutes: 60, LastPlayed: now.AddDate(0, 0, -5).Unix()},
		{AppID: 3, Name: "Old Game", Minutes: 30, LastPlayed: now.AddDate(-2, 0, 0).Unix()},
		{AppID: 4, Name: "Never Played", Minutes: 0},
		{AppID: 5, Name: "Waiting Game", Minutes: 45, LastPlayed: now.Unix()},
	}
	cache, played, queued, err := importSteamLibrary(gamesPath, backlogPath, gamestats.Cache{}, library, 365, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if played != 2 || queued != 2 {
		t.Fatalf("unexpected classification: %d played, %d backlog", played, queued)
	}
	var archive gameArchive
	if err := readJSON(gamesPath, &archive); err != nil {
		t.Fatal(err)
	}
	if len(archive.Games) != 3 || stringValue(archive.Games[0]["status"]) != "completed" || stringValue(archive.Games[1]["status"]) != "playing" {
		t.Fatalf("manual record changed or recent game missing: %#v", archive.Games)
	}
	if stringValue(archive.Games[2]["label"]) != "rainy day game" || stringValue(archive.Games[2]["resume_note"]) != "At the old bridge." {
		t.Fatal("personal annotations were lost when backlog game became active")
	}
	var backlog []map[string]any
	if err := readJSON(backlogPath, &backlog); err != nil {
		t.Fatal(err)
	}
	if len(backlog) != 2 || cache["old-game"].Hours == nil || cache["waiting"].Hours == nil {
		t.Fatalf("backlog hours or entries missing: %#v %#v", backlog, cache)
	}
	matching, _ := json.Marshal(archive.Games)
	if !bytes.Contains(matching, []byte(`"app_id":5`)) {
		t.Fatal("existing backlog title did not gain its Steam mapping")
	}
	if _, _, _, err := importSteamLibrary(gamesPath, backlogPath, cache, library, 365, now.AddDate(2, 0, 0), false); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(gamesPath, &archive); err != nil {
		t.Fatal(err)
	}
	if len(archive.Games) != 1 || stringValue(archive.Games[0]["id"]) != "manual" {
		t.Fatalf("stale imported game did not move while manual record stayed put: %#v", archive.Games)
	}
}

func TestSteamWinsDuplicateLutrisTitle(t *testing.T) {
	dir := t.TempDir()
	gamesPath := filepath.Join(dir, "games.json")
	backlogPath := filepath.Join(dir, "games-backlog.json")
	if err := os.WriteFile(gamesPath, []byte(`{"games":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backlogPath, []byte(`[]`), 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	providers := []providerGame{
		{Provider: "lutris", ExternalID: 3, Name: "Shadow of the Tomb Raider", Hours: 100, LastPlayed: now.Unix()},
		{Provider: "steam", ExternalID: 750920, Name: "Shadow of the Tomb Raider", Hours: 1, LastPlayed: now.AddDate(0, 0, -2).Unix()},
	}
	cache, _, _, err := reconcileLibraries(gamesPath, backlogPath, gamestats.Cache{}, providers, 30, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if cache["shadow-of-the-tomb-raider"].Provider != "steam" {
		t.Fatalf("Steam did not win duplicate title: %#v", cache)
	}
}
