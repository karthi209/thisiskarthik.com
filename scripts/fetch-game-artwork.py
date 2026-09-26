#!/usr/bin/env python3
"""Fetch local Steam portrait art for the game archive.

Matches are explicit: tracked AppIDs come from the archive and older manually
recorded games use the reviewed mapping below. Ambiguous titles are reported and
left untouched instead of receiving an unrelated image.
"""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import tempfile
import urllib.error
import urllib.request
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GAMES_PATH = ROOT / "content/games.json"
BACKLOG_PATH = ROOT / "content/games-backlog.json"
ART_DIR = ROOT / "content/images/games"

# Older completed records predate provider imports. These IDs were checked
# against their Steam store entries; absent/ambiguous games are deliberately
# omitted and listed after the run.
REVIEWED_APP_IDS = {
    "grand-theft-auto-vice-city": 12110,
    "grand-theft-auto-san-andreas": 12120,
    "max-payne": 12140,
    "grand-theft-auto-v": 271590,
    "call-of-duty-4-modern-warfare": 7940,
    "call-of-duty-modern-warfare-2": 10180,
    "call-of-duty-modern-warfare-3": 42680,
    "crysis": 17300,
    "crysis-2": 108800,
    "crysis-3": 1282690,
    "call-of-duty-black-ops": 42700,
    "call-of-duty-black-ops-2": 202970,
    "mass-effect": 17460,
    "mass-effect-2": 24980,
    "far-cry-4": 298110,
    "mass-effect-3": 1238020,
    "battlefield-3": 1238820,
    "firewatch": 383870,
    "elden-ring": 1245620,
    "ghost-of-tsushima": 2215430,
}

EXISTING_ART = {
    "cyberpunk-2077": "cyberpunk-cover.jpg",
    "detroit-become-human": "detroit-cover.jpg",
    "elden-ring": "elden-ring.jpg",
    "ghost-of-tsushima": "ghost-of-tsushima.jpg",
    "soma": "soma.jpg",
    "skyrim-special-edition": "skyrim-cover.jpg",
}

# Exact classic-PC matches reviewed separately because these titles do not have
# Steam portrait assets. LaunchBox links point to the first front-cover image on
# each matching Windows record; Midnight Racing uses its catalogued cover.
ARCHIVE_ART_URLS = {
    "icy-tower": "https://images.launchbox-app.com/c41c0548-d311-4f37-a859-b1b09fa90a7d.jpg",
    "glace": "https://images.launchbox-app.com/97444d6c-bdca-4bfc-9fd0-cd8173473f4a.png",
    "midnight-racing": "https://media.senscritique.com/media/000020073612/0/midnight_racing.jpg",
    "dirt-track-racing-2": "https://images.launchbox-app.com/e34886e0-8252-449f-933c-c662930bca00.jpg",
    "harry-potter-sorcerers-stone": "https://images.launchbox-app.com/r2_67bad026-bea4-45be-a35f-a3542524dc8f.jpg",
    "igi-2": "https://images.launchbox-app.com/r2_e9ca9d53-03fd-40cb-916f-183b13c4a567.png",
    "battlefield-2": "https://images.launchbox-app.com/d0a92a2c-abd2-4ea8-abfe-941d62fee94e.jpg",
    "iron-nest-heavy-turret-simulator": "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/2950790/e6c9e990560078611472e5dab98e79a405ed33f4/header.jpg",
    "ea-sports-fc-26": "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/3405690/2d96aa1b06e453cd62dae9029d412f19e61932c3/header.jpg",
}


def read_archive(path: Path):
    value = json.loads(path.read_text())
    return value, value["games"] if isinstance(value, dict) else value


def app_id(game: dict) -> int | None:
    tracked = game.get("tracking", {})
    return tracked.get("app_id") or REVIEWED_APP_IDS.get(game["id"])


def download(url: str, target: Path) -> bool:
    request = urllib.request.Request(url, headers={"User-Agent": "thisiskarthik.com artwork sync"})
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            body = response.read()
    except (urllib.error.HTTPError, urllib.error.URLError, TimeoutError):
        return False
    if not (body.startswith(b"\xff\xd8\xff") or body.startswith(b"\x89PNG")):
        return False
    target.write_bytes(body)
    return True


