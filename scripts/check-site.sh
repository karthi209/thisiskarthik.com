#!/usr/bin/env bash
set -eu

if [ ! -f public/index.html ] || [ ! -f public/rss.xml ] || [ ! -f public/js/theme.js ] || [ ! -f public/css/permanent.css ] || [ ! -f public/spine-squiggle.svg ]; then
  echo "Generated site is missing required files." >&2
  exit 1
fi

page_count=0
while IFS= read -r page; do
  page_count=$((page_count + 1))
  grep -q 'js/theme.js' "$page"
done < <(find public -name index.html -type f -print)

if [ "$page_count" -lt 6 ]; then
  echo "Expected at least six generated HTML pages; found $page_count." >&2
  exit 1
fi

grep -q "karthik-theme" public/js/theme.js
grep -q "theme-toggle" public/index.html
grep -q "theme-mark" public/index.html
grep -q "☙" public/index.html
grep -q "❧" public/js/theme.js
grep -q "animation: none" public/css/permanent.css
grep -q "class=\"timeline\"" public/index.html
grep -q "memories, before they become stories" public/index.html
grep -q "started in 2026, still writing" public/index.html
grep -q "for later, when i forget" public/index.html
grep -q "masthead-byline\">by karthik" public/index.html
grep -q "class=\"timeline essay-timeline\"" public/essays/index.html
grep -q 'data-archive-number="003"' public/index.html
grep -q 'data-archive-number="001"' public/index.html
grep -q "timeline-month-marker\">Jul" public/index.html
grep -q "timeline-month-marker\">Feb" public/index.html
grep -q "timeline-weekday\">thu" public/index.html
grep -q "blockquote::before" public/css/permanent.css
grep -q "text-transform: lowercase" public/css/permanent.css
grep -q "html\\[data-theme=\"dark\"\\]" public/css/permanent.css
grep -q -- "--page-bg: #161512" public/css/permanent.css
grep -q -- "--paper: #161512" public/css/permanent.css
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

echo "Validated $page_count journal pages, navigation, theme assets, and RSS."
