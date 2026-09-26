package gamestats

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func clientFor(body string, code int) *http.Client {
	return &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})}
}

func TestSteamMappingPrecisionAndFailures(t *testing.T) {
	ctx := context.Background()
	client := clientFor(`{"response":{"games":[{"appid":292030,"playtime_forever":1123,"rtime_last_played":1700000000},{"appid":10,"playtime_forever":0},{"appid":11},{"appid":12,"playtime_forever":-1}]}}`, 200)
	rows, err := FetchSteam(ctx, client, SteamEndpoint, "secret", "123", []int64{292030, 10})
	if err != nil {
		t.Fatal(err)
	}
	if *rows[292030].Hours != 18.72 || rows[292030].LastPlayed != "2023-11-14T22:13:20Z" || *rows[10].Hours != 0 || len(rows) != 2 {
		t.Fatalf("bad conversion: %#v", rows)
	}
	for _, body := range []string{`{"response":{}}`, `{"response":{"game_count":0}}`, `broken`} {
		if _, err := FetchSteam(ctx, clientFor(body, 200), SteamEndpoint, "secret", "123", nil); err == nil {
			t.Fatal("accepted missing/private/invalid response")
		}
	}
	if _, err := FetchSteam(ctx, clientFor("", 403), SteamEndpoint, "secret", "123", nil); err == nil {
		t.Fatal("accepted HTTP failure")
	}
	if _, err := FetchSteam(ctx, client, SteamEndpoint, "", "", nil); err == nil {
		t.Fatal("accepted missing credentials")
	}
	bad := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { return nil, io.ErrUnexpectedEOF })}
	if _, err := FetchSteam(ctx, bad, SteamEndpoint, "super-secret", "123", nil); err == nil || strings.Contains(err.Error(), "super-secret") {
		t.Fatal("transport error leaks key or disappeared")
	}
}

func TestFetchSteamLibraryIncludesPlayedAndUnplayed(t *testing.T) {
	body := `{"response":{"games":[{"appid":10,"name":"Played","playtime_forever":90,"rtime_last_played":1700000000},{"appid":20,"name":"Unplayed","playtime_forever":0}]}}`
	games, err := FetchSteamLibrary(context.Background(), clientFor(body, 200), SteamEndpoint, "secret", "123")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 2 || games[0].Name != "Played" || games[0].Minutes != 90 || games[1].Minutes != 0 {
		t.Fatalf("full library was not preserved: %#v", games)
	}
}

func TestMergePreservesUnmatchedAndFailedProviders(t *testing.T) {
	old := Cache{"known": {Provider: "steam", Hours: Hours(4), LastPlayed: "2023-11-14T22:13:20Z"}, "missing": {Provider: "steam", Hours: Hours(7)}, "local": {Provider: "lutris", Hours: Hours(3)}, "historical": {Provider: "snapshot", Hours: Hours(8)}}
	records := []Record{{ID: "known", Title: "Same title", Tracking: &Tracking{Provider: "steam", AppID: 1}}, {ID: "missing", Title: "Same title", Tracking: &Tracking{Provider: "steam", AppID: 2}}, {ID: "local", Title: "Local", Tracking: &Tracking{Provider: "lutris", GameID: 1}}}
	var output bytes.Buffer
	opts := Options{Provider: "all", SteamKey: "key", SteamID: "123", LutrisDB: filepath.Join(t.TempDir(), "absent.db"), Client: clientFor(`{"response":{"games":[{"appid":1,"playtime_forever":1123},{"appid":999,"playtime_forever":99999}]}}`, 200)}
	next, updated, _, failures := Sync(context.Background(), records, old, opts, &output)
	if updated != 1 || failures != 1 || len(next) != 4 || *next["known"].Hours != 18.72 || next["known"].LastPlayed != old["known"].LastPlayed {
		t.Fatalf("bad merge: %v %s", next, output.String())
	}
	if *old["known"].Hours != 4 || !reflect.DeepEqual(next["local"], old["local"]) || !reflect.DeepEqual(next["missing"], old["missing"]) {
		t.Fatal("failed/unmatched data changed")
	}
	opts.Client = clientFor("failure", 503)
	next, updated, _, failures = Sync(context.Background(), records, old, opts, io.Discard)
	if updated != 0 || failures != 2 || !reflect.DeepEqual(old, next) {
		t.Fatal("provider failure destroyed cache")
	}
}

