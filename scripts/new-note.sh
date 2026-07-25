#!/usr/bin/env bash
set -eu

entry_date="${JOURNAL_DATE:-$(date +%F)}"
entry_time="$(date +%H:%M)"
entry_year="${entry_date%%-*}"
date_remainder="${entry_date#*-}"
entry_month="${date_remainder%%-*}"
entry_dir="content/journal/${entry_year}/${entry_month}"
entry_file="${entry_dir}/${entry_date}-$(date +%H%M%S).md"

mkdir -p "$entry_dir"

write_note() {
  note_body="$1"
  note_title="${2:-}"
  printf '%s\n' \
    '---' \
    "title: ${note_title}" \
    "date: ${entry_date}" \
    "time: ${entry_time}" \
    '---' \
    '' \
    "$note_body" > "$entry_file"
}

if [ "$#" -gt 1 ]; then
  note_title="$1"
  shift
  write_note "$*" "$note_title"
  printf 'Added %s\n' "$entry_file"
  exit 0
fi

if [ "$#" -gt 0 ]; then
  write_note "$*" ""
  printf 'Added %s\n' "$entry_file"
  exit 0
fi

write_note "Write here." "A small thing"
"${EDITOR:-vi}" "$entry_file"
