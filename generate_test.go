package main

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProcessJournalFileNeedsNoTitle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "2026-07-25.md")
	source := `---
title: Evening light
date: 2026-07-25
time: 18:30
mood: hopeful
tags: chennai, small-win
---

Noticed the evening light staying a little longer.`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}

	entry, err := processJournalFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if entry.DateISO != "2026-07-25" || entry.TimeLabel != "6:30 pm" {
		t.Fatalf("unexpected date labels: %s, %s", entry.DateISO, entry.TimeLabel)
	}
	if entry.Weekday != "saturday" || entry.WeekdayShort != "sat" {
		t.Fatalf("unexpected weekday labels: %s, %s", entry.Weekday, entry.WeekdayShort)
	}
	if entry.Mood != "hopeful" || len(entry.Tags) != 2 {
		t.Fatalf("optional metadata was not preserved: %#v", entry)
	}
	if entry.Title != "Evening light" {
		t.Fatalf("entry title was not preserved: %q", entry.Title)
	}
	if !strings.Contains(string(entry.Content), "evening light") {
		t.Fatalf("markdown body was not rendered: %s", entry.Content)
	}
}

func TestBuildTimelineContainsOnlyJournalNotes(t *testing.T) {
	noteDate := time.Date(2026, 7, 25, 18, 30, 0, 0, time.UTC)
	entries := []JournalEntry{{
		Title: "A small thing", Date: noteDate, DateISO: "2026-07-25", Day: "25", Month: "Jul", Year: 2026,
	}}

	years, entriesCount, daysCount, firstYear := buildTimeline(entries)
	if len(years) != 1 || len(years[0].Items) != 1 {
		t.Fatalf("unexpected timeline grouping: %#v", years)
	}
	if years[0].Items[0].Kind != "note" || years[0].Items[0].Title != "A small thing" {
		t.Fatalf("timeline note was not preserved: %#v", years[0].Items)
	}
	if entriesCount != 1 || daysCount != 1 || firstYear != 2026 {
		t.Fatalf("unexpected timeline stats: %d %d %d", entriesCount, daysCount, firstYear)
	}
}

func TestProcessAndGroupPhotos(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "desk.md")
	first := `---
title: My desk before leaving
date: 2026-07-24
image: /images/photos/desk.webp
image_alt: A desk with a notebook and keys
caption: Everything where I left it.
---`
	if err := os.WriteFile(firstPath, []byte(first), 0600); err != nil {
		t.Fatal(err)
	}
	secondPath := filepath.Join(dir, "rain.md")
	second := `---
title: Evening rain
date: 2026-06-19
image: /images/photos/rain.webp
image_alt: Rain falling outside a window
---`
	if err := os.WriteFile(secondPath, []byte(second), 0600); err != nil {
		t.Fatal(err)
	}

	firstPhoto, err := processPhotoFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	secondPhoto, err := processPhotoFile(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if firstPhoto.Slug != "my-desk-before-leaving" || firstPhoto.Caption == "" {
		t.Fatalf("photo fields were not preserved: %#v", firstPhoto)
	}

	groups := groupPhotos([]PhotoEntry{firstPhoto, secondPhoto})
	if len(groups) != 1 || len(groups[0].Months) != 2 {
		t.Fatalf("unexpected photo grouping: %#v", groups)
	}
	if groups[0].Months[0].Name != "July" || groups[0].Months[1].Name != "June" {
		t.Fatalf("photo months are not chronological: %#v", groups[0].Months)
	}
}

func TestProcessPhotoFileRequiresAltText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-alt.md")
	source := `---
title: Missing alt
date: 2026-07-24
image: /images/photos/missing.webp
---`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := processPhotoFile(path); err == nil {
		t.Fatal("expected a photo without alt text to be rejected")
	}
}

func TestProcessPhotoFileRejectsUnsafeSlug(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unsafe-slug.md")
	source := `---
title: Unsafe slug
date: 2026-07-24
slug: ../outside
image: /images/photos/unsafe.webp
image_alt: A test photograph
---`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := processPhotoFile(path); err == nil {
		t.Fatal("expected an unsafe photo slug to be rejected")
	}
}