def write_json(path: Path, value) -> None:
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(json.dumps(value, indent=2, ensure_ascii=False) + "\n")
    temporary.replace(path)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--force", action="store_true", help="replace existing local artwork")
    args = parser.parse_args()
    if not shutil.which("magick"):
        parser.error("ImageMagick's magick command is required")

    documents = [read_archive(GAMES_PATH), read_archive(BACKLOG_PATH)]
    ART_DIR.mkdir(parents=True, exist_ok=True)
    fetched = reused = 0
    unresolved: list[str] = []
    sources: list[tuple[str, str]] = []

    with tempfile.TemporaryDirectory(prefix="game-art-") as scratch:
        scratch = Path(scratch)
        for _, games in documents:
            for game in games:
                existing = EXISTING_ART.get(game["id"])
                if existing and (ART_DIR / existing).is_file() and not args.force:
                    game["portrait"] = f"images/games/{existing}"
                    reused += 1
                    sources.append((game["title"], "existing local Steam promotional art"))
                    continue

                steam_id = app_id(game)
                archive_source = ARCHIVE_ART_URLS.get(game["id"])
                if not steam_id and not archive_source:
                    if not game.get("portrait") and not game.get("artifact"):
                        unresolved.append(game["title"])
                    continue

                filename = f"{game['id']}-portrait.webp"
                output = ART_DIR / filename
                source = (
                    f"https://cdn.akamai.steamstatic.com/steam/apps/{steam_id}/library_600x900.jpg"
                    if steam_id
                    else archive_source
                )
                if output.is_file() and not args.force:
                    game["portrait"] = f"images/games/{filename}"
                    reused += 1
                    sources.append((game["title"], archive_source or source))
                    continue

                original = scratch / f"{game['id']}.image"
                if steam_id:
                    candidates = [
                        f"https://cdn.akamai.steamstatic.com/steam/apps/{steam_id}/library_600x900_2x.jpg",
                        source,
                        f"https://cdn.akamai.steamstatic.com/steam/apps/{steam_id}/header.jpg",
                    ]
                    if archive_source:
                        candidates.append(archive_source)
                else:
                    candidates = [archive_source]
                used = next((url for url in candidates if download(url, original)), "")
                if not used:
                    unresolved.append(game["title"])
                    continue
                if used.endswith("/header.jpg"):
                    backdrop = scratch / f"{game['id']}-backdrop.png"
                    subprocess.run(
                        ["magick", str(original), "-auto-orient", "-resize", "300x450^", "-gravity", "center", "-extent", "300x450", "-blur", "0x18", str(backdrop)],
                        check=True,
                    )
                    subprocess.run(
                        ["magick", str(backdrop), "(", str(original), "-auto-orient", "-resize", "280x420", ")", "-gravity", "center", "-composite", "-quality", "82", str(output)],
                        check=True,
                    )
                else:
                    subprocess.run(
                        ["magick", str(original), "-auto-orient", "-resize", "300x450^", "-gravity", "center", "-extent", "300x450", "-quality", "82", str(output)],
                        check=True,
                    )
                game["portrait"] = f"images/games/{filename}"
                fetched += 1
                sources.append((game["title"], used))

    for path, (document, _) in zip((GAMES_PATH, BACKLOG_PATH), documents):
        write_json(path, document)

    source_lines = [
        "# Game artwork",
        "",
        "Local promotional artwork used only inside expanded game entries.",
        "Artwork belongs to the respective game publishers. No runtime hotlinks.",
        "",
    ]
    source_lines.extend(f"- {title}: {source}" for title, source in sorted(sources))
    (ART_DIR / "SOURCES.md").write_text("\n".join(source_lines) + "\n")

    print(f"{fetched} downloaded, {reused} reused, {len(unresolved)} unresolved")
    for title in unresolved:
        print(f"  unresolved: {title}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
