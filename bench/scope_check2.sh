#!/bin/bash
# Check exact mtimes of site/ files vs my first bench/ write (23:55).
echo "=== my bench/ files (I wrote these) ==="
stat -c '%y  %n' /mnt/f/GitHub/asbuilt/bench/competitors.json /mnt/f/GitHub/asbuilt/bench/measure.py 2>/dev/null
echo
echo "=== site/ files NOT written by me ==="
stat -c '%y  %n' /mnt/f/GitHub/asbuilt/site/src/bench.js \
  /mnt/f/GitHub/asbuilt/site/src/data.js \
  /mnt/f/GitHub/asbuilt/site/package.json \
  /mnt/f/GitHub/asbuilt/site/index.html \
  /mnt/f/GitHub/asbuilt/assets/brand/logo.svg 2>/dev/null
echo
echo "=== count of repo files modified in the window 23:50-23:55 (before my first write) ==="
find /mnt/f/GitHub/asbuilt -newermt '2026-11-13 23:50' ! -newermt '2026-11-13 23:55' -type f 2>/dev/null | wc -l
echo "=== count of repo files modified in 23:55-00:03 (my window) ==="
find /mnt/f/GitHub/asbuilt -newermt '2026-11-13 23:55' ! -newermt '2026-11-13 00:03' -type f 2>/dev/null | wc -l
echo "=== non-bench files modified in MY window ==="
find /mnt/f/GitHub/asbuilt -newermt '2026-11-13 23:55' ! -newermt '2026-11-13 00:03' -type f 2>/dev/null | grep -v '/bench/' | head -20
echo "(end)"
