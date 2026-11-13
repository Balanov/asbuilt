#!/bin/bash
# Confirm no files outside bench/ were modified during this session.
echo "=== files modified in the repo since 23:50 on 2026-11-13 ==="
find /mnt/f/GitHub/asbuilt -newermt '2026-11-13 23:50' -type f 2>/dev/null \
  | grep -v '^/mnt/f/GitHub/asbuilt/bench/' | head -20
echo "(end of list — empty means nothing outside bench/ was touched)"
echo
echo "=== everything modified in the repo since 23:50 ==="
find /mnt/f/GitHub/asbuilt -newermt '2026-11-13 23:50' -type f 2>/dev/null | sort
echo
echo "=== the two deliverables ==="
stat -c '%s bytes  %y  %n' /mnt/f/GitHub/asbuilt/bench/competitors.json /mnt/f/GitHub/asbuilt/bench/COMPETITORS.md
