package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"for-later-when-i-forget/internal/gamestats"
)

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	flags := flag.NewFlagSet("games-sync", flag.ContinueOnError)
	human := flags.String("games", "content/games.json", "human archive (read only)")
	backlog := flags.String("backlog", "content/games-backlog.json", "backlog archive")
	stats := flags.String("stats", "content/game-stats.json", "generated stats output")
	provider := flags.String("provider", "all", "all, steam or lutris")
	lutris := flags.String("lutris-db", os.Getenv("LUTRIS_DB"), "optional Lutris SQLite path")
	dry := flags.Bool("dry-run", false, "report changes without writing")
	list := flags.Bool("list-lutris", false, "list local Lutris identifiers without syncing")
	importSteam := flags.Bool("import-steam-library", false, "merge every visible Steam game into the archives")
	staleDays := flags.Int("stale-days", 30, "provider games older than this go to backlog")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if *list {
		path, err := gamestats.LutrisPath(ctx, *lutris)
		if err != nil {
			return failure(err)
		}
		rows, err := gamestats.ReadLutris(ctx, path)
		if err != nil {
			return failure(err)
		}
		fmt.Println(path)
		for _, g := range rows {
			fmt.Printf("%d\t%s\t%s\n", g.ID, g.Slug, g.Name)
		}
		return 0
	}
	if *importSteam {
		if *staleDays < 1 {
			return failure(fmt.Errorf("--stale-days must be at least 1"))
		}
		cache, err := gamestats.Read(*stats)
		if err != nil {
			return failure(fmt.Errorf("refusing to overwrite unreadable stats: %w", err))
		}
		client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
		library, err := gamestats.FetchSteamLibrary(ctx, client, gamestats.SteamEndpoint, os.Getenv("STEAM_API_KEY"), os.Getenv("STEAM_ID"))
		if err != nil {
			return failure(err)
		}
		providers := make([]providerGame, 0, len(library))
		for _, game := range library {
			providers = append(providers, providerGame{Provider: "steam", ExternalID: game.AppID, Name: game.Name, Hours: float64(game.Minutes) / 60, LastPlayed: game.LastPlayed})
		}
		lutrisCount := 0
		if path, pathErr := gamestats.LutrisPath(ctx, *lutris); pathErr == nil {
			if rows, readErr := gamestats.ReadLutris(ctx, path); readErr == nil {
				for _, game := range rows {
					hours := 0.0
					if game.Hours != nil {
						hours = *game.Hours
					}
					last := int64(0)
					if game.LastPlayed != nil {
						last = *game.LastPlayed
					}
					providers = append(providers, providerGame{Provider: "lutris", ExternalID: game.ID, Name: game.Name, Hours: hours, LastPlayed: last})
					lutrisCount++
				}
			}
		}
		next, played, queued, err := reconcileLibraries(*human, *backlog, cache, providers, *staleDays, time.Now(), *dry)
		if err != nil {
			return failure(err)
		}
		if !*dry {
			if err := gamestats.WriteAtomic(*stats, next); err != nil {
				return failure(err)
			}
		}
		if *dry {
			fmt.Println("Dry run; no files changed.")
		}
		fmt.Printf("Providers: %d Steam, %d Lutris; %d active, %d backlog.\n", len(library), lutrisCount, played, queued)
		return 0
	}
	if *provider != "all" && *provider != "steam" && *provider != "lutris" {
		return failure(fmt.Errorf("--provider must be all, steam or lutris"))
	}
	// Prevent accidental overwrite of the human archive, including symlinks/hardlinks.
	a, _ := filepath.Abs(*human)
	b, _ := filepath.Abs(*stats)
	ai, ae := os.Stat(a)
	bi, be := os.Stat(b)
	if a == b || (ae == nil && be == nil && os.SameFile(ai, bi)) {
		return failure(fmt.Errorf("stats output must differ from the human archive"))
	}
	records, err := gamestats.ReadRecords(*human)
	if err != nil {
		return failure(err)
	}
	cache, err := gamestats.Read(*stats)
	if err != nil {
		return failure(fmt.Errorf("refusing to overwrite unreadable stats: %w", err))
	}
	fmt.Println("Game stats sync")
	next, updated, unchanged, failures := gamestats.Sync(ctx, records, cache, gamestats.Options{Provider: *provider, SteamKey: os.Getenv("STEAM_API_KEY"), SteamID: os.Getenv("STEAM_ID"), LutrisDB: *lutris}, os.Stdout)
	if !*dry && updated > 0 {
		if err = gamestats.WriteAtomic(*stats, next); err != nil {
			return failure(err)
		}
	}
	if *dry {
		fmt.Print("\nDry run: ")
	}
	fmt.Printf("%d updated, %d unchanged.\n", updated, unchanged)
	if failures > 0 {
		return 1
	}
	return 0
}
func failure(err error) int { fmt.Fprintln(os.Stderr, err); return 1 }
