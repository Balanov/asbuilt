#!/bin/bash
# Probe 2: setpci register names, non-root behavior, dmidecode exit codes.
set -u
echo "=== setpci --help (register name list, PCIe part) ==="
setpci --help 2>&1 | grep -iE 'lnk|extended|config' | head -20
echo
echo "=== setpci -v on first device, raw extended config ==="
setpci -v -s 1b77:00:00.0 0x12 2>&1 | head -5
echo "--- setpci with '--dpc' / device listing ---"
setpci -s 1b77:00:00.0 LNKSTA 2>&1 | head -3
echo
echo "=== lspci -vvv capabilities section (root) ==="
lspci -vvv -s 1b77:00:00.0 2>&1 | grep -iA2 'capabilities' | head -10
echo
echo "=== NON-ROOT tests (user nobody) ==="
echo "--- lshw -short as nobody ---"
su -s /bin/bash nobody -c 'lshw -short' 2>&1 | head -4
echo "--- lshw -json as nobody rc ---"
su -s /bin/bash nobody -c 'lshw -json >/dev/null 2>/tmp/lshw_noroot.err; echo rc=$?' 2>&1 | head -2
cat /tmp/lshw_noroot.err 2>/dev/null | head -3
echo "--- dmidecode as nobody ---"
su -s /bin/bash nobody -c 'dmidecode' 2>&1 | head -4
echo "--- dmidecode as nobody rc ---"
su -s /bin/bash nobody -c 'dmidecode >/dev/null 2>&1; echo rc=$?' 2>&1 | head -2
echo "--- lspci -vvv as nobody ---"
su -s /bin/bash nobody -c 'lspci -vvv -s 1b77:00:00.0' 2>&1 | grep -iE 'denied|capabilit|LnkSta' | head -5
echo
echo "=== dmidecode exit code as ROOT ==="
dmidecode >/dev/null 2>&1; echo "root dmidecode rc=$?"
dmidecode -t 17 >/dev/null 2>&1; echo "root dmidecode -t 17 rc=$?"
lshw -json >/dev/null 2>&1; echo "root lshw -json rc=$?"
echo
echo "=== does lscpu see DIMMs? (memory population) ==="
lscpu -J 2>/dev/null | grep -icE 'dimm|slot|module' 
echo "=== lshw memory class ==="
lshw -json 2>/dev/null | grep -iA4 '"class" : "memory"' | head -20
