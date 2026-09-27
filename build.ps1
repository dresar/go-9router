# Go 9Router - One-Click Build Script
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "   Building Go 9Router (Unified Build)   " -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan

# 1. Build Go Binaries
Write-Host "`n[1/2] Compiling Go Binaries (9router.exe & gateway.exe)..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o bin/9router.exe ./cmd/9router
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to compile 9router.exe" -ForegroundColor Red
    exit 1
}

go build -ldflags="-s -w" -o bin/gateway.exe ./cmd/gateway
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to compile gateway.exe" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Successfully compiled bin/9router.exe and bin/gateway.exe!" -ForegroundColor Green

# 2. Status
$binSize = (Get-Item "bin/9router.exe").Length / 1MB
Write-Host "`n[2/2] Build Complete!" -ForegroundColor Cyan
Write-Host "   Binary size: $('{0:N2}' -f $binSize) MB" -ForegroundColor White
Write-Host "   Binary path: bin/9router.exe" -ForegroundColor White
Write-Host "`nTo run the server & CLI:" -ForegroundColor Yellow
Write-Host "   .\bin\9router.exe                  # Starts server and opens browser" -ForegroundColor White
Write-Host "   .\bin\9router.exe status           # Checks gateway status" -ForegroundColor White
Write-Host "   .\bin\9router.exe menu             # Interactive terminal menu" -ForegroundColor White
Write-Host "   .\bin\9router.exe xai video --help # Generate Grok Imagine videos" -ForegroundColor White
Write-Host "`nTo run the Next.js Frontend Dashboard:" -ForegroundColor Yellow
Write-Host "   cd frontend; bun run dev           # Runs dev server on port 20127" -ForegroundColor White
