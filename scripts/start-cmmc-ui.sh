#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
umask 077
printf 'FreeCompliance offline CMMC browser workbench.\n'
printf 'No cloud, telemetry, certification or SPRS submission.\n'
printf 'Keep this window running. Open the private local URL printed below.\n'
exec ./freecompliance cmmc ui --data-dir ./private-assessment --port 8765
