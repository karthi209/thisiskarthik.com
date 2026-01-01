# Content Creation Guide

## Writings (Blog Posts)

Writings are created as Markdown files in `content/posts/` organized by year.

### Structure

```
content/posts/
  └── YYYY/
      └── your-post-title.md
```

### Format

Each post is a Markdown file with YAML frontmatter:

```markdown
---
title: "Your Post Title"
date: "2024-01-15"
category: "life"
slug: "your-post-title"
draft: false
cover_image: "/images/your-image.png"  # optional
---

Your post content in Markdown goes here.

You can use **bold**, *italic*, and [links](https://example.com).

## Headings

- Bullet points
- More points

1. Numbered lists
2. Work too
```

### Frontmatter Fields

- **title** (required): The post title
- **date** (optional): Publication date in `YYYY-MM-DD` format. If omitted, uses file modification time
- **category** (optional): Post category (defaults to "life")
- **slug** (optional): URL slug. If omitted, generated from title
- **draft** (optional): Set to `true` to exclude from published site
- **cover_image** (optional): Path to cover image

### Example

Create `content/posts/2024/my-first-post.md`:

```markdown
---
title: "My First Post"
date: "2024-01-15"
category: "writing"
slug: "my-first-post"
---

This is my first blog post. Welcome!
```

Then run `make generate` to build the site.

---

## Notes

**Status**: Currently not implemented. The notes page exists but doesn't read from files yet.

To implement, you would need to:
1. Create `content/notes/` directory
2. Add markdown files similar to posts
3. Update `generate.go` to process notes similar to posts

---

## Library

**Status**: Currently not implemented. The library page exists but doesn't read from files yet.

To implement, you would need to:
1. Create `content/library/` directory  
2. Add markdown or JSON files for library items
3. Update `generate.go` to process library items

---

## Images

Place images for blog posts in `content/images/`. They will be copied to `public/images/` during generation.

### Recommended Format: WebP

**For best performance, use WebP format:**
- 25-50% smaller file sizes than PNG/JPEG
- Better compression with same quality
- Excellent browser support (97%+)
- Faster page loads

**Supported formats:** `.webp`, `.svg`, `.jpg`, `.jpeg`, `.png`, `.gif`

**Conversion example:**
```bash
# Convert PNG to WebP (requires cwebp tool)
cwebp -q 80 input.png -o output.webp

# Or use ImageMagick
convert input.png -quality 80 output.webp
```

### Recommended Resolutions

**For optimal performance and quality:**

#### Cover Images (Blog Post Headers)
- **Resolution:** 1200×630px (1.9:1 aspect ratio)
- **File size target:** < 150KB
- **Use case:** Featured images, social sharing

#### Inline Content Images
- **Resolution:** 1200×800px max (or maintain aspect ratio)
- **File size target:** < 200KB per image
- **Use case:** Images within blog post content
- **Note:** Images scale down automatically, so larger is fine if needed

#### Small Illustrations/Icons
- **Resolution:** 400×400px or smaller
- **File size target:** < 50KB
- **Use case:** Decorative elements, small graphics

#### General Guidelines
- **Maximum width:** 1200px (matches typical content width)
- **Aspect ratio:** Maintain original (avoid distortion)
- **Retina displays:** 2x resolution is acceptable (e.g., 2400px wide for 1200px display)
- **Compression:** Quality 75-85 for WebP/JPEG (good balance)

**Example workflow:**
```bash
# Resize and convert to WebP
convert input.jpg -resize 1200x -quality 80 output.webp

# Or with cwebp (after resizing separately)
cwebp -q 80 -resize 1200 0 input.jpg -o output.webp
```

### Organization

**Recommended structure:**
```
content/images/
  ├── covers/          # Cover/featured images (1200×630px)
  └── posts/           # Blog post content images (max 1200px width)
```

**Simple alternative (flat):**
```
content/images/
  └── your-image.webp  # All images in one folder
```

See `docs/IMAGE_ORGANIZATION.md` for detailed organization guide.

### Usage

Reference images in your markdown:
```markdown
![Alt text](/images/posts/your-image.webp)
# or for flat structure:
![Alt text](/images/your-image.webp)
```

**Note:** All formats are accepted, but WebP is strongly recommended for performance. Images automatically scale to fit the content width.

