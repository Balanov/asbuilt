#!/bin/bash
# Final verification: fixture attributes + deliverable consistency.
set -u
echo "=== fixture PCIe attributes (source of the asbuilt link verdict) ==="
find /tmp/server -name 'current_link_speed' -o -name 'max_link_speed' -o -name 'current_link_width' 2>/dev/null | head -10
echo "--- sample values ---"
for f in $(find /tmp/server -name 'current_link_speed' -o -name 'max_link_speed' 2>/dev/null | head -4); do
  printf '%s = %s\n' "$f" "$(cat "$f" 2>/dev/null)"
done
echo
echo "=== fixture top-level tree ==="
find /tmp/server -maxdepth 3 -type d 2>/dev/null | head -12
echo
echo "=== deliverables present ==="
ls -l /mnt/f/GitHub/asbuilt/bench/
echo
echo "=== JSON still valid ==="
python3 /mnt/f/GitHub/asbuilt/bench/validate.py
echo
echo "=== cross-check: JSON means vs markdown table ==="
python3 - <<'PYEOF'
import json, re
doc = json.load(open('/mnt/f/GitHub/asbuilt/bench/competitors.json'))
md = open('/mnt/f/GitHub/asbuilt/bench/COMPETITORS.md', encoding='utf-8').read()
# strip thousands separators the md uses for readability
flat = md.replace(' ', '').replace('\u00a0','')
for name in ('asbuilt','lshw','lscpu','lspci','dmidecode','pcicrawler','perfspect','sat-hwmatch'):
    e = doc['tool'][name]
    if e['runtime_us_mean'] is None:
        print(f"{name:12s} null (not measurable) -> md shows em-dash: {'—' in md}")
        continue
    m = e['runtime_us_mean']
    s = f"{m:,}".replace(',', '')
    plain = str(m)
    ok = (plain in flat) or (s in flat)
    print(f"{name:12s} mean={m:<8} present_in_md={ok}")
PYEOF
