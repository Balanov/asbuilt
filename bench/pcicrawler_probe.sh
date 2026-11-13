#!/bin/bash
# Install pcicrawler into an ISOLATED target dir (system site-packages untouched),
# measure its real installed footprint, then try running it.
set -u
rm -rf /tmp/pcicrawler_env
pip install --quiet --target /tmp/pcicrawler_env pcicrawler 2>&1 | tail -3
echo "=== installed footprint (isolated target dir) ==="
du -sb /tmp/pcicrawler_env 2>/dev/null
echo "--- per-package bytes ---"
du -sb /tmp/pcicrawler_env/* 2>/dev/null | sort -rn | head -12
echo
echo "=== dependency set actually resolved ==="
ls /tmp/pcicrawler_env | tr '\n' ' '
echo
echo "=== try running pcicrawler (root, this WSL machine) ==="
PYTHONPATH=/tmp/pcicrawler_env python3 -c "from pcicrawler.cli import main; main()" --help 2>&1 | head -6
echo "--- tree view ---"
PYTHONPATH=/tmp/pcicrawler_env python3 -c "from pcicrawler.cli import main; main()" -t 2>&1 | head -15
echo "--- json view (bytes) ---"
PYTHONPATH=/tmp/pcicrawler_env python3 -c "from pcicrawler.cli import main; main()" -j 2>/dev/null | wc -c
echo "--- rc ---"
PYTHONPATH=/tmp/pcicrawler_env python3 -c "from pcicrawler.cli import main; main()" -t >/dev/null 2>&1; echo "rc=$?"
