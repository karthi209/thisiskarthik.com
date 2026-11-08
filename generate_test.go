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

func TestPhotoYearOnlyDoesNotInventMonthOrDay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "year-only.md")
	source := "---\ntitle: The clock tower\ndate: 2023\nimage: /images/photos/2023/clock.webp\nimage_alt: A clock tower\n---"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	photo, err := processPhotoFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if photo.Year != 2023 || photo.DateISO != "2023" || photo.DateLabel != "2023" || photo.Day != "" || photo.MonthNumber != 0 || photo.MonthLong != "" {
		t.Fatalf("year-only photo gained invented precision: %#v", photo)
	}
	groups := groupPhotos([]PhotoEntry{photo})
	if len(groups) != 1 || len(groups[0].Months) != 1 || groups[0].Months[0].Name != "" {
		t.Fatalf("year-only photo gained a month heading: %#v", groups)
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
	if data.Years[0].Games[0].ArtifactURL != "" {
		t.Fatal("cover must not become a second artifact image")
	}
	if data.Playing[0].ArtifactURL != "/test/images/memory.webp" {
		t.Fatal("explicit saved artifact was lost")
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
	for _, want := range []string{"playing now", "Earlier", "most played", "by year", "by playtime", `data-game-panel="years"`, "game-chronology-grid", "game-card", "game-personal-label", "strange and lovely", "where I left off", "Return to the red room.", `src="/test/images/memory.webp"`, `alt="A red room"`, "Now &lt;game&gt;", "Playtime", "12.4h", `aria-label="favourite"`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, bad := range []string{"unknown", "N/A", "snapshot", "lifetime", "Steam", "999"} {
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
	if records["icy-tower"].Finished != "2006" || records["icy-tower"].ApproxYear != "" || records["firewatch"].Finished != "2020" {
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
		if g.Portrait != "" && g.PortraitURL == "" {
			t.Fatalf("%s portrait was not prepared for rendering", g.Title)
		}
		if g.Artifact == "" && g.ArtifactURL != "" {
			t.Fatalf("%s cover was duplicated as an artifact", g.Title)
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
	for _, want := range []string{"hundreds of hours, probably", "Driving around.", "unidentified", "early 2000s"} {
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
	for _, want := range []string{`class="games-page books-page collection-page collection-page--books"`, "collection-shelf", "shelf-entry", `alt="Cover of The Left Hand of Darkness"`} {
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
	for _, want := range []string{`data-game-view="years" aria-pressed="true"`, `data-game-view="playtime"`, `data-game-panel="years">`, `data-game-panel="playtime" hidden`, `class="games-page collection-page collection-page--games"`, `class="collection-shelf game-chronology-grid"`, `id="game-road-rash"`, "year by year"} {
		if !strings.Contains(html, want) {
			t.Errorf("games missing %q", want)
		}
	}
	if strings.Contains(html, `data-game-view="shelf"`) || strings.Contains(html, `data-random-game`) {
		t.Error("removed game controls still rendered")
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

func TestHalfStarRatings(t *testing.T) {
	for _, value := range []float64{0, 0.5, 2.5, 4.5, 5} {
		if !validRating(&value) {
			t.Errorf("rejected valid rating %v", value)
		}
	}
	for _, value := range []float64{-0.5, 2.3, 5.5} {
		if validRating(&value) {
			t.Errorf("accepted invalid rating %v", value)
		}
	}
	value := 2.5
	stars := ratingStars(&value)
	want := []string{"full", "full", "half", "empty", "empty"}
	for i := range want {
		if stars[i] != want[i] {
			t.Errorf("star %d = %s; want %s", i, stars[i], want[i])
		}
	}
	if !validRating(nil) {
		t.Fatal("unrated entries should be valid")
	}
}

func TestOriginalReleaseAndPersonalHistoryStaySeparate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "games.json")
	source := `{"games":[
 {"id":"older-game","title":"Older game","status":"completed","finished":"2026","original_release":"2001-12-22","original_release_source":"https://example.com/release"},
 {"id":"newer-game","title":"Newer game","status":"completed","finished":"2024","original_release":"2020"},
 {"id":"remembered","title":"Remembered","status":"completed","approxYear":"early 2000s","sortYear":2001,"original_release":"1994"}]}`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := loadGames(path)
	if err != nil {
		t.Fatal(err)
	}
	if data.Years[0].Year != "2026" || data.Years[0].Games[0].ID != "older-game" {
		t.Fatal("release dates changed the personal archive order")
	}
	game := data.Years[0].Games[0]
	if game.OriginalReleaseLabel() != "2001" || game.PersonalDateLabel() != "2026" || game.PersonalDateHeading() != "Completed" {
		t.Fatalf("date semantics lost: %#v", game)
	}
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "games.html", data); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<dt>Released</dt>`, `<time datetime="2001">2001</time>`, `<dt>Completed</dt><dd>2026</dd>`, `<dt>Completed (approx.)</dt><dd>early 2000s</dd>`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing distinct date display %q", want)
		}
	}
	// Backlogged records receive the same catalogue metadata without invented play dates.
	backlog := `[{"id":"waiting","title":"Waiting","original_release":"2010-05"}]`
	if err := os.WriteFile(filepath.Join(dir, "games-backlog.json"), []byte(backlog), 0600); err != nil {
		t.Fatal(err)
	}
	data, err = loadGames(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Backlog) != 1 || data.Backlog[0].OriginalReleaseLabel() != "2010" || data.Backlog[0].PersonalDateLabel() != "Not dated" {
		t.Fatal("backlog dates lost or invented")
	}
}

func TestOriginalReleaseValidation(t *testing.T) {
	for _, date := range []string{"", "1994", "2002-09", "2020-07-17"} {
		if err := validateGameRelease(Game{OriginalRelease: date}); err != nil {
			t.Errorf("valid precision %q: %v", date, err)
		}
	}
	for _, game := range []Game{
		{OriginalRelease: "2020-02-30"},
		{OriginalRelease: "circa 2000"},
		{OriginalRelease: "2020", OriginalReleaseSource: "javascript:alert(1)"},
		{OriginalRelease: "2020", OriginalReleaseSource: "https:///missing-host"},
		{OriginalReleaseSource: "https://example.com/release"},
	} {
		if err := validateGameRelease(game); err == nil {
			t.Errorf("accepted invalid release: %#v", game)
		}
	}
}

func TestGameStatsCountUniqueRecordsAndHalfStars(t *testing.T) {
	hours, rating, zero := 12.5, 4.5, 0.0
	done := Game{ID: "done", Status: "completed", Hours: &hours, Rating: &rating, OriginalRelease: "1990"}
	playing := Game{ID: "now", Status: "playing", Rating: &zero}
	data := GamesPageData{
		Playing:  []Game{playing},
		Years:    []GameYear{{Year: "2026", Games: []Game{done}}, {Year: "2007", Games: []Game{{ID: "old", Status: "played", HoursNote: "many hours"}}}},
		Backlog:  []Game{{ID: "queued", Status: "backlog"}},
		Playtime: []Game{done},
	}
	stats := summarizeGames(data)
	if stats.Played != 2 || stats.Completed != 1 || stats.Playing != 1 || stats.Backlog != 1 || stats.Rated != 2 || stats.HoursLabel != "12.5h" {
		t.Fatalf("incorrect statistics: %#v", stats)
	}
	if stats.Ratings[9].Count != 1 || stats.Ratings[0].Count != 1 || stats.Ratings[10].Count != 0 {
		t.Fatal("half stars or zero rating lost")
	}
	if stats.Years[0].Label != "2007" || stats.Years[1].Label != "2026" {
		t.Fatal("chart used release dates or lost personal years")
	}
	empty := summarizeGames(GamesPageData{})
	if empty.Rated != 0 || empty.HoursLabel != "0.0h" {
		t.Fatal("empty archive invented statistics")
	}
	for _, bar := range empty.Ratings {
		if bar.Count != 0 || bar.Percent != 0 {
			t.Fatal("empty rating histogram invented bars")
		}
	}
}

func TestGameArchiveRoadRashUpdate(t *testing.T) {
	data, err := loadGames("content/games.json")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, year := range data.Years {
		for _, g := range year.Games {
			if g.ID == "brainsbreaker" {
				t.Fatal("removed game remains in archive")
			}
			if g.ID == "road-rash" {
				found = true
				if year.Year != "2007" || g.Played != "2007" || g.PortraitURL == "" {
					t.Fatal("Road Rash year or cover missing")
				}
			}
		}
	}
	if !found {
		t.Fatal("Road Rash missing from dated archive")
	}
}

func TestBookStatsUseOnlyRecordedReadingData(t *testing.T) {
	four, half, zero, pages := 4.0, 4.5, 0.0, 250
	data := BooksPageData{
		Read:    []Book{{ID: "first", Author: "An author", Status: "read", Rating: &four, Finished: "2024", Pages: &pages}, {ID: "second", Author: "An author", Status: "read", Rating: &half, Finished: "2026-02"}, {ID: "undated", Author: "Another author", Status: "read"}},
		Reading: []Book{{ID: "current", Author: "Third author", Status: "reading"}},
		Planned: []Book{{ID: "waiting", Author: "An author", Status: "planned", Rating: &zero, Pages: &pages, Finished: "2027"}},
	}
	stats := summarizeBooks(data)
	if stats.Read != 3 || stats.Reading != 1 || stats.Planned != 1 || stats.Authors != 3 || stats.Rated != 3 || stats.AverageLabel != "2.83" {
		t.Fatalf("incorrect book summary: %#v", stats)
	}
	if stats.PagesRead != 250 || stats.BooksWithPages != 1 {
		t.Fatal("unfinished books or unknown pages counted as read")
	}
	if len(stats.Years) != 2 || stats.Years[0].Label != "2024" || stats.Years[1].Label != "2026" {
		t.Fatal("undated or unfinished books gained a reading year")
	}
	if stats.Ratings[9].Count != 1 || stats.Ratings[0].Count != 1 {
		t.Fatal("half-star or zero rating missing")
	}
	if stats.AuthorCounts[0].Label != "An author" || stats.AuthorCounts[0].Count != 3 {
		t.Fatal("author chart doesn't reflect all shelves")
	}
	empty := summarizeBooks(BooksPageData{})
	if empty.AverageLabel != "—" || empty.Rated != 0 || len(empty.Years) != 0 || empty.BooksWithPages != 0 {
		t.Fatal("empty reading data invented statistics")
	}
}

func TestBooksRejectInvalidOptionalReadingData(t *testing.T) {
	for _, fields := range []string{`"pages":0`, `"pages":-5`, `"finished":"2026-02-30"`, `"started":"2026-05","finished":"2026-04"`, `"rating":3.7`} {
		path := filepath.Join(t.TempDir(), "books.json")
		if err := os.WriteFile(path, []byte(`{"books":[{"id":"test","title":"Test","author":"An author","status":"read",`+fields+`}]}`), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadBooks(path); err == nil {
			t.Errorf("accepted invalid book fields: %s", fields)
		}
	}
}

func TestBooksOverviewPreservesUndatedHistory(t *testing.T) {
	data, err := loadBooks("content/books.json")
	if err != nil {
		t.Fatal(err)
	}
	if data.Stats.Read != 3 || data.Stats.Reading != 4 || data.Stats.Planned != 16 || data.Stats.Rated != 0 || len(data.Stats.Years) != 0 {
		t.Fatalf("incorrect recorded book statistics: %#v", data.Stats)
	}
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "books.html", data); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"My books", "Authors on my shelves", "My ratings", "No ratings yet.", "Find on Goodreads", `aria-label="Reading shelves"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing book overview feature: %s", want)
		}
	}
	for _, bad := range []string{"Read by year", "Pages recorded", "on the nightstand", "the reading pile"} {
		if strings.Contains(out.String(), bad) {
			t.Errorf("unrecorded or decorative book text: %s", bad)
		}
	}
}
