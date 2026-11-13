#!/bin/bash
# Diagnostic: verify asbuilt output mode / --root behavior and tool availability.
set -u
A=/tmp/asbuilt
echo "=== md5 + bytes for each invocation ==="
run() {
  local label="$1"; shift
  local out
  out=$("$@" 2>&1)
  printf '%-42s bytes=%-8s md5=%s\n' "$label" "$(printf '%s' "$out" | wc -c)" "$(printf '%s' "$out" | md5sum | cut -c1-12)"
}
run "check plain"            $A --root /tmp/server check
run "check --json (post-sub)" $A --root /tmp/server check --json
run "check --json (pre-sub)"  $A --json --root /tmp/server check
run "manifest plain"          $A --root /tmp/server manifest
run "manifest --json (post)"  $A --root /tmp/server manifest --json
run "manifest --json (pre)"   $A --json --root /tmp/server manifest

echo
echo "=== first 3 lines of each (text vs json detection) ==="
for inv in "--root /tmp/server check" "--root /tmp/server check --json" "--json --root /tmp/server check" "--root /tmp/server manifest" "--root /tmp/server manifest --json" "--json --root /tmp/server manifest"; do
  printf '%-42s => ' "$inv"
  $A $inv 2>&1 | head -1
done

echo
echo "=== does --json manifest on fixture report fixture CPU (Xeon) or live CPU (AMD)? ==="
$A --json --root /tmp/server manifest 2>&1 | grep -iE '"model"|"vendor"|"logical|hostname' | head -8
echo "--- live machine for comparison ---"
$A --json manifest 2>&1 | grep -iE '"model"|"vendor"|hostname' | head -8

echo
echo "=== tool availability ==="
for t in python3 pip3 pip lscpu lspci setpci dmidecode lshw hyperfine; do
  if command -v "$t" >/dev/null 2>&1; then
    printf '%-12s %s\n' "$t" "$(command -v $t)"
  else
    printf '%-12s MISSING\n' "$t"
  fi
done
