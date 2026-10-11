# Photos

Each Markdown file describes one photograph and generates one page. Organizing
files into `YYYY/MM/` folders is recommended:

```text
content/photos/2026/07/my-desk-before-the-odyssey.md
```

```markdown
---
title: My desk before leaving for The Odyssey
date: 2026-07-24
slug: my-desk-before-the-odyssey
image: /images/photos/2026/07/my-desk.webp
image_alt: A desk with a notebook and keys before leaving home
caption: Everything left where I would find it later.
draft: false
---
```

`title`, `date`, `image`, and `image_alt` are required. `slug`, `caption`, and
`draft` are optional. When omitted, the slug is generated from the title.

Put photograph files under `content/images/`; their directory structure is
preserved when copied to `/images/`.

The archive shows linked titles grouped by year and month. The individual page
contains only the title, date, photograph, and optional caption.

Keep the photographs personal without identifying people or private surroundings.
Prefer objects, computers, books, projects, and ordinary scenes. Review the actual
image for recognizable people, account screens, papers, addresses, licence plates,
and private location clues. Such an image needs explicit approval for that specific
image before public use. Availability is not permission to publish it.

Text-only memories are welcome; never use unrelated imagery as fake history.
Normalize orientation before adding an image. The site build strips EXIF/GPS,
XMP/IPTC, comments, and PNG text metadata from supported published raster copies
(JPEG, PNG, WebP), preserving compressed pixels and color profiles. Original files
are unchanged. Other formats still need a metadata review; stripping metadata
does not remove identifying information visible in the picture.

Do not commit new sensitive originals. `content/private/` is excluded from site
asset copying and ignored by Git, but neither action removes earlier Git history.
See the repository's `AGENTS.md` for the editorial privacy principle.

The photo index uses a dark contact sheet grouped by year and month. Each thumbnail
links to its full photograph page; cropped previews do not change the source image.
An empty archive remains empty until approved photo entries are added. Test fixtures
for the contact sheet are not published as real memories.
