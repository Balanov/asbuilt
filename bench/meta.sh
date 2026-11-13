#!/bin/bash
# Collect license + size metadata from installed packages.
set -u
for p in lshw dmidecode pciutils; do
  echo "=== $p copyright (first 14 lines) ==="
  head -14 "/usr/share/doc/$p/copyright" 2>/dev/null
  echo
done
echo "=== binary sizes (bytes) ==="
stat -c '%s %n' /usr/bin/lshw /usr/sbin/dmidecode /usr/bin/lspci /usr/bin/setpci /usr/bin/lscpu /usr/bin/python3 2>&1
echo
echo "=== dpkg Installed-Size (KiB) and bytes ==="
for p in lshw dmidecode pciutils python3 libpci3; do
  k=$(dpkg-query -W -f='${Installed-Size}' "$p" 2>/dev/null)
  echo "$p ${k}KiB $(( ${k:-0} * 1024 )) bytes"
done
echo
echo "=== asbuilt binary size ==="
stat -c '%s %n' /tmp/asbuilt
