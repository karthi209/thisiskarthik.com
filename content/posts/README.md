# essays directory

This directory contains longer essays in Markdown format.

## Structure

Organize essays by year/month for better scalability:

```
content/posts/
  ├── 2024/
  │   ├── 01/
  │   │   ├── my-first-essay.md
  │   │   └── another-essay.md
  │   └── 02/
  │       └── february-essay.md
  └── 2025/
      └── 01/
          └── new-year-essay.md
```

Or organize by category:

```
content/posts/
  ├── tech/
  │   ├── essay-1.md
  │   └── essay-2.md
  ├── life/
  │   └── essay-3.md
  └── music/
      └── essay-4.md
```

## Markdown Format

Each essay should have frontmatter at the top:

```markdown
---
title: "My Essay Title"
category: tech
date: 2024-01-15
slug: my-essay-title
draft: false
edition: "v1.0"
---

Your markdown content here...
```

### Frontmatter Fields

- `title` (required): Essay title
- `category` (optional): One of: tech, life, music, games, movies, tv, books (default: life)
- `date` (optional): Publication date (ISO format or YYYY-MM-DD)
- `slug` (optional): URL slug (auto-generated from title if not provided)
- `draft` (optional): Set to `true` for draft essays
- `edition` (optional): Edition/version string
