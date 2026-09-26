"""Read only the observed Lutris 0.5.22 schema; never migrate or edit its database."""
import json
import pathlib
import sqlite3
import sys

path = pathlib.Path(sys.argv[1]).resolve(strict=True)
with sqlite3.connect(path.as_uri() + "?mode=ro", uri=True, timeout=3) as db:
    db.execute("PRAGMA query_only = ON")
    db.execute("BEGIN")
    columns = {row[1] for row in db.execute("PRAGMA table_info(games)")}
    required = {"id", "name", "slug", "playtime", "lastplayed"}
    if not required <= columns:
        raise ValueError("Unsupported Lutris schema")
    db.row_factory = sqlite3.Row
    rows = db.execute("SELECT id, name, slug, playtime, lastplayed FROM games ORDER BY id")
    print(json.dumps([dict(row) for row in rows], allow_nan=False))
