# Image Organization Guide

## Recommended Structure

For a minimal, performant static site, organize images as follows:

```
content/images/
  ├── covers/          # Cover/featured images (1200×630px)
  │   ├── post-slug-1.webp
  │   └── post-slug-2.webp
  │
  └── posts/           # Blog post content images (max 1200px width)
      ├── 2024/        # Optional: organize by year
      │   └── image-name.webp
      └── 2025/
          └── image-name.webp
```

## Simple Flat Structure (Alternative)

If you prefer minimal organization:

```
content/images/
  ├── cover-post-slug.webp    # Cover images
  └── post-image-name.webp    # Content images
```

## Best Practices

### 1. **Naming Convention (Book-Style for Maximum Efficiency)**

For maximum efficiency and clarity, use a **book-like sequential naming system**:

**Pattern:** `post-slug-figure-01.webp`, `post-slug-figure-02.webp`, etc.

```
✅ Best (Book-Style Sequential):
- we-need-to-fix-out-footpaths-figure-01.webp
- we-need-to-fix-out-footpaths-figure-02.webp
- we-need-to-fix-out-footpaths-figure-03.webp

✅ Good (Descriptive + Number):
- footpath-broken-sidewalk-01.webp
- footpath-wheelchair-road-02.webp
- footpath-parking-issue-03.webp

✅ Acceptable (Simple Sequential):
- footpath-01.webp
- footpath-02.webp
- footpath-03.webp

❌ Avoid:
- IMG_1234.jpg (no context)
- image (1).png (spaces, parentheses)
- MyImage!.webp (uppercase, special chars)
- 04.webp (no descriptive prefix)
```

**Why book-style naming?**
- **Sequential numbering** makes it easy to reference ("see figure 3")
- **Post slug prefix** groups images by post automatically
- **Zero-padded numbers** (01, 02, not 1, 2) sort correctly
- **Consistent pattern** makes images easy to find and reference

### 2. **Organization by Use Case**

**Cover Images** (`content/images/covers/`):
- Featured images for blog posts
- Social sharing images
- 1200×630px, optimized for WebP
- Reference in frontmatter: `cover_image: /images/covers/post-slug.webp`

**Content Images** (`content/images/posts/`):
- Images within blog post content
- Max 1200px width
- Reference in markdown: `![Alt text](/images/posts/image-name.webp)`

### 3. **Optional: Organize by Year**

For easier management with many images:

```
content/images/posts/
  ├── 2024/
  │   ├── urban-planning-1.webp
  │   └── transit-notes.webp
  └── 2025/
      └── footpath-issues.webp
```

**Pros:** Easier to find images by year  
**Cons:** Slightly longer paths in markdown

### 4. **Workflow**

1. **Add images** to `content/images/posts/` (or subdirectories)
2. **Optimize** with the script:
   ```bash
   # Optimize content images
   ./scripts/optimize-images.sh -t content content/images/posts/
   
   # Optimize cover images
   ./scripts/optimize-images.sh -t cover content/images/covers/
   ```
3. **Reference in markdown**:
   ```markdown
   ![Description](/images/posts/image-name.webp)
   ```
4. **Generate site** - images are automatically copied to `public/images/`

## Current Structure

Your current flat structure works fine:

```
content/images/
  ├── footpath-01.jpg
  ├── footpath-02.jpg
  └── ...
```

**To reference:**
```markdown
![Footpath issue](/images/footpath-01.webp)
```

## Migration Path

If you want to organize existing images:

```bash
# Create organized structure
mkdir -p content/images/covers content/images/posts

# Move existing images (if needed)
mv content/images/footpath-*.jpg content/images/posts/

# Optimize to WebP
./scripts/optimize-images.sh -t content content/images/posts/
```

## Recommendations

**For minimalism and performance:**
- ✅ Use WebP format (optimize with script)
- ✅ Descriptive filenames
- ✅ Flat structure in `posts/` (no year subdirs unless you have 100+ images)
- ✅ Separate `covers/` folder for featured images
- ✅ Keep total image count manageable (organize by year only if needed)

**File size targets:**
- Content images: < 200KB each
- Cover images: < 150KB each
- Total per post: < 1MB (for fast loading)

## Example Post Structure

```
content/
  ├── posts/
  │   └── 2025/
  │       └── 05/
  │           └── my-post.md
  └── images/
      ├── covers/
      │   └── my-post-cover.webp
      └── posts/
          ├── my-post-image-1.webp
          └── my-post-image-2.webp
```

**In markdown:**
```markdown
---
title: "My Post"
cover_image: /images/covers/my-post-cover.webp
---

![First image](/images/posts/my-post-image-1.webp)

![Second image](/images/posts/my-post-image-2.webp)
```

## Book-Style Naming: Practical Example

For your current post "we-need-to-fix-out-footpaths-this-is-not-ok":

**Recommended naming:**
```
content/images/2025/05/
  ├── footpaths-figure-01.webp  (broken footpath)
  ├── footpaths-figure-02.webp  (wheelchair on road)
  ├── footpaths-figure-03.webp  (parking issue)
  └── footpaths-figure-04.webp  (another example)
```

**Why this works:**
1. **Short prefix** (`footpaths-`) - easy to type, groups related images
2. **Sequential numbering** - matches book figure references
3. **Zero-padded** (`01`, not `1`) - sorts correctly in file managers
4. **Consistent pattern** - easy to remember and reference

**In your markdown:**
```markdown
![Broken Footpaths](/images/2025/05/footpaths-figure-01.webp)

![Consequence](/images/2025/05/footpaths-figure-02.webp)
```

**Alternative (even shorter):**
```
footpath-01.webp
footpath-02.webp
footpath-03.webp
```

This is fine if you only have one type of image per post. Use the longer `post-slug-figure-XX` pattern if you have multiple image types.