func makeLutris(t *testing.T, path string) {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Fatal("Python 3 is required for the Lutris tests")
	}
	script := `import sqlite3,sys
with sqlite3.connect(sys.argv[1]) as db:
 db.execute('create table games (id integer primary key, name text, slug text, playtime real, lastplayed integer)')
 db.executemany('insert into games values (?,?,?,?,?)',[(2,'Ghost','ghost',8.5649755,1700000000),(3,'Old','old',None,None),(4,'Zero','zero',0,0)])`
	if output, err := exec.Command("python3", "-c", script, path).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
}

func TestLutrisReadOnlyAndDiscovery(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".var/app/net.lutris.Lutris/data/lutris")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "custom.db")
	makeLutris(t, path)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	discovered, err := DiscoverLutris(context.Background(), home, "")
	if err != nil || discovered != path {
		t.Fatalf("discovery: %q %v", discovered, err)
	}
	rows, err := ReadLutris(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0].ID != 2 || rows[1].Hours != nil || *rows[2].Hours != 0 {
		t.Fatalf("invalid rows: %#v", rows)
	}
	records := []Record{{ID: "ghost", Title: "Ghost", Tracking: &Tracking{Provider: "lutris", GameID: 2}}}
	next, n, _, f := Sync(context.Background(), records, Cache{}, Options{Provider: "lutris", LutrisDB: path}, io.Discard)
	if n != 1 || f != 0 || *next["ghost"].Hours != 8.56 {
		t.Fatal("Lutris conversion failed")
	}
	after, _ := os.ReadFile(path)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("Lutris source was modified")
	}
	second := filepath.Join(dir, "other.db")
	makeLutris(t, second)
	if _, err := DiscoverLutris(context.Background(), home, ""); err == nil {
		t.Fatal("ambiguous databases accepted")
	}
	if _, err := ReadLutris(context.Background(), filepath.Join(dir, "missing.db")); err == nil {
		t.Fatal("missing database accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, "missing.db")); !os.IsNotExist(err) {
		t.Fatal("reader created missing database")
	}
	invalid := filepath.Join(dir, "invalid.db")
	if err := os.WriteFile(invalid, []byte("not a database"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadLutris(context.Background(), invalid); err == nil {
		t.Fatal("invalid schema accepted")
	}
}

func TestCacheAtomicRoundTripAndHumanValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stats.json")
	cache := Cache{"game": {Provider: "steam", Hours: Hours(12.345)}}
	if err := WriteAtomic(path, cache); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil || !reflect.DeepEqual(got, cache) {
		t.Fatal("cache roundtrip failed", err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, ".game-stats-*.tmp"))
	if len(files) != 0 {
		t.Fatal("temp cache left behind")
	}
	if err := os.WriteFile(path, []byte(`{"game":{"hours":-1}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path); err == nil {
		t.Fatal("negative stats accepted")
	}
	human := filepath.Join(dir, "games.json")
	for _, source := range []string{`{"games":[{"title":"Missing id"}]}`, `{"games":[{"id":"a","title":"A","tracking":{"provider":"bad"}}]}`, `{"games":[{"id":"a","title":"A","tracking":{"provider":"steam","app_id":1}},{"id":"b","title":"B","tracking":{"provider":"steam","app_id":1}}]}`} {
		if err := os.WriteFile(human, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadRecords(human); err == nil {
			t.Fatal("invalid human mapping accepted")
		}
	}
}
