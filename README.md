# On Matters Local and Otherwise - Static Site

A static site generator for the blog, built with Go. This is a complete rewrite of the React-based frontend into a static HTML/CSS/Go architecture while preserving the exact visual design.

## Structure

- `content/posts/` - Markdown blog posts with YAML frontmatter
- `content/images/` - Images referenced in posts
- `templates/` - Go HTML templates for each page type
- `static/` - Static assets (CSS, images, fonts, etc.)
- `public/` - Generated static HTML site (this is what you deploy)

## Setup

One-command setup to install all dependencies:

```bash
make setup
```

This will install:
- Go (golang)
- WebP tools (cwebp)
- ImageMagick
- Go module dependencies

## Quick Start

```bash
# 1. Install dependencies
make setup

# 2. Optimize images (optional, but recommended)
make optimize

# 3. Generate the static site
make generate

# 4. Serve locally
make serve
```

## Commands

All operations go through `make`:

- `make setup` - Install all dependencies (Go, WebP tools, ImageMagick, Go modules)
- `make optimize` - Optimize all images in `content/images/` to WebP format
- `make generate` - Generate the static site (output in `./public/`)
- `make build` - Alias for `generate`
- `make serve` - Generate and serve site locally on http://localhost:5173
- `make clean` - Remove generated public directory
- `make help` - Show all available commands

## Development

1. Add/edit markdown files in `content/posts/`
2. Add images to `content/images/` and run `make optimize` to convert them to WebP
3. Run `make generate` to rebuild
4. Run `make serve` to generate and serve locally (uses Go HTTP server)

## Font License

This site uses **IM Fell DW Pica** font by Igino Marini, licensed under the SIL Open Font License, Version 1.1.

- **Copyright**: (c) 2010, Igino Marini (mail@iginomarini.com)
- **License**: SIL Open Font License, Version 1.1
- **Full license text**: See `static/fonts/LICENSE.txt` or `public/fonts/LICENSE.txt`
- **License FAQ**: https://openfontlicense.org

The font files and license are included in the `static/fonts/` directory and are automatically copied to the output during generation, ensuring compliance with the license requirements.

