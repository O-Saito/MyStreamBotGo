@echo off
setlocal enabledelayedexpansion

if "%~1"=="" (
    echo Usage: %~nx0 ^<version^>
    echo Example: %~nx0 1.0.0
    exit /b 1
)

set VERSION=%1
set BUILD_DIR=build\v%VERSION%
set BINARY_NAME=mystreambot.exe

echo ^==^> Building MyStreamBot v%VERSION%...

if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
mkdir "%BUILD_DIR%"

for /f "tokens=*" %%a in ('wmic os get localdatetime ^| find "."') do set DTS=%%a
set BUILD_DATE=!DTS:~0,4!-!DTS:~4,2!-!DTS:~6,2!T!DTS:~8,2!:!DTS:~10,2!:!DTS:~12,2!Z
for /f "tokens=*" %%h in ('git rev-parse --short HEAD 2^>nul') do set COMMIT_HASH=%%h
if not defined COMMIT_HASH set COMMIT_HASH=none

go build -ldflags "-X main.Version=%VERSION% -X main.BuildDate=%BUILD_DATE% -X main.CommitHash=%COMMIT_HASH%" -o "%BUILD_DIR%\%BINARY_NAME%" .

echo ^==^> Copying files...

::copy init.txt            "%BUILD_DIR%\" >nul
copy twitchsubtypes.json "%BUILD_DIR%\" >nul
xcopy /e /i /y definitions "%BUILD_DIR%\definitions\" >nul
xcopy /e /i /y modules     "%BUILD_DIR%\modules\" >nul
xcopy /e /i /y web         "%BUILD_DIR%\web\" >nul
::xcopy /e /i /y db          "%BUILD_DIR%\db\" >nul
::xcopy /e /i /y logs        "%BUILD_DIR%\logs\" >nul

echo Done ^→^ %BUILD_DIR%
