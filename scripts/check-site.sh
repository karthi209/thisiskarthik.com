#!/usr/bin/env bash
set -eu

if [ ! -f public/index.html ] || [ ! -f public/rss.xml ] || [ ! -f public/css/site.css ]; then
  echo "Generated site is missing required files." >&2
  exit 1
fi

page_count=0
while IFS= read -r page; do
  page_count=$((page_count + 1))
  grep -q 'name="viewport"' "$page"
done < <(find public -name index.html -type f -print)

if [ "$page_count" -lt 6 ]; then
  echo "Expected at least six generated HTML pages; found $page_count." >&2
  exit 1
fi

if grep -RqiE 'user-scalable=no|maximum-scale=1' public templates --include='*.html'; then
  echo "Mobile zoom restrictions leaked into a page." >&2
  exit 1
fi

for page in \
  public/games/index.html \
  public/books/index.html \
  public/photos/index.html
do
  if [ ! -f "$page" ]; then
    echo "Generated site is missing $page." >&2
    exit 1
  fi
done

grep -q "@media (prefers-reduced-motion: reduce)" public/css/site.css
grep -q "class=\"timeline\"" public/index.html
grep -q "daily journal" public/index.html
grep -q "life, more or less in the order it happened" public/index.html
grep -q 'class="timeline-origin">the beginning' public/index.html
grep -q "for later, when i forget" public/index.html
grep -q "masthead-byline\">by karthik" public/index.html
grep -q "class=\"timeline essay-timeline\"" public/essays/index.html
grep -q 'href="/games"' public/index.html
grep -q 'href="/books"' public/index.html
grep -q 'href="/photos"' public/index.html
grep -q 'aria-current="page">games' public/games/index.html
grep -q 'aria-current="page">books' public/books/index.html
grep -q 'aria-current="page">photos' public/photos/index.html
grep -q 'id="reading-heading"' public/books/index.html
grep -q 'id="planned-heading"' public/books/index.html
grep -q "The Left Hand of Darkness" public/books/index.html
grep -q "The Last Wish" public/books/index.html
grep -q "The Discovery of India" public/books/index.html
grep -q 'class="book-cover"' public/books/index.html
grep -q 'src="/images/books/the-left-hand-of-darkness.webp"' public/books/index.html
test -s public/images/books/the-left-hand-of-darkness.webp
test -s public/images/books/the-discovery-of-india.webp
grep -q 'id="playing-heading"' public/games/index.html
grep -q 'data-game-view="playtime"' public/games/index.html
grep -q 'id="playtime-heading"' public/games/index.html
grep -q 'js/games.js' public/games/index.html
test -s public/js/games.js
grep -q 'class="game-year"' public/games/index.html
grep -q 'id="games-year-2026"' public/games/index.html
test -s public/images/games/elden-ring.jpg
test -s public/images/games/ghost-of-tsushima.jpg
test -s public/images/games/soma.jpg
grep -q "no photographs here yet" public/photos/index.html
grep -q "photo-figure img" public/css/site.css
grep -q "height: auto" public/css/site.css
grep -q 'data-archive-number="003"' public/index.html
grep -q 'data-archive-number="001"' public/index.html
grep -q "timeline-month-marker\">Jul" public/index.html
grep -q "timeline-month-marker\">Feb" public/index.html
grep -q "timeline-weekday\">thu" public/index.html
grep -q "memory-image" public/css/site.css
grep -q "image_caption" content/journal/README.md
grep -q "aspect-ratio: 4 / 3" public/css/site.css
grep -q 'content: "no. " attr(data-archive-number)' public/css/site.css
grep -q "@media (max-width: 620px)" public/css/site.css
grep -q "grid-template-columns: 54px minmax(0, 1fr)" public/css/site.css

if grep -RqiE 'timeline-colophon|❦' public templates --include='*.html'; then
  echo "Decorative colophon leftovers leaked into generated pages." >&2
  exit 1
fi

if grep -RqiE 'ART_CREDITS|art credits' public templates --include='*.html'; then
  echo "Art credits link leaked into generated pages." >&2
  exit 1
fi

if grep -q "timeline-item--essay" public/index.html; then
  echo "Essays leaked into the short-entry timeline." >&2
  exit 1
fi

if grep -q ">entries<" public/index.html; then
  echo "Dashboard-style stats leaked into the journal intro." >&2
  exit 1
fi

if grep -Eqi 'memory-details|memory-meta|memory-artifact|i.ll probably remember|i.ll probably forget|<dt>place|<dt>with|movie ticket / screen 4' public/index.html; then
  echo "Archive metadata leaked into the simple timeline." >&2
  exit 1
fi

if [ -d public/writings ] || grep -Rqi 'writings' public templates generate.go --include='*.html' --include='*.go'; then
  echo "Old writings terminology leaked into generated pages or source templates." >&2
  exit 1
fi

if grep -RqiE 'blog posts|Blog Posts|Writings|writings' README.md content/posts/README.md scripts/README.md; then
  echo "Old essay terminology leaked into documentation." >&2
  exit 1
fi

if grep -RqiE 'site-switcher|library\.thisiskarthik\.com|notes from chennai' public --include='*.html'; then
  echo "Removed site-switcher or Library content leaked into generated pages." >&2
  exit 1
fi

if grep -Eqi 'entry-warning|entity-notices|random-insight|handdrawn.css|editorial.css|readability.css' public/index.html; then
  echo "Experimental homepage design leaked into the permanent layout." >&2
  exit 1
fi

if find public/css -maxdepth 1 -type f ! -name site.css | grep -q .; then
  echo "Unused CSS files leaked into the generated site." >&2
  exit 1
fi

if [ -f public/js/home.js ]; then
  echo "Unused homepage script leaked into the generated site." >&2
  exit 1
fi

if grep -Eqi 'placeholder game|neocities mode' public/index.html; then
  echo "Placeholder homepage content leaked into the generated site." >&2
  exit 1
fi

if grep -Eqi 'camera model|aperture|shutter speed|exif|<dt>|masonry' public/photos/index.html templates/photo.html templates/photos.html; then
  echo "Portfolio or camera metadata leaked into Photos." >&2
  exit 1
fi



if grep -Eqi 'older snapshot|lifetime|hours unknown|>library</a>' public/games/index.html; then
  echo "Removed game UI leaked into the archive." >&2
  exit 1
fi
if [ -d public/library ] || rg -qi 'href="[^"]*/library(?:/|\")|>library</a>|library-(?:index|scope|timeline|entry|note)' public templates generate.go; then
  echo "Legacy library pages or UI leaked into the site." >&2
  exit 1
fi

for font in ibm-plex-sans-regular.woff2 ibm-plex-sans-semibold.woff2 ibm-plex-mono-regular.woff2 ibm-plex-mono-italic.woff2; do
  test -s "public/fonts/$font"
done
while IFS= read -r page; do
  grep -q 'css/site.css?v=1' "$page"
  grep -q 'name="color-scheme" content="dark"' "$page"
done < <(find public -name index.html -type f -print)
if rg -q 'permanent.css|flow.css|fonts.googleapis.com|site-wallpaper|spine-squiggle' public templates --glob '*.html' --glob '*.css'; then
  echo "Retired design assets leaked into the site." >&2
  exit 1
fi
echo "Validated $page_count site pages, local fonts, navigation, assets, and RSS."
