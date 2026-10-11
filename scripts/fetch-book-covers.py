#!/usr/bin/env python3
"""Fetch and optimize Open Library cover art for the reading log."""

from __future__ import annotations

import json
import shutil
import subprocess
import tempfile
import urllib.parse
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BOOKS_PATH = ROOT / "content/books.json"
ART_DIR = ROOT / "content/images/books"

COVER_IDS = {
    "sword-of-destiny": 6757701,
    "the-death-of-ivan-ilyich-and-other-stories": 7576980,
    "the-last-wish": 7360819,
}


def fetch(url: str) -> bytes:
    request = urllib.request.Request(url, headers={"User-Agent": "thisiskarthik.com cover sync"})
    with urllib.request.urlopen(request, timeout=30) as response:
        return response.read()


def main() -> int:
    if not shutil.which("magick"):
        raise SystemExit("ImageMagick's magick command is required")
    books = json.loads(BOOKS_PATH.read_text())["books"]
    ART_DIR.mkdir(parents=True, exist_ok=True)
    sources = ["# Book cover sources", "", "Cover images supplied by Open Library.", ""]
    with tempfile.TemporaryDirectory(prefix="book-covers-") as scratch_name:
        scratch = Path(scratch_name)
        for book in books:
            cover_id = COVER_IDS.get(book["id"])
            if cover_id is None:
                query = urllib.parse.urlencode({"title": book["title"], "author": book["author"], "fields": "cover_i", "limit": 10})
                result = json.loads(fetch(f"https://openlibrary.org/search.json?{query}"))
                match = next((doc for doc in result.get("docs", []) if doc.get("cover_i")), None)
                if not match:
                    print(f"No cover found: {book['title']} — {book['author']}")
                    continue
                cover_id = match["cover_i"]
            source_url = f"https://covers.openlibrary.org/b/id/{cover_id}-L.jpg"
            original = scratch / f"{book['id']}.jpg"
            original.write_bytes(fetch(source_url))
            target = ART_DIR / f"{book['id']}.webp"
            subprocess.run(["magick", str(original), "-auto-orient", "-resize", "420x630>", "-strip", "-quality", "82", str(target)], check=True)
            sources.append(f"- {book['title']} — {book['author']}: {source_url}")
            print(f"Fetched {book['title']}")
    (ART_DIR / "SOURCES.md").write_text("\n".join(sources) + "\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
