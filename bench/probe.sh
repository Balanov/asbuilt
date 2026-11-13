#!/bin/bash
# Functional probe of each competitor tool: does it run, what does it emit, how big.
set -u
probe() {
  local label="$1"; shift
  local out rc
  out=$("$@" 2>/tmp/probe.err); rc=$?
  printf '=== %s ===\n' "$label"
  printf 'rc=%s bytes=%s\n' "$rc" "$(printf '%s' "$out" | wc -c)"
  printf 'stderr: %s\n' "$(head -c 300 /tmp/probe.err | tr '\n' ' ')"
  printf 'head: %s\n' "$(printf '%s' "$out" | head -3 | tr '\n' '|')"
  printf '\n'
}
probe "lshw -json"        lshw -json
probe "lshw -short"       lshw -short
probe "dmidecode"         dmidecode
probe "dmidecode -t 17"   dmidecode -t 17
probe "lscpu -J"          lscpu -J
probe "lspci -vvv"        lspci -vvv
probe "lspci"             lspci

echo "=== setpci LnkSta on each device ==="
for d in $(lspci | awk '{print $1}'); do
  echo "--- $d ---"
  setpci -s "$d" LnkSta 2>&1 | head -4
  setpci -s "$d" LnkCap 2>&1 | head -3
done

echo
echo "=== current_link_speed sysfs attrs on this WSL kernel ==="
ls /sys/bus/pci/devices/*/current_link_speed 2>&1 | head -3
ls /sys/bus/pci/devices/*/current_link_width 2>&1 | head -3
echo "sysfs pci device count: $(ls -d /sys/bus/pci/devices/*/ 2>/dev/null | wc -l)"

echo
echo "=== lshw / dmidecode without root ==="
echo "(current uid: $(id -u))"
echo "=== dmidecode /dev/mem ==="
ls -l /dev/mem 2>&1 | head -2

echo
echo "=== installed sizes (dpkg Installed-Size in KiB) ==="
for p in lshw dmidecode pciutils python3 python3-pip libpci3 libsysfs2; do
  v=$(dpkg-query -W -f='${Installed-Size}' "$p" 2>/dev/null)
  printf '%-14s %s KiB\n' "$p" "${v:-N/A}"
done
