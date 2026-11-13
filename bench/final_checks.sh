#!/bin/bash
# Final checks: asbuilt as non-root, lscpu/util-linux sizes, asbuilt license file.
set -u
echo "=== asbuilt as user nobody (fixture) ==="
su -s /bin/bash nobody -c '/tmp/asbuilt --root /tmp/server check >/tmp/nb.out 2>/tmp/nb.err; echo "rc=$?"' 2>&1
echo "stdout bytes: $(wc -c < /tmp/nb.out)"
echo "stderr: $(head -c 200 /tmp/nb.err)"
echo "--- findings line ---"
grep -E '^FINDINGS' /tmp/nb.out
echo
echo "=== asbuilt live check as nobody ==="
su -s /bin/bash nobody -c '/tmp/asbuilt check >/tmp/nb2.out 2>/tmp/nb2.err; echo "rc=$?"' 2>&1
echo "stdout bytes: $(wc -c < /tmp/nb2.out)"
echo "stderr: $(head -c 200 /tmp/nb2.err)"
echo
echo "=== util-linux / lscpu sizes ==="
dpkg-query -W -f='${Package} ${Installed-Size} KiB\n' util-linux 2>/dev/null
stat -c '%s %n' /usr/bin/lscpu
echo
echo "=== lscpu dimm/slot visibility ==="
lscpu -J | grep -icE 'dimm|slot|module'
echo "=== lshw memory subtree ==="
lshw -json 2>/dev/null | grep -icE '"description" : "System memory"'
echo "=== pcicrawler tree stdout (root) ==="
PYTHONPATH=/tmp/pcicrawler_env python3 -c "from pcicrawler.cli import main; main()" -t 2>/dev/null | head -5
echo "=== pcicrawler json stdout (root) ==="
PYTHONPATH=/tmp/pcicrawler_env python3 -c "from pcicrawler.cli import main; main()" -j 2>/dev/null | head -c 400
