# Daily journal

Each Markdown file in this folder becomes a note directly on the homepage timeline.
Organizing them into `YYYY/MM/` folders is recommended, but not required.

Only the date and the words are needed:

```markdown
---
date: 2026-07-25
---

The small thing I want to remember.
```

Optional fields:

```yaml
title: Evening light
time: 18:30
mood: content
tags: chennai, life
image: /images/2026/07/ticket.webp
image_alt: a small paper ticket kept from the day
image_caption: the ticket i almost threw away.
draft: false
```

The title is optional; when present, it appears in bold above the note. There is no
slug, category, or separate page to maintain. Run `make note` to create and open a
correctly named entry for today. Longer pieces belong in `content/posts/` and only
appear in the Essays section.

Use `image` only when the note needs one small artifact: a ticket, a photo, a
receipt, a map crop, something that helps the memory breathe. Most days can stay
plain.
