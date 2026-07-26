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

func TestProcessLibraryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "the-odyssey.md")
	source := `---
title: The Odyssey
release_year: 2026
date: 2026-07-24
rating: 4
---

Worth seeing on the biggest screen possible.`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}

	entry, err := processLibraryFile(path, "films")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Title != "The Odyssey" || entry.ReleaseYear != 2026 || entry.Section != "films" {
		t.Fatalf("library identity was not preserved: %#v", entry)
	}
	if entry.Rating != 4 || entry.RatingStars != "★★★★☆" || entry.RatingLabel != "4 out of 5" {
		t.Fatalf("unexpected rating: %#v", entry)
	}
	if entry.DateISO != "2026-07-24" || entry.Month != "Jul" || entry.Day != "24" {
		t.Fatalf("unexpected library date labels: %#v", entry)
	}
	if !strings.Contains(string(entry.Content), "biggest screen") {
		t.Fatalf("library note was not rendered: %s", entry.Content)
	}
}

func TestProcessLibraryFileRejectsInvalidRating(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.md")
	source := `---
title: Too Many Stars
date: 2026-07-24
rating: 6
---

A note.`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := processLibraryFile(path, "films"); err == nil {
		t.Fatal("expected an invalid rating to be rejected")
	}
}

func TestProcessLibraryFileAcceptsYearOnlyDateAndHalfRating(t *testing.T) {
	path := filepath.Join(t.TempDir(), "red-dead-redemption-2.md")
	source := `---
title: Red Dead Redemption 2
release_year: 2018
date: 2025
rating: 4.5
---

One of the greatest of all time.`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}

	entry, err := processLibraryFile(path, "games")
	if err != nil {
		t.Fatal(err)
	}
	if entry.DateISO != "2025" || entry.DateLabel != "2025" || entry.HasFullDate {
		t.Fatalf("year-only date was not preserved honestly: %#v", entry)
	}
	if entry.Rating != 4.5 || entry.RatingStars != "★★★★½" || entry.RatingLabel != "4.5 out of 5" {
		t.Fatalf("half-star rating was not preserved: %#v", entry)
	}
}

func TestLoadLibraryEntriesSortsNewestFirstAndSkipsDrafts(t *testing.T) {
	root := t.TempDir()
	filmsDir := filepath.Join(root, "films")
	if err := os.MkdirAll(filmsDir, 0700); err != nil {
		t.Fatal(err)
	}
	entries := map[string]string{
		"older.md": `---
title: Older
date: 2026-07-01
rating: 3
---

Older note.`,
		"newer.md": `---
title: Newer
date: 2026-07-24
rating: 5
---

Newer note.`,
		"draft.md": `---
title: Draft
date: 2026-07-25
rating: 5
draft: true
---

Draft note.`,
	}
	for name, source := range entries {
		if err := os.WriteFile(filepath.Join(filmsDir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}

	loaded, err := loadLibraryEntries(root)
	if err != nil {
		t.Fatal(err)
	}
	films := loaded["films"]
	if len(films) != 2 || films[0].Title != "Newer" || films[1].Title != "Older" {
		t.Fatalf("library entries were not filtered and sorted: %#v", films)
	}
	if films[0].ArchiveNo != "002" || films[1].ArchiveNo != "001" {
		t.Fatalf("unexpected library archive numbers: %#v", films)
	}
}

func TestBuildLibraryIndexSections(t *testing.T) {
	entries := map[string][]LibraryEntry{
		"films": {
			{
				Title:       "The Odyssey",
				ReleaseYear: 2026,
				DateISO:     "2026-07-24",
				Day:         "24",
				Month:       "Jul",
				Year:        2026,
			},
			{Title: "Blood Diamond", Year: 2024},
		},
		"books": {
			{Title: "The Iliad", Year: 2025},
		},
	}

	sections := buildLibraryIndexSections(entries)
	if len(sections) != 5 {
		t.Fatalf("expected five library summaries, got %d", len(sections))
	}
	films := sections[0]
	if !films.HasEntries || films.CountLabel != "2 films remembered" || films.YearsLabel != "2024 — 2026" {
		t.Fatalf("unexpected films summary: %#v", films)
	}
	if films.Description == "" || films.ArchiveLabel != "films remembered" {
		t.Fatalf("films summary lost its description or archive label: %#v", films)
	}
	if sections[1].HasEntries {
		t.Fatalf("empty tv section was marked populated: %#v", sections[1])
	}
	books := sections[3]
	if books.CountLabel != "1 book" || books.YearsLabel != "2025" {
		t.Fatalf("unexpected singular books summary: %#v", books)
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

	date := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
	libraryData := LibraryPageData{
		PageType: "library",
		BasePath: "/",
		Section:  librarySections[0],
		Years: []LibraryYear{{
			Year: 2026,
			Entries: []LibraryEntry{{
				Title:        "The Odyssey",
				ReleaseYear:  2026,
				RatingStars:  "★★★★☆",
				RatingLabel:  "4 out of 5",
				Date:         date,
				DateISO:      "2026-07-24",
				DateLabel:    "24 Jul 2026",
				Day:          "24",
				Month:        "Jul",
				WeekdayShort: "fri",
				Year:         2026,
				HasFullDate:  true,
				Content:      template.HTML("<p>Worth seeing.</p>"),
				ArchiveNo:    "001",
			}},
		}},
	}
	var libraryOutput bytes.Buffer
	if err := templates.ExecuteTemplate(&libraryOutput, "library-section.html", libraryData); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(libraryOutput.String(), "The Odyssey") || !strings.Contains(libraryOutput.String(), "★★★★☆") {
		t.Fatalf("populated library template did not render its entry: %s", libraryOutput.String())
	}

	libraryIndexData := LibraryIndexPageData{
		PageType: "library",
		BasePath: "/",
		Sections: buildLibraryIndexSections(map[string][]LibraryEntry{
			"films": {
				{
					Title:       "The Odyssey",
					ReleaseYear: 2026,
					DateISO:     "2026-07-24",
					DateLabel:   "24 Jul 2026",
					Day:         "24",
					Month:       "Jul",
					Year:        2026,
				},
			},
		}),
	}
	var libraryIndexOutput bytes.Buffer
	if err := templates.ExecuteTemplate(&libraryIndexOutput, "library.html", libraryIndexData); err != nil {
		t.Fatal(err)
	}
	indexHTML := libraryIndexOutput.String()
	if !strings.Contains(indexHTML, "1 film remembered") ||
		!strings.Contains(indexHTML, `<a href="/library/films">films</a>`) ||
		strings.Contains(indexHTML, "browse films") ||
		strings.Contains(indexHTML, "The Odyssey") {
		t.Fatalf("library landing did not render its collection summary: %s", indexHTML)
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
