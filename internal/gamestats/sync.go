package gamestats

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

type Options struct {
	Provider, SteamKey, SteamID, LutrisDB string
	Client                                *http.Client
	SteamURL                              string
}

// Sync merges only matched observations. Every unmatched or failed record survives unchanged.
func Sync(ctx context.Context, records []Record, existing Cache, opts Options, out io.Writer) (Cache, int, int, int) {
	next := Cache{}
	for id, s := range existing {
		next[id] = s
	}
	updated, unchanged, failures := 0, 0, 0
	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	}
	if opts.SteamURL == "" {
		opts.SteamURL = SteamEndpoint
	}
	for _, provider := range []string{"steam", "lutris"} {
		if opts.Provider != "all" && opts.Provider != provider {
			continue
		}
		var selected []Record
		var ids []int64
		for _, g := range records {
			if g.Tracking != nil && g.Tracking.Provider == provider {
				id := g.Tracking.AppID
				if provider == "lutris" {
					id = g.Tracking.GameID
				}
				if id == 0 {
					fmt.Fprintf(out, "warning: %s has no %s tracking identifier\n", g.Title, provider)
					continue
				}
				selected = append(selected, g)
				ids = append(ids, id)
			}
		}
		if len(selected) == 0 {
			continue
		}
		fmt.Fprintf(out, "\n%s\n", provider)
		var fetched map[int64]Stat
		var err error
		if provider == "steam" {
			fetched, err = FetchSteam(ctx, opts.Client, opts.SteamURL, opts.SteamKey, opts.SteamID, ids)
		} else {
			var path string
			path, err = LutrisPath(ctx, opts.LutrisDB)
			if err == nil {
				var rows []LutrisGame
				rows, err = ReadLutris(ctx, path)
				fetched = map[int64]Stat{}
				for _, row := range rows {
					if row.Hours == nil || *row.Hours < 0 {
						continue
					}
					s := Stat{Provider: "lutris", Hours: Hours(*row.Hours)}
					if row.LastPlayed != nil {
						s.LastPlayed = timestamp(*row.LastPlayed)
					}
					fetched[row.ID] = s
				}
			}
		}
		if err != nil {
			fmt.Fprintf(out, "warning: %v; previous %s stats kept\n", err, provider)
			failures++
			continue
		}
		for _, g := range selected {
			id := g.Tracking.AppID
			if provider == "lutris" {
				id = g.Tracking.GameID
			}
			s, ok := fetched[id]
			if !ok {
				fmt.Fprintf(out, "warning: %s has no matching playtime; previous stats kept\n", g.Title)
				continue
			}
			old := next[g.ID]
			if s.LastPlayed == "" && old.Provider == s.Provider {
				s.LastPlayed = old.LastPlayed
			}
			if reflect.DeepEqual(old, s) {
				unchanged++
			} else {
				updated++
			}
			fmt.Fprintf(out, "  %s  %s → %s\n", g.Title, formatHours(old.Hours), formatHours(s.Hours))
			next[g.ID] = s
		}
	}
	return next, updated, unchanged, failures
}
func formatHours(h *float64) string {
	if h == nil {
		return "—"
	}
	return fmt.Sprintf("%.2fh", *h)
}
