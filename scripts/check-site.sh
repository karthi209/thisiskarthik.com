#!/usr/bin/env bash
set -eu

if [ ! -f public/index.html ] || [ ! -f public/rss.xml ] || [ ! -f public/css/permanent.css ] || [ ! -f public/spine-squiggle.svg ]; then
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
  public/library/index.html \
  public/library/films/index.html \
  public/library/games/index.html \
  public/games/index.html \
  public/books/index.html \
  public/library/books/index.html \
  public/photos/index.html
do
  if [ ! -f "$page" ]; then
    echo "Generated site is missing $page." >&2
    exit 1
  fi
done

grep -q "animation: none" public/css/permanent.css
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
grep -q 'href="/library/films"' public/library/index.html
grep -q 'href="/games"' public/library/index.html
grep -q 'href="/books"' public/library/index.html
grep -q 'class="section-deck"' public/library/index.html
grep -q "worlds explored, finished or left unfinished" public/library/index.html
grep -q "class=\"library-index\"" public/library/index.html
grep -q "4 books" public/library/index.html
grep -q 'id="reading-heading"' public/books/index.html
grep -q "The Left Hand of Darkness" public/books/index.html
grep -q "The Last Wish" public/books/index.html
grep -q "Blood Diamond" public/library/films/index.html
grep -q "Red Dead Redemption 2" public/library/games/index.html
grep -q 'id="playing-heading"' public/games/index.html
grep -q 'data-game-view="playtime"' public/games/index.html
grep -q 'id="playtime-heading"' public/games/index.html
grep -q 'js/games.js' public/games/index.html
test -s public/js/games.js
grep -q 'class="game-year"' public/games/index.html
grep -q 'id="games-year-2026"' public/games/index.html
grep -Eq '[0-9]+ games' public/library/index.html
test -s public/images/games/elden-ring.jpg
test -s public/images/games/ghost-of-tsushima.jpg
test -s public/images/games/soma.jpg
grep -q "films remembered" public/library/films/index.html
grep -q "class=\"library-scope\"" public/library/films/index.html
grep -Eq "<span>[0-9]+ entries</span>" public/library/films/index.html
if grep -q "data-archive-number" public/library/films/index.html; then
  echo "Library accession numbers leaked into the films archive." >&2
  exit 1
fi
grep -q "no photographs here yet" public/photos/index.html
grep -q "photo-figure img" public/css/permanent.css
grep -q "height: auto" public/css/permanent.css
grep -q 'data-archive-number="003"' public/index.html
grep -q 'data-archive-number="001"' public/index.html
grep -q "timeline-month-marker\">Jul" public/index.html
grep -q "timeline-month-marker\">Feb" public/index.html
grep -q "timeline-weekday\">thu" public/index.html
grep -q "blockquote::before" public/css/permanent.css
grep -q "text-transform: lowercase" public/css/permanent.css
grep -q -- "--spine-filter:" public/css/permanent.css
grep -q "p:has(> img:only-child)" public/css/permanent.css
grep -q "memory-image" public/css/permanent.css
grep -q "image_caption" content/journal/README.md
grep -q "width: min(68%, 380px)" public/css/permanent.css
grep -q "margin: 3rem auto 2.25rem" public/css/permanent.css
grep -q "aspect-ratio: 4 / 3" public/css/permanent.css
grep -q 'content: "no. " attr(data-archive-number)' public/css/permanent.css
grep -q "@media (max-width: 620px)" public/css/permanent.css
grep -q "grid-template-columns: 54px minmax(0, 1fr)" public/css/permanent.css
grep -q "width: calc(100% - 6px)" public/css/permanent.css

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

if find public/css -maxdepth 1 -type f ! -name permanent.css | grep -q .; then
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

echo "Validated $page_count site pages, navigation, assets, and RSS."

if grep -Eqi 'older snapshot|lifetime|hours unknown|>library</a>' public/games/index.html; then
  echo "Removed game UI leaked into the archive." >&2
  exit 1
fi
if [ -f public/library/tv/index.html ] || [ -f public/library/music/index.html ]; then
  echo "Unused placeholder collections were generated." >&2
  exit 1
fi
