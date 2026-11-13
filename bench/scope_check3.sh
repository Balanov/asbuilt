#!/bin/bash
# Correct scope check with proper upper bound (2026-11-14 00:03).
echo "=== files modified 23:54 (session start) -> 00:03, grouped by top dir ==="
find /mnt/f/GitHub/asbuilt -newermt '2026-11-13 23:54' ! -newermt '2026-11-14 00:03' -type f 2>/dev/null \
  | sed 's|^/mnt/f/GitHub/asbuilt/||' | cut -d/ -f1-2 | sort | uniq -c | sort -rn | head -15
echo
echo "=== newest 8 files OUTSIDE bench/ (should be another session's work, not mine) ==="
find /mnt/f/GitHub/asbuilt -newermt '2026-11-13 23:54' ! -newermt '2026-11-14 00:03' -type f 2>/dev/null \
  | grep -v '/bench/' | xargs -r stat -c '%y %n' 2>/dev/null | sort -r | head -8
echo
echo "=== my own writes (bench/ only) ==="
find /mnt/f/GitHub/asbuilt/bench -newermt '2026-11-13 23:54' -type f 2>/dev/null | wc -l
echo "files in bench/ written by me"
echo
echo "=== confirm asbuilt sources/bin untouched by me ==="
stat -c '%y %n' /mnt/f/GitHub/asbuilt/bin/asbuilt-linux-amd64 /mnt/f/GitHub/asbuilt/LICENSE 2>/dev/null
find /mnt/f/GitHub/asbuilt/tests/fixtures/server -newermt '2026-11-13 23:54' -type f 2>/dev/null | wc -l
echo "fixture files modified since session start (must be 0)"
