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
draft: false
```

The title is optional; when present, it appears in bold above the note. There is no
slug, category, or separate page to maintain. Run `make note` to create and open a
correctly named entry for today. Longer pieces belong in `content/posts/` and only
appear in the Essays section.
