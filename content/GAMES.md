# Games: facts and memories

The routine is small:

1. Play normally; Steam and Lutris track playtime.
2. Run `go run ./cmd/games-sync` (or `make games-sync`) from the repository root.
3. Review `git diff -- content/game-stats.json`, run `make check`, commit/publish.
4. Edit `content/games.json` only when a status changes or there is something to remember.

Run `make games-artwork` when new provider games are imported. It downloads reviewed
Steam portrait art, resizes it to a small local WebP, records the local `portrait`
path, and updates `content/images/games/SOURCES.md`. Ambiguous or unavailable titles
are reported rather than guessed. Images are lazy-loaded only after a game opens.

Run `make games-sync` to reconcile the entire visible Steam library and local Lutris
database. Completed statuses and entered years remain authoritative. Steam is
authoritative when the same title exists in both providers; Lutris adds only local
titles absent from Steam. Games played within 30 days are `playing`; positive
playtime without a usable date is `played`; zero
playtime or a last-played date older than 30 days goes to `content/games-backlog.json`. Explicitly completed records
and their manually entered years never move. Hours remain available
in `content/game-stats.json`. Preview a different threshold without writing:

```bash
set -a; source .env; set +a
go run ./cmd/games-sync --import-steam-library --stale-days 30 --dry-run
```

## Two files, two owners

- `content/games.json`: human-owned titles, status, dates, personal labels,
  return notes, memories, artwork, and provider mappings. Sync preserves these
  fields when a provider moves a game between Playing and Backlog.
- `content/game-stats.json`: optional observations keyed by stable game ID:
  provider, hours, and last-played timestamp when available. Sync only writes this file.

Deleting stats never deletes a memory or changes status. Missing or malformed stats
are omitted from the site; malformed stats also produce a build warning. The sync
tool refuses to replace an unreadable cache. Initial supplied playtimes are kept
as `provider: "snapshot"` until a matching provider supplies fresh data. These
snapshots remain recoverable through version control; they are not live measurements.

```json
{
  "games": [
    {
      "id": "elden-ring",
      "title": "Elden Ring",
      "status": "playing",
      "platform": "PC",
      "started": "2026-09-23",
      "label": "beautiful mess",
      "resume_note": "Meet Takemura at the diner; relearn the quickhacks.",
      "artifact": "images/games/cyberpunk-cover.jpg",
      "artifact_alt": "Cyberpunk 2077 cover art",
      "tracking": {"provider": "steam", "app_id": 1245620}
    },
    {"id": "old-memory", "title": "A childhood game", "status": "played"}
  ]
}
```

Keep explicit stable IDs when adding tracked games. Title-only records can render
with a derived ID, but sync requires explicit unique IDs so renaming a game cannot
break the join. Valid statuses: `playing`, `completed`, `paused`, `dropped`, `played`
(default). Only personal status determines whether a game belongs in Currently
Playing or History. Completing a game moves it into History; no duplication.

Optional personal fields: `platform`, `source`, `started`, `finished`, `played`,
`label` (a personal phrase instead of a score), `resume_note` (where to continue),
`notes` (a lasting memory), `favorite` (boolean), `artifact` and `artifact_alt`,
`cover` (landscape), `portrait` (portrait), and `tracking`. Images use local
`images/games/...` paths backed by files in `content/images/games/`. An artifact,
or existing portrait when no artifact is set, appears only after opening a game.
The archive does not build a cover grid. The older numeric `rating` field remains
supported for existing records, but personal labels are preferred.
Provider/source is not shown repeatedly in public metadata.

Dates accept `YYYY`, `YYYY-MM`, or `YYYY-MM-DD`. History groups by finished date,
then remembered play date (`played`), then start date. Undated games stay in
“earlier”. Last-played observations do not change the personal chronology.
Missing dates/hours are omitted, not replaced with fake zeros. Known zero is valid.

Year summaries count historical games in that group and sum their available total
hours. “Hours recorded” describes those games' cumulative playtime, not an estimate
of activity within that calendar year. This is not a per-session annual tracker.

Detailed pages, screenshots, reviews, choices, and tags can later attach to these
IDs. New supported fields should be added deliberately; the generator rejects
misspellings rather than silently discarding personal facts.

