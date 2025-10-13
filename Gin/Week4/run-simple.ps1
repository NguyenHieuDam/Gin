# Simple News Aggregator Runner (No Docker Required)
Write-Host "🚀 Starting Simple News Aggregator (No Docker Required)" -ForegroundColor Green
Write-Host ""

Write-Host "📦 Building Go application..." -ForegroundColor Yellow
go build -o bin/simple.exe ./cmd/simple

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""
Write-Host "🌐 Starting web server..." -ForegroundColor Cyan
Write-Host "📡 API: http://localhost:8080/api/v1/news" -ForegroundColor Cyan
Write-Host "🌐 Web: http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""

./bin/simple.exe
