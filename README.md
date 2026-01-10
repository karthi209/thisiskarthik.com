# The Book of Odds and Ends

**A personal blog with a little bit of character.**

---

## What is this?

This is the source code for my personal website, and I built it because I got tired of platforms that change their terms of service every Tuesday. I post a lot on Twitter, and I realized late that you don't actually own anything you post there, so everything I write online, this website will have a copy of it all, and it's just HTML and CSS compiled into static pages using a custom Go static site generator... just my words on my domain under my control.

## Structure

The repository structure is pretty straightforward, and it's organized into a few key directories that make sense when you look at them.

- `content/` : The actual writing (Markdown files)
- `templates/` : How pages get assembled (Go HTML templates)
- `static/` : CSS, fonts, images, the usual stuff
- `public/` : The compiled output, that we deploy to static servers like gtihub pages and yadayada

## How to Build

The site uses a custom static site generator written in Go, and it's intentionally simple, and if you can't read the code and understand it in one sitting, I've failed.

### Build Commands

```bash
make setup      # Install dependencies (Go, imaging tools)
make generate   # Compile the site to /public directory
make serve      # Dev server with hot reload (port 5174)
make clean      # Remove generated files
```

### Manual Build

```bash
go run generate.go  # Build site
go run serve.go     # Dev server
```

## Tech Stack

- **Generator**: Custom Go static site generator
- **Markdown**: Goldmark for parsing
- **Templates**: Go's `html/template` package
- **Font**: Bricolage Grotesk (Google Fonts)
- **Styling**: Pure CSS, no frameworks
- **Deployment**: Static files, works with GitHub Pages, Netlify, etc.

## Design

- **Theme**: Dark mode only (#1e1e1e background, #e8e8e8 text)
- **Accent**: Cyan (#00d9ff) for links and highlights
- **Typography**: Bricolage Grotesk - modern geometric sans-serif
- **Layout**: Clean, minimal, inspired by seated.ro aesthetic
- **Responsive**: Mobile, tablet, and desktop optimized

