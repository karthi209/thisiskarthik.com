# for later, when i forget

a small static journal by karthik.

this site is built for short daily notes first: ordinary things, remembered quickly, laid out on a compact ink timeline. essays still have their own archive, but the homepage stays personal and dense instead of turning into a newspaper.

## the shape

- `content/journal/` holds small timeline notes.
- `content/posts/` holds longer essays.
- `content/library/` holds chronological film, television, game, book, and music notes.
- `content/photos/` holds photograph records and captions.
- `content/images/` holds images used across the site.
- `templates/` holds the go html templates.
- `static/` holds the few permanent assets: css, favicon, theme script, and the squiggly spine.
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

library entries and photographs have their own small formats documented in
`content/library/README.md` and `content/photos/README.md`.

## design

the design is intentionally plain white paper and warm ink. there are no cards, no editorial boxes, no decorative backgrounds, and no art-credit trail pretending the site is more illustrated than it is.

the timeline spine, month/date labels, quote marks, and small metadata use the same hand-drawn language. the rendered site is lowercase on purpose, including essay titles, because the voice should feel like a journal rather than a publication.

## responsive behavior

the layout is designed around one reading column:

- desktop: wide enough for comfortable essay reading, with the spine offset beside the notes.
- tablet: the same timeline rhythm, with tighter spacing and no layout shift between timeline and essays.
- mobile: a narrower date rail, shorter weekday labels, and compact note spacing so the page scans without horizontal scrolling.

`make check` validates the generated pages, key assets, responsive-friendly timeline markup, rss, removed legacy links, and cleanup rules for unused css/js.

## commands

```bash
make build      # generate public/
make check      # tests, build, and generated-site validation
make test       # go tests for timeline parsing and ordering
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
