#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
umask 077
printf 'OpenCMMC offline demonstration. Use synthetic evidence only, no CUI.\n'
binary=./opencmmc
if [ ! -x "$binary" ]; then binary=./freecompliance; fi
mkdir -p demo-evidence
if [ ! -e demo-workspace.json ]; then
  "$binary" cmmc init --level 2 --scope "Local demo - no CUI" --out demo-workspace.json
fi
"$binary" cmmc assess --input demo-workspace.json --evidence-dir demo-evidence --json demo-readiness.json
printf 'Report: demo-readiness.json (not a certification or SPRS score)\n'
