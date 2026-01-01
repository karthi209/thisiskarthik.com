# Blog Posts Directory

This directory contains all blog posts in Markdown format.

## Structure

Organize posts by year/month for better scalability:

```
content/posts/
  ├── 2024/
  │   ├── 01/
  │   │   ├── my-first-post.md
  │   │   └── another-post.md
  │   └── 02/
  │       └── february-post.md
  └── 2025/
      └── 01/
          └── new-year-post.md
```

Or organize by category:

```
content/posts/
  ├── tech/
  │   ├── post-1.md
  │   └── post-2.md
  ├── life/
  │   └── post-3.md
  └── music/
      └── post-4.md
```

## Markdown Format

Each post should have frontmatter at the top:

```markdown
---
title: "My Blog Post Title"
category: tech
date: 2024-01-15
slug: my-blog-post-title
cover_image: /images/covers/my-cover.jpg
draft: false
edition: "v1.0"
---

Your markdown content here...
```

### Frontmatter Fields

- `title` (required): Post title
- `category` (optional): One of: tech, life, music, games, movies, tv, books (default: life)
- `date` (optional): Publication date (ISO format or YYYY-MM-DD)
- `slug` (optional): URL slug (auto-generated from title if not provided)
- `cover_image` (optional): Path to cover image (relative to /images/)
- `draft` (optional): Set to `true` for draft posts
- `edition` (optional): Edition/version string

