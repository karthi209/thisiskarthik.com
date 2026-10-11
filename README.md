# for later, when i forget

a small static journal by karthik.

this site is built for short daily notes first: ordinary things, remembered quickly, laid out on a compact ink timeline. essays still have their own archive, but the homepage stays personal and dense instead of turning into a newspaper.

## the shape

- `content/journal/` holds small timeline notes.
- `content/posts/` holds longer essays.
- `content/books.json` holds the reading log at `/books`.
- `content/games.json` holds the personal game archive at `/games`; optional `content/game-stats.json` holds provider observations (workflow in `content/GAMES.md`).
- `content/photos/` holds photograph records and captions.
- `content/images/` holds images used across the site.
- `templates/` holds the go html templates.
- `static/` holds the few permanent assets: one stylesheet, local fonts, the favicon, and small notebook illustrations.
- `public/` is generated output and is safe to rebuild.

## writing

make a new small note:

```bash
make note
```

the smallest useful note looks like this:

```markdown
---
date: 2026-07-25
---

something i want to remember.
```

optional fields:

```yaml
title: kept this moment
time: 22:10
mood: content enough
tags: life, chennai
draft: false
```

longer pieces go in `content/posts/`. they appear only in the essays section, not on the homepage timeline.

the game archive and photographs have their own small formats documented in
`content/GAMES.md` and `content/photos/README.md`. The reading log is the small,
self-explanatory collection in `content/books.json`.

## design

the site uses a charcoal background, warm cream text, thin border boxes, and slightly offset frames. IBM Plex Sans is the reading face; Plex Mono carries navigation, dates, and metadata. Fraunces headings and small Caveat annotations add character without competing with the writing. Fonts are local, with their licenses in `static/fonts/`.

`static/css/site.css` is the single stylesheet, shared through `site-styles` in `templates/partials.html`. Timeline notes, essays, books, games, photos, and prose pages use the same design. Games opens on the cover shelf; the journal and playtime views remain available. Content keeps its original capitalization.

## responsive behavior

the layout uses one compact reading column. On mobile the date rail narrows, navigation wraps, shelves and photos use two columns, and prose keeps room to breathe. Keyboard focus and reduced-motion preferences are supported.

`make check` validates generated pages, local fonts, navigation, assets, rss, and removal of obsolete CSS and legacy routes.

## commands

```bash
make build      # generate public/
make check      # tests, build, and generated-site validation
make test       # go tests for timeline parsing and ordering
make games-sync # reconcile Steam + Lutris; completed records stay manual
make games-import # alias for games-sync
make note       # create today's short journal note
make preview    # local dev server on port 5174
make serve      # same as preview
make clean      # remove public/
make optimize   # optimize content/images to webp
make deploy     # build and deploy to github pages
```

manual equivalents:

```bash
go run generate.go
go run serve.go
```

## deploy checklist

```bash
make check
make deploy
```

the deploy target builds the static site and publishes `public/`.
