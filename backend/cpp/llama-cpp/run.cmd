@echo off
setlocal enabledelayedexpansion

set CURDIR=%~dp0
pushd "%CURDIR%"

echo CPU info:
wmic cpu get name
wmic os get caption

set BINARY=llama-cpp-fallback.exe

if exist "%CURDIR%llama-cpp-cpu-all.exe" (
    set BINARY=llama-cpp-cpu-all.exe
)

if defined LLAMACPP_GRPC_SERVERS (
    if exist "%CURDIR%llama-cpp-grpc.exe" (
        set BINARY=llama-cpp-grpc.exe
    )
)

rem Add lib directory to PATH if it exists (mirrors LD_LIBRARY_PATH in run.sh)
if exist "%CURDIR%lib\" (
    set "PATH=%CURDIR%lib;%PATH%"
)

echo Using binary: %BINARY%

"%CURDIR%%BINARY%" %*
set EXITCODE=%ERRORLEVEL%

popd
exit /b %EXITCODE%