@echo off
setlocal
cd /d "%~dp0"
echo OpenCMMC offline demonstration. Use synthetic evidence only, no CUI.
if exist ".\opencmmc.exe" (
  set "CMMC_EXE=.\opencmmc.exe"
) else (
  set "CMMC_EXE=.\freecompliance.exe"
)
if not exist "demo-evidence" mkdir "demo-evidence"
if not exist "demo-workspace.json" (
  "%CMMC_EXE%" cmmc init --level 2 --scope "Local demo - no CUI" --out "demo-workspace.json"
  if errorlevel 1 goto failed
)
"%CMMC_EXE%" cmmc assess --input "demo-workspace.json" --evidence-dir "demo-evidence" --json "demo-readiness.json"
if errorlevel 1 goto failed
echo Report: demo-readiness.json (not a certification or SPRS score)
pause
exit /b 0
:failed
echo Demonstration failed. Review error above.
pause
exit /b 1
