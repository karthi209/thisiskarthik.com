# Books

`content/books.json` contains the reading log. Each record has a unique `id`,
`title`, `author`, and `status`: `reading`, `read`, or `planned`. Covers are local
reviewed artwork. Notes are optional.

Optional fields:

- `rating`: a number from 0 to 5 in half-star steps; omit it when unrated.
- `started` and `finished`: `YYYY`, `YYYY-MM`, or `YYYY-MM-DD`. Keep only the
  precision you remember. Do not use publication dates as personal reading dates.
- `pages`: the positive page count of the edition actually read, when known.

The overview counts reading statuses, distinct author labels across all shelves,
and assigned ratings. The author chart shows the five most represented authors
across the shelves, including planned books; it does not imply those books were
finished. Average rating includes assigned zero ratings and leaves unrated books
out. The histogram uses the same half-star bins as games.

Read-by-year charts appear only for `read` books with a recorded `finished` date.
Pages recorded sums known page counts of `read` books; missing page counts stay
unknown. Neither statistic is inferred from covers, editions, or the build date.

Cards show optional dates and pages when present. The Goodreads link searches by
title and author; it does not connect an account or import reading history.
The overview is built as static HTML with shared collection styles and no chart
library or third-party runtime requests.
