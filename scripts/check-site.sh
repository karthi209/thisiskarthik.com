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
grep -q "animation: none" public/css/permanent.css
grep -q "class=\"timeline\"" public/index.html
grep -q "life, more or less in the order it happened" public/index.html
grep -q ">entries<" public/index.html
grep -q "for later, when i forget" public/index.html
grep -q "masthead-byline\">by karthik" public/index.html
grep -q "class=\"timeline essay-timeline\"" public/writings/index.html
grep -q "timeline-month-marker\">Jul" public/index.html
grep -q "timeline-month-marker\">Feb" public/index.html
grep -q "timeline-weekday\">thu" public/index.html
grep -q "blockquote::before" public/css/permanent.css
grep -q "text-transform: lowercase" public/css/permanent.css
grep -q "@media (max-width: 620px)" public/css/permanent.css
grep -q "grid-template-columns: 68px minmax(0, 1fr)" public/css/permanent.css

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
