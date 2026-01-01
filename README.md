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
- `make deploy` - Build and deploy to GitHub Pages
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

## Deployment to GitHub Pages

### First-time Setup

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Pages**
3. Under **Source**, select **Deploy from a branch**
4. Choose **gh-pages** branch and **/ (root)** folder
5. Click **Save**

### Deploying

For **user/organization pages** (e.g., `username.github.io`):
```bash
make deploy
```

For **project pages** (e.g., `username.github.io/repo-name`), set the base path:
```bash
BASE_PATH="/repo-name/" make deploy
```

Or run the script directly:
```bash
BASE_PATH="/repo-name/" ./scripts/deploy.sh
```

**Note:** 
- The `BASE_PATH` environment variable sets the base path for all assets and links
- For user/organization pages, use `BASE_PATH="/"` (default)
- For project pages, use `BASE_PATH="/repo-name/"` (replace `repo-name` with your repository name)
- The script will build the site, create/update the `gh-pages` branch, and push to GitHub

Your site will be available at:
- User/org pages: `https://<username>.github.io/`
- Project pages: `https://<username>.github.io/<repo-name>/`

(It may take a few minutes for GitHub Pages to update after deployment)

