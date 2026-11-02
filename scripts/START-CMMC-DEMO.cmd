@echo off
setlocal
cd /d "%~dp0"
echo FreeCompliance offline demonstration. Use synthetic evidence only, no CUI.
if not exist "demo-evidence" mkdir "demo-evidence"
if not exist "demo-workspace.json" (
  ".\freecompliance.exe" cmmc init --level 2 --scope "Local demo - no CUI" --out "demo-workspace.json"
  if errorlevel 1 goto failed
)
".\freecompliance.exe" cmmc assess --input "demo-workspace.json" --evidence-dir "demo-evidence" --json "demo-readiness.json"
if errorlevel 1 goto failed
echo Report: demo-readiness.json (not a certification or SPRS score)
pause
exit /b 0
:failed
echo Demonstration failed. Review error above.
pause
exit /b 1
