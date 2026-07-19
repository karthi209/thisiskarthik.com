#!/usr/bin/env bash
set -eu

if [ ! -f public/index.html ] || [ ! -f public/rss.xml ] || [ ! -f public/js/theme.js ] || [ ! -f public/js/home.js ]; then
  echo "Generated site is missing required files." >&2
  exit 1
fi

page_count=0
while IFS= read -r page; do
  page_count=$((page_count + 1))
  grep -q 'class="site-switcher"' "$page"
  grep -q 'https://library.thisiskarthik.com/' "$page"
  grep -q 'js/theme.js' "$page"
done < <(find public -name index.html -type f -print)

if [ "$page_count" -lt 6 ]; then
  echo "Expected at least six generated HTML pages; found $page_count." >&2
  exit 1
fi

grep -q "karthik-theme" public/js/theme.js
grep -q "vadivelu" public/js/home.js
grep -q "Enlighten Me Before Sleep" public/index.html

if grep -Eqi 'placeholder game|neocities mode' public/index.html; then
  echo "Placeholder homepage content leaked into the generated site." >&2
  exit 1
fi

echo "Validated $page_count Journal pages, reciprocal navigation, theme assets, and RSS."
