package main

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestWeeklyHistoryGroupsContributionsByMonday(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE contributions (
			id INTEGER PRIMARY KEY,
			amount_cents INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`); err != nil {
		t.Fatal(err)
	}
	for _, timestamp := range []string{
		"2026-09-14T10:00:00Z",
		"2026-09-20T23:59:59Z",
		"2026-09-21T00:00:00Z",
	} {
		if _, err := db.Exec(
			"INSERT INTO contributions (amount_cents, created_at) VALUES (?, ?)",
			contributionCents,
			timestamp,
		); err != nil {
			t.Fatal(err)
		}
	}

	history, err := (&application{db: db}).weeklyHistory(time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	if len(history.Weeks) != 8 {
		t.Fatalf("expected 8 weeks, got %d", len(history.Weeks))
	}
	if got := history.Weeks[6].Count; got != 2 {
		t.Fatalf("expected 2 contributions for Sep 14, got %d", got)
	}
	if got := history.Weeks[7].Count; got != 1 {
		t.Fatalf("expected 1 contribution for Sep 21, got %d", got)
	}
	if got := history.Weeks[6].Height; got != 100 {
		t.Fatalf("expected Sep 14 bar to be 100%%, got %d%%", got)
	}
	if got := history.Weeks[7].Height; got != 50 {
		t.Fatalf("expected Sep 21 bar to be 50%%, got %d%%", got)
	}
}
