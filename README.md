# Hmm... Blog?

**A personal blog that probably won't break in five years.**

---

## What is this?

This is the source code for my personal website. I built it because I got tired of platforms that change their terms of service every Tuesday and frameworks who's dependencies break every friday.

It's just HTML and CSS compiled into static pages. No databases, no JavaScript frameworks that will be obsolete next month, no external services that will shut down and leave me stranded. It's boring, and that's the point.

## Why?

Most websites are built like houses of cards. This one is built like a brick. It might not be flashy, but it'll probably still work when everything else has broken.

- **Static**: Pure HTML and CSS. No server-side anything. If it works on my machine, it works everywhere.
- **Readable**: I care more about whether you can actually read the text than whether the animations are smooth.
- **Independent**: Doesn't rely on any external services. If GitHub goes down, I can host this on a toaster.

## Typography

I spent way too much time picking fonts. Here's what I ended up with:

- **IM Fell Great Primer**: Used for headings. It's old, it's imperfect.
- **Libertinus Serif**: Used for body text. It's calm, readable, and doesn't scream at you. A rare quality these days.

Both fonts are served locally because I don't trust Google Fonts to still exist in 2030.

## Structure

The repository structure is pretty straightforward. If you can't figure it out, that's on you.

- `content/` : The actual writing (Markdown files)
- `templates/` : How pages get assembled (Go HTML templates)
- `static/` : CSS, fonts, images, the usual suspects
- `public/` : The compiled output, ready to deploy

## How to Build

The site uses a custom static site generator written in Go. It's intentionally simple. If you can't read the code and understand it in one sitting, I've failed.

To build it:

```bash
make setup      # Install dependencies (Go, imaging tools)
make generate   # Compile the site to /public directory
make serve      # Preview locally
```

To clean up:

```bash
make clean
```