The backlog lives separately in `content/games-backlog.json`. It appears as a quiet,
name-only list after the played history and is excluded from the played-game count
and stats synchronization.

## Steam

Set `STEAM_API_KEY` and your numeric SteamID64 (`STEAM_ID`) in your shell or private
environment manager. Obtain a personal key from https://steamcommunity.com/dev/apikey.
Do not commit it. `.env` is ignored by Git. `make games-sync` loads it when present;
`go run ./cmd/games-sync` uses only the current shell environment.

```bash
export STEAM_ID='your-numeric-steamid64'
read -rs -p 'Steam API key: ' STEAM_API_KEY
export STEAM_API_KEY
go run ./cmd/games-sync --provider steam --dry-run
```

The official [GetOwnedGames endpoint](https://partner.steamgames.com/doc/webapi/iplayerservice)
is called through `api.steampowered.com`, with an AppID filter. Game details must be
visible to the key/account. The sync imports only explicitly mapped archive games,
never the whole owned collection. Minutes become hours rounded to two decimals;
`rtime_last_played`, when present and positive, becomes UTC RFC3339. An absent
last-played value does not clear an earlier observation from the same provider.

Live Steam access was verified on 2026-09-26 and returned observations for nine
mapped games. Elden Ring and The Blood of Dawnwalker were absent from the response;
their existing data was preserved. Missing/private responses, HTTP failures and
transport errors are also covered by fixtures. Provider observations can change
between syncs; tests assert personal facts independently of those mutable totals.

## Lutris (read only)

Inspection of this machine on 2026-09-26 found:

- Flatpak Lutris **0.5.22**.
- Database: `~/.var/app/net.lutris.Lutris/data/lutris/pga.db`.
- `games.id` is the local integer primary key; `name` and `slug` identify the entry.
- `playtime` is a nullable REAL measured in **hours**.
- `lastplayed` is a nullable Unix timestamp. Installed `lutris/game.py` increments
  playtime by `timer.duration / 3600` and sets lastplayed from `time.time()`.
- Ghost of Tsushima: local ID **2**, slug `ghost-of-tsushima`, mapped in the archive.
- SOMA was **not present**. Its provider remains Lutris without an invented ID.
- Elden Ring also exists locally (ID 1), but retains your explicit Steam mapping.
  Totals from two providers are not added together.

The Go CLI invokes a small embedded Python 3 script using the standard-library
SQLite module. Python 3 is required for Lutris reads and the tests; the static site
and Steam reader remain Go-only. No additional Go SQLite/CGO dependency is needed.
The connection uses a file URI with `mode=ro`, `PRAGMA query_only=ON`, and a read
transaction. It queries only id/name/slug/playtime/lastplayed. There is no database
migration, repair, or write operation. Tests verify the source DB is unchanged.

Discovery checks compatible `.db` files in the Flatpak and native Lutris data
folders and `$XDG_DATA_HOME/lutris`, verifies required columns, and rejects ambiguous
matches. A custom Lutris `pga_path` is supported through an explicit override:

```bash
go run ./cmd/games-sync --list-lutris
go run ./cmd/games-sync --provider lutris --dry-run
go run ./cmd/games-sync --lutris-db /path/to/pga.db
# Or set LUTRIS_DB.
```

Once SOMA is installed/registered, use `--list-lutris` and record its real integer
`game_id`. After replacing a Lutris database, recheck IDs before syncing; they are
stable within that local database, not global game identifiers.

## Failure behaviour

A missing credential, network failure, private Steam library, unavailable Lutris
DB, unmatched game, or absent playtime preserves previous stats. Successful games
can update even when the other provider fails. Provider failures return exit status
1 after saving successful updates; warnings name unmatched games. Dry run never
writes. Cache output is deterministic and replaced atomically after a complete
write and flush. Unchanged syncs do not rewrite it.

Optional flags: `--games`, `--stats`, `--provider all|steam|lutris`, `--lutris-db`,
`--list-lutris`, `--dry-run`. Output cannot alias the personal archive.
Run only one sync at a time; no background scheduler or daemon is installed.
