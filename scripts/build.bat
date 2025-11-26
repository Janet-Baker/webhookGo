@echo off
REM 构建脚本 - 用于本地测试构建过程
REM 用法: build.bat

setlocal enabledelayedexpansion

echo 🚀 开始构建 webhookGo...

REM 创建输出目录
if not exist dist mkdir dist

REM 构建配置
set VERSION=%1
if "%VERSION%"=="" set VERSION=dev
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set BUILD_TIME=%dt:~0,4%-%dt:~4,2%-%dt:~6,2%_%dt:~8,2%:%dt:~10,2%:%dt:~12,2%
set GIT_COMMIT=unknown

REM 尝试获取Git提交ID
for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i

REM 构建标志
set LDFLAGS=-s -w -X main.version=%VERSION% -X main.buildTime=%BUILD_TIME% -X main.gitCommit=%GIT_COMMIT%

echo 📦 版本: %VERSION%
echo ⏰ 构建时间: %BUILD_TIME%
echo 🔧 Git提交: %GIT_COMMIT%
echo.

echo 📥 检查和更新依赖...
go mod tidy
go mod verify

echo.
REM 开始构建
call :build linux amd64 "" webhookGo-linux-amd64
call :build linux arm64 "" webhookGo-linux-arm64
call :build linux 386 "" webhookGo-linux-386
call :build linux arm 7 webhookGo-linux-armv7
call :build windows amd64 ".exe" webhookGo-windows-amd64
call :build windows 386 ".exe" webhookGo-windows-386
call :build darwin amd64 "" webhookGo-darwin-amd64
call :build darwin arm64 "" webhookGo-darwin-arm64
call :build freebsd amd64 "" webhookGo-freebsd-amd64

echo.
echo ✅ 构建完成！
echo 📁 输出目录: dist\
echo 📋 文件列表:
dir dist\

goto :eof

:build
set GOOS=%1
set GOARCH=%2
set EXT=%3
set OUTPUT=%4

echo 🔨 构建 %OUTPUT%%EXT%...

set CGO_ENABLED=0
set GOOS=%GOOS%
set GOARCH=%GOARCH%

go build -ldflags="%LDFLAGS%" -o "dist\%OUTPUT%%EXT%" .

REM 创建压缩包
cd dist
if "%GOOS%"=="windows" (
    powershell -command "Compress-Archive -Path '%OUTPUT%%EXT%' -DestinationPath '%OUTPUT%.zip' -Force"
    del "%OUTPUT%%EXT%"
) else (
    tar -czf "%OUTPUT%.tar.gz" "%OUTPUT%%EXT%"
    del "%OUTPUT%%EXT%"
)
cd ..

goto :eof