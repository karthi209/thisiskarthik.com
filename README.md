# Theriyala, But Moving

**A personal journal with a little bit of character.**

---

## What is this?

This is the source code for my personal blogsite, and I built it because I got tired of platforms that change their terms of service every Tuesday. I post a lot on Twitter, and I realized late that you don't actually own anything you post there, so everything I write online, this website will have a copy of it all, and it's just HTML and CSS compiled into static pages using a custom Go static site generator... just my words on my domain under my control.

This journal is one half of a pair. Structured ratings, reviews, and collection entries live in [Karthik's Library](https://library.thisiskarthik.com/), while broader essays—including essays about media—stay here. The sites share an identity and theme preference, but they build and deploy independently.

## Structure

The repository structure is pretty straightforward, and it's organized into a few key directories that make sense when you look at them.

- `content/` : The actual writing (Markdown files)
- `templates/` : How pages get assembled (Go HTML templates, with shared shell partials)
- `static/` : CSS, fonts, images, the usual stuff
- `public/` : The compiled output, that we deploy to static servers like gtihub pages and yadayada

## How to Build

The site uses a custom static site generator written in Go, and it's intentionally simple, and if you can't read the code and understand it in one sitting, I've failed.

### Build Commands

```bash
make build      # Compile the site to /public directory
make check      # Build and validate pages, navigation, theme assets, and RSS
make preview    # Dev server with hot reload (port 5174)
make setup      # Install dependencies (Go, WebP, ImageMagick)
make generate   # Compile the site to /public directory
make serve      # Dev server with hot reload (port 5174)
make optimize   # Optimize images to WebP
make deploy     # Build and deploy to GitHub Pages
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
- **Font**: Patrick Hand for prose, Inter for interface text, Kalam for English display, Kavivanar for Tamil, and the system monospace stack for metadata
- **Styling**: Pure vanilla CSS, no frameworks
- **Deployment**: Simple static directory, GitHub Pages target

## Design

- **Theme**: Paper-and-ink light and dark palettes shared conceptually with the Library.
- **Accent**: Purple, pink, yellow, and cyan handwritten details around a quieter reading surface.
- **Typography**: Patrick Hand for prose, Inter for interface text, Kalam for English display, and Kavivanar for Tamil.
- **Navigation**: A shared Karthik site switcher connects the independently deployed Journal and Library.
- **Responsive**: Reading widths and controls adapt for desktop, tablet, and mobile.
