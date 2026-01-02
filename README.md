# On Matters Local and Otherwise

**An archival website and personal journal.**

---

## Preface

This repository contains the source code and content for *On Matters Local and Otherwise*, a personal website designed to function as a long-term digital archive built on the conviction that a personal website should be a quiet place not a product, platform, or a growth engine. I built is as my digital book. Static, readable, and preservable.

## Philosophy

The site is constructed to withstand the passage of time. It avoids complex build chains, client-side frameworks, and external dependencies that rot.

- **Static**: The site compiles to pure HTML and CSS. It requires no database and no runtime server logic.
- **Readable**: Typography and layout are prioritized above all else.
- **Independent**: It relies on no external platforms for its existence.

## Typography

The typographic voice of this website is specific and intentional.

- **IM Fell Great Primer**: Used for headings and titles. This is a revival of type cut by Dirck Voskens in the 17th century and later used by William Caslon. It provides the authorial voice—imperfect, historical, and human.
- **Libertinus Serif**: Used for body text. Chosen for its calm, ink-like quality and readability in long-form passages.

Both typefaces are served locally to ensure permanence.

## The Archive

The structure of this repository is intended to be self-evident to any future archivist.

- `content/` : The writing itself, stored as Markdown files.
- `templates/` : The logic for how pages are assembled (Go HTML templates).
- `static/` : The permanent assets (CSS, fonts, images).
- `public/` : The compiled result, ready for reading.

## Operation

The site uses a custom static site generator written in Go (`generate.go`). It is designed to be simple enough to read in one sitting.

To build the archive from source:

```bash
make setup      # Install dependencies (Go, imaging tools)
make generate   # Compile the site to internal /public directory
make serve      # Preview the site locally
```

To clean the workspace:

```bash
make clean
```

---

*First Digital Edition.*
*Compiled with care.*

