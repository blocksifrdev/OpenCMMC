#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
umask 077
printf 'OpenCMMC offline CMMC browser workbench.\n'
printf 'No cloud, telemetry, certification or SPRS submission.\n'
printf 'Keep this window running. Open the private local URL printed below.\n'
exec ./opencmmc cmmc ui --data-dir ./private-assessment --port 8765 --open-browser
