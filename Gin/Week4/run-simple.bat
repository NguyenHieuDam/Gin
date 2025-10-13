@echo off
echo 🚀 Starting Simple News Aggregator (No Docker Required)
echo.

echo 📦 Building Go application...
go build -o bin/simple.exe ./cmd/simple

if %ERRORLEVEL% neq 0 (
    echo ❌ Build failed!
    pause
    exit /b 1
)

echo ✅ Build successful!
echo.
echo 🌐 Starting web server...
echo 📡 API: http://localhost:8080/api/v1/news
echo 🌐 Web: http://localhost:8080
echo.
echo Press Ctrl+C to stop the server
echo.

bin/simple.exe

pause
