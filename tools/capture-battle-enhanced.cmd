@echo off
setlocal
cd /d "%~dp0"
if not exist "%~dp0gin3.exe" (
  echo Copy this CMD next to gin3.exe, then run it again.
  pause
  exit /b 2
)
set "LOGH3_BATTLE_LOG=%TEMP%\logh3-battle-enhanced-%RANDOM%-%RANDOM%.log"
>"%LOGH3_BATTLE_LOG%" echo LOGH3 battle diagnostic - enhanced AI
>>"%LOGH3_BATTLE_LOG%" "%~dp0gin3.exe" -version
>>"%LOGH3_BATTLE_LOG%" 2>&1 "%~dp0gin3.exe" -ai 1 -music-quality hd %*
set "LOGH3_BATTLE_EXIT=%ERRORLEVEL%"
>>"%LOGH3_BATTLE_LOG%" echo ExitCode=%LOGH3_BATTLE_EXIT%
echo Log: %LOGH3_BATTLE_LOG%
if not "%LOGH3_BATTLE_EXIT%"=="0" (
  type "%LOGH3_BATTLE_LOG%"
  pause
)
exit /b %LOGH3_BATTLE_EXIT%
