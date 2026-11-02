@echo off
setlocal
cd /d "%~dp0"
echo FreeCompliance offline CMMC browser workbench.
echo No cloud, subscription, telemetry, CMMC certification or SPRS submission.
echo Keep this window running. Open the private local URL printed below.
".\freecompliance.exe" cmmc ui --data-dir ".\private-assessment" --port 8765 --open-browser
if errorlevel 1 (
  echo UI could not start. Check if another program is using port 8765.
  pause
  exit /b 1
)
