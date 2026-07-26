# Library

The library is a chronological record of things watched, played, read, and
listened to. Entries live in one of these folders:

```text
content/library/films/
content/library/tv/
content/library/games/
content/library/books/
content/library/music/
```

Folders may be divided by year and month. A film entry, for example, can live at
`content/library/films/2026/07/the-odyssey.md`.

```markdown
---
title: The Odyssey
release_year: 2026
date: 2026-07-24
rating: 4
draft: false
---

Worth seeing on the biggest screen possible.

Unfortunately I watched it in a theatre with terrible speakers.
```

`title`, `date`, `rating`, and the personal note are required. Use a full date
when it is known; a year such as `2025` is also accepted when the exact day has
been forgotten. Ratings run from 0.5 to 5 in half-star steps. `release_year` and
`draft` are optional.

Entries appear newest first on their section page. They do not have individual
pages, posters, covers, or additional metadata.