func TestNewContentTemplatesRenderPopulatedEntries(t *testing.T) {
	templates, err := loadTemplates()
	if err != nil {
		t.Fatal(err)
	}

	photoData := PhotoPageData{
		PageType: "photos",
		BasePath: "/",
		Photo: PhotoEntry{
			Title:     "Evening rain",
			DateISO:   "2026-07-24",
			DateLabel: "24 Jul 2026",
			Image:     "/images/photos/rain.webp",
			ImageAlt:  "Rain outside a window",
			Caption:   "The rain stayed for an hour.",
		},
	}
	var photoOutput bytes.Buffer
	if err := templates.ExecuteTemplate(&photoOutput, "photo.html", photoData); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(photoOutput.String(), "Rain outside a window") || !strings.Contains(photoOutput.String(), "The rain stayed for an hour.") {
		t.Fatalf("photo detail template did not render its image and caption: %s", photoOutput.String())
	}
}

func TestGamesArchive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "games.json")
	source := `{"games":[
 {"id":"old","title":"Childhood game"},
 {"id":"year","title":"Year only","status":"paused","played":"2026"},
 {"id":"now","title":"Now <game>","status":"playing","started":"2026-09","label":"strange and lovely","resume_note":"Return to the red room.","artifact":"images/memory.webp","artifact_alt":"A red room","cover":"images/cover.webp"},
 {"id":"done","title":"Done","status":"completed","finished":"2026-08-08","rating":0,"favorite":true,"portrait":"images/portrait.jpg"}]}`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	stats := `{"now":{"provider":"steam","hours":0},"done":{"provider":"snapshot","hours":12.4},"not-curated":{"hours":999}}`
	if err := os.WriteFile(filepath.Join(dir, "game-stats.json"), []byte(stats), 0600); err != nil {
		t.Fatal(err)
	}
	originalBase := basePath
	basePath = "/test/"
	t.Cleanup(func() { basePath = originalBase })
	data, err := loadGames(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Playing) != 1 || len(data.Undated) != 1 || len(data.Years) != 1 || len(data.Years[0].Games) != 2 {
		t.Fatalf("incorrect lifecycle grouping: %#v", data)
	}
	if data.Years[0].Games[0].ID != "done" || data.Years[0].Games[1].DateLabel != "2026" {
		t.Fatal("chronology or date precision lost")
	}
	if data.Undated[0].Hours != nil || data.Undated[0].HoursLabel != "" || data.Playing[0].HoursLabel != "0h" {
		t.Fatal("unknown vs zero lost")
	}
	if data.Years[0].Summary != "2 games · 12.4 hours recorded" {
		t.Fatal(data.Years[0].Summary)
	}
	if len(data.Playtime) != 1 || data.Playtime[0].ID != "done" {
		t.Fatalf("playtime view includes zero-hour games: %#v", data.Playtime)
	}
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "games.html", data); err != nil {
		t.Fatal(err)
	}
	html := out.String()
	for _, want := range []string{"Currently playing", "Earlier", "Most played", "by playtime", "find something I forgot", "data-memory-game", "game-status-stamp", "game-leader", "game-personal-label", "strange and lovely", "where I left off", "Return to the red room.", `src="/test/images/memory.webp"`, `alt="A red room"`, "Now &lt;game&gt;", "Playtime", "12.4h", `aria-label="favourite"`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, bad := range []string{"backlog", "unknown", "N/A", "snapshot", "lifetime", "Steam", "999"} {
		if strings.Contains(html, bad) {
			t.Errorf("unwanted public text %q", bad)
		}
	}
	if strings.Count(html, `id="game-now"`) != 1 {
		t.Fatal("current game duplicated")
	}
	// Changing only personal status moves a game from now into history.
	source = strings.Replace(source, `"status":"playing"`, `"status":"completed"`, 1)
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	data, err = loadGames(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Playing) != 0 || len(data.Years[0].Games) != 3 {
		t.Fatal("completion did not move game into history")
	}
}

func TestGamesPlaytimeIncludesBacklog(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "games.json"), []byte(`{"games":[{"id":"archive","title":"Archive game"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "games-backlog.json"), []byte(`[{"id":"queued","title":"Queued game"},{"id":"unplayed","title":"Unplayed game"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "game-stats.json"), []byte(`{"archive":{"hours":2},"queued":{"hours":4.5},"unplayed":{"hours":0}}`), 0600); err != nil {
		t.Fatal(err)
	}

	data, err := loadGames(filepath.Join(dir, "games.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Playtime) != 2 || data.Playtime[0].ID != "queued" || data.Playtime[1].ID != "archive" {
		t.Fatalf("backlog playtime missing or incorrectly sorted: %#v", data.Playtime)
	}
}

func TestGamesOptionalStats(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "games.json")
	if err := os.WriteFile(path, []byte(`{"games":[{"id":"memory","title":"A memory"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"", `{"broken":`, `{"memory":{"hours":-4}}`} {
		if raw != "" {
			if err := os.WriteFile(filepath.Join(dir, "game-stats.json"), []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
		}
		data, err := loadGames(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(data.Undated) != 1 || data.Undated[0].Hours != nil || data.Undated[0].Status != "played" {
			t.Fatal("missing stats damaged personal record")
		}
	}
}

func TestGamesRejectInvalidRecords(t *testing.T) {
	for _, source := range []string{
		`[{"title":"A","status":"backlog"}]`,
		`[{"title":"A","rating":6}]`,
		`[{"title":"A","hours":4}]`,
		`[{"title":"A","started":"2026-09","finished":"2026-08"}]`,
		`[{"title":"A","finshed":"2026"}]`,
		`[{"title":"A","started":"2026-02-30"}]`,
		`[{"id":"same","title":"A"},{"id":"same","title":"B"}]`,
		`[{"title":"A","portrait":"https://example.com/a.jpg"}]`,
	} {
		path := filepath.Join(t.TempDir(), "games.json")
		if err := os.WriteFile(path, []byte(`{"games":`+source+`}`), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadGames(path); err == nil {
			t.Errorf("accepted invalid record %s", source)
		}
	}
}

func TestGamesStarterFacts(t *testing.T) {
	data, err := loadGames("content/games.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Playing) < 2 {
		t.Fatal("expected the manually recorded current games")
	}
	records := map[string]Game{}
	for _, g := range data.Playing {
		records[g.ID] = g
	}
	for _, g := range data.Undated {
		records[g.ID] = g
	}
	for _, year := range data.Years {
		for _, g := range year.Games {
			if g.Status == "playing" {
				t.Fatal("active game in history")
			}
			records[g.ID] = g
		}
	}
	if len(records) < 31 {
		t.Fatalf("personal history was lost; got %d records", len(records))
	}
	witcher := records["the-witcher-3"]
	if witcher.Finished != "2026" || witcher.Status != "completed" {
		t.Fatal("Witcher personal record changed")
	}
	if records["red-dead-redemption-2"].Finished != "2024" {
		t.Fatal("Red Dead Redemption 2 year changed")
	}
	if records["icy-tower"].ApproxYear != "early 2000s" || records["firewatch"].Finished != "2020" {
		t.Fatal("expanded game history changed")
	}
	if len(data.Backlog) == 0 {
		t.Fatal("provider-derived backlog is empty")
	}
	for _, g := range data.Backlog {
		records[g.ID] = g
	}
	for _, g := range records {
		if g.Status == "playing" && g.ArtifactURL != "" {
			t.Fatalf("%s currently playing has cover art", g.Title)
		}
		if g.ArtifactURL == "" && g.Portrait != "" && g.Status != "playing" {
			t.Fatalf("%s portrait was not prepared for rendering", g.Title)
		}
		for _, art := range []string{g.Cover, g.Portrait, g.Artifact} {
			if art != "" {
				if _, err := os.Stat(filepath.Join(contentDir, art)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestBooksStarterFacts(t *testing.T) {
	data, err := loadBooks("content/books.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Read) != 3 || data.Read[0].Title != "The Last Wish" {
		t.Fatalf("unexpected read books: %#v", data.Read)
	}
	if len(data.Reading) != 4 || data.Reading[0].Title != "The Left Hand of Darkness" {
		t.Fatalf("unexpected current books: %#v", data.Reading)
	}
	if data.Reading[0].Author != "Ursula K. Le Guin" || data.Reading[0].CoverURL != "/images/books/the-left-hand-of-darkness.webp" {
		t.Fatalf("book presentation metadata was not prepared: %#v", data.Reading[0])
	}
	if len(data.Planned) != 16 || data.Planned[0].Title != "Hyperion" {
		t.Fatalf("unexpected planned books: %#v", data.Planned)
	}
}

func TestGamesRememberedEras(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "games.json")
	source := `{"games":[{"id":"borrowed","title":"Borrowed game","approxYear":"early 2000s","sortYear":2001,"machine":"family","source":"friend's CD","hoursNote":"hundreds of hours, probably","memory":"Driving around."},{"id":"unknown","title":"Forgotten shooter","status":"unidentified","approxYear":"childhood"}],"milestones":[{"id":"family","title":"Family PC","date":"~2000","sortYear":2000,"hardware":"CRT; specs unknown","photo":"images/family.webp","photoAlt":"Childhood computer"}]}`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "game-stats.json"), []byte(`{"borrowed":{"provider":"snapshot","hours":150}}`), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := loadGames(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Years) != 2 || data.Years[0].Year != "early 2000s" || data.Years[1].Year != "~2000" || len(data.Undated) != 1 {
		t.Fatalf("uncertain chronology lost: %#v", data)
	}
	g := data.Years[0].Games[0]
	if g.Date != "" || g.Hours != nil || len(data.Playtime) != 0 || g.MachineLabel != "Family PC" {
		t.Fatalf("memory acquired false precision: %#v", g)
	}
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "games.html", data); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"hundreds of hours, probably", "Driving around.", "game-era-photo", "unidentified", "data-memory-game", "early 2000s"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(out.String(), "150h") {
		t.Fatal("fabricated childhood telemetry rendered")
	}
}

func TestCollectionPresentations(t *testing.T) {
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	books, err := loadBooks("content/books.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := templates.ExecuteTemplate(&out, "books.html", books); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"collection-dark", "collection-shelf", "shelf-entry", `alt="Cover of The Left Hand of Darkness"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("books missing %q", want)
		}
	}
	out.Reset()
	games, err := loadGames("content/games.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := templates.ExecuteTemplate(&out, "games.html", games); err != nil {
		t.Fatal(err)
	}
	html := out.String()
	for _, want := range []string{`data-game-view="shelf"`, `data-game-panel="shelf">`, `data-game-panel="years" hidden`, `data-game-view="shelf" aria-pressed="true"`, `class="games-page collection-dark"`, `id="shelf-road-rash"`, "title lost to history"} {
		if !strings.Contains(html, want) {
			t.Errorf("games missing %q", want)
		}
	}
	for _, g := range games.Playing {
		if strings.Contains(html, `id="shelf-`+g.ID+`"`) {
			t.Errorf("currently playing duplicated on shelf: %s", g.ID)
		}
	}
	out.Reset()
	fixture := PhotosPageData{PageType: "photos", BasePath: "/preview/", Years: groupPhotos([]PhotoEntry{{Title: "A quiet object", Slug: "quiet-object", Year: 2026, MonthNumber: 10, MonthLong: "October", Day: "04", DateISO: "2026-10-04", Image: "/preview/images/test.webp", ImageAlt: "An object on a plain surface"}})}
	if err := templates.ExecuteTemplate(&out, "photos.html", fixture); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`class="photo-contact-sheet"`, `href="/preview/photos/quiet-object"`, `alt="An object on a plain surface"`, `src="/preview/images/test.webp"`, `id="photo-year-2026"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("photos missing %q", want)
		}
	}
}
