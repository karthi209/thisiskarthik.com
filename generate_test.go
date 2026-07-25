package main

import (
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
