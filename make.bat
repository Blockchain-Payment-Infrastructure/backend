@echo off
REM Batch script for managing the project

:: Parse the first argument as the command
SET CMD=%1

IF "%CMD%"=="" (
    SET CMD=all
)

:: Build the application
IF /I "%CMD%"=="all" (
    CALL :build
    CALL :test
    GOTO :EOF
)

IF /I "%CMD%"=="build" (
    CALL :build
    GOTO :EOF
)

IF /I "%CMD%"=="docs" (
    CALL :docs
    GOTO :EOF
)

IF /I "%CMD%"=="clean-docs" (
    CALL :clean_docs
    GOTO :EOF
)

IF /I "%CMD%"=="run" (
    CALL :run
    GOTO :EOF
)

IF /I "%CMD%"=="docker-run" (
    CALL :docker_run
    GOTO :EOF
)

IF /I "%CMD%"=="docker-down" (
    CALL :docker_down
    GOTO :EOF
)

IF /I "%CMD%"=="test" (
    CALL :test
    GOTO :EOF
)

IF /I "%CMD%"=="itest" (
    CALL :itest
    GOTO :EOF
)

IF /I "%CMD%"=="clean" (
    CALL :clean
    GOTO :EOF
)

IF /I "%CMD%"=="watch" (
    CALL :watch
    GOTO :EOF
)

ECHO Unknown command: %CMD%
GOTO :EOF

:: ===============================
:: Commands
:: ===============================

:build
    ECHO Building...
    go build -o main cmd\api\main.go
    GOTO :EOF

:docs
    ECHO Building docs...
    swag init -g cmd\api\main.go
    GOTO :EOF

:clean_docs
    ECHO Cleaning old docs...
    IF EXIST docs\swagger.json del /F /Q docs\swagger.json
    IF EXIST docs\swagger.yaml del /F /Q docs\swagger.yaml
    GOTO :EOF

:run
    go run cmd\api\main.go
    GOTO :EOF

:docker_run
    docker compose up --build || docker-compose up --build
    GOTO :EOF

:docker_down
    docker compose down || docker-compose down
    GOTO :EOF

:test
    ECHO Testing...
    go test ./tests -v
    GOTO :EOF

:itest
    ECHO Running integration tests...
    go test ./internal/database -v
    GOTO :EOF

:clean
    ECHO Cleaning...
    IF EXIST main del /F /Q main
    GOTO :EOF

:watch
    powershell -Command "if (Get-Command air -ErrorAction SilentlyContinue) { exit 0 } else { exit 1 }"
    IF %ERRORLEVEL% EQU 0 (
        air
    ) ELSE (
        SET /P choice=Go's 'air' is not installed. Install now? [Y/n] 
        IF /I "%choice%"=="n" (
            ECHO You chose not to install air. Exiting...
            EXIT /B 1
        ) ELSE (
            go install github.com/air-verse/air@latest
            air
        )
    )
    GOTO :EOF

