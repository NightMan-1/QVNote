#!/usr/bin/env python3
"""Finds orphaned files in QVNote note resources/ folders.

Walks the data directory (the one with *.qvnotebook folders), collects file
names referenced in each note's content.json (quiver-image-url/<file>,
quiver-file-url/<file>, /resources/<nb>/<note>/<file>) and compares them with
the actual files in the note's resources/ folder.

Nothing is deleted: the script only generates a platform-dependent cleanup
script (cleanup_orphaned_resources.sh on mac/linux, .bat on Windows) with
rm/del commands for every orphaned file. Review and run it manually.

Usage:
    python3 find_orphaned_resources.py <data_dir>

Standard library only, Python 3.x.
"""

import json
import os
import platform
import sys


def find_referenced_names(content_json_path):
    """Returns the set of resource file names referenced by a content.json.

    A file counts as referenced if its name appears anywhere in the raw file
    text — this intentionally avoids regexes: references exist in several
    formats (quiver-image-url/<f>, quiver-file-url/<f>, /resources/.../<f>)
    and random 32-char names make false positives practically impossible.
    """
    try:
        with open(content_json_path, encoding="utf-8") as f:
            raw = f.read()
    except OSError as e:
        print("  WARNING: cannot read {}: {}".format(content_json_path, e))
        return ""

    # actual note text: cells[].data, plus the raw file as a fallback
    names_text = raw
    try:
        data = json.loads(raw)
        cells_text = "".join(c.get("data", "") for c in data.get("cells", []))
        if cells_text:
            names_text = cells_text + "\n" + raw
    except ValueError:
        pass  # not valid json — raw text check still works
    return names_text


def main():
    if len(sys.argv) != 2:
        print(__doc__)
        sys.exit(1)
    data_dir = os.path.abspath(sys.argv[1])
    if not os.path.isdir(data_dir):
        print("Error: {} is not a directory".format(data_dir))
        sys.exit(1)

    orphaned = []  # full paths
    checked_files = 0
    notes_with_resources = 0

    for root, dirs, files in os.walk(data_dir):
        if os.path.basename(root) != "resources":
            continue
        note_dir = os.path.dirname(root)
        if not note_dir.endswith(".qvnote"):
            continue
        content_json = os.path.join(note_dir, "content.json")
        if not os.path.isfile(content_json):
            continue
        notes_with_resources += 1
        names_text = find_referenced_names(content_json)
        for name in files:
            checked_files += 1
            if name not in names_text:
                orphaned.append(os.path.join(root, name))

    # generate the cleanup script for the current platform
    is_windows = platform.system() == "Windows"
    out_name = "cleanup_orphaned_resources.bat" if is_windows else "cleanup_orphaned_resources.sh"
    out_path = os.path.join(os.getcwd(), out_name)
    with open(out_path, "w", encoding="utf-8", newline="\r\n" if is_windows else "\n") as f:
        if is_windows:
            f.write("@echo off\r\nrem {} orphaned resource files\r\n".format(len(orphaned)))
            for p in orphaned:
                f.write('del "{}"\r\n'.format(p))
        else:
            f.write("#!/bin/sh\n# {} orphaned resource files\nset -e\n".format(len(orphaned)))
            for p in orphaned:
                f.write("rm -- '{}'\n".format(p.replace("'", "'\\''")))
    if not is_windows:
        os.chmod(out_path, 0o755)

    total_size = sum(os.path.getsize(p) for p in orphaned)
    print("notes with resources : {}".format(notes_with_resources))
    print("files checked        : {}".format(checked_files))
    print("orphaned files       : {} ({:.1f} MB)".format(len(orphaned), total_size / 1024 / 1024))
    print("cleanup script       : {}".format(out_path))
    if orphaned:
        print("review it, then run it manually — nothing was deleted")


if __name__ == "__main__":
    main()
