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
