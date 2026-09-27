$ErrorActionPreference = "Stop"
try {
    $loginRes = Invoke-WebRequest -Uri "http://127.0.0.1:20128/api/auth/login" -Method POST -Body '{"password":"admin1234"}' -ContentType "application/json" -SessionVariable session
    Write-Host "1. Login Status: $($loginRes.StatusCode)"

    $clientRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/providers/client?sort=provider" -WebSession $session
    Write-Host "2. Provider Client: $($clientRes.connections.Length) connections, $($clientRes.providerOptions.Length) options"

    $chartRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/usage/chart?period=7d" -WebSession $session
    Write-Host "3. Usage Chart Buckets: $($chartRes.Length)"

    $reqLogsRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/usage/request-logs" -WebSession $session
    Write-Host "4. Request Logs Count: $($reqLogsRes.Length)"

    $providersRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/usage/providers" -WebSession $session
    Write-Host "5. Usage Providers Count: $($providersRes.providers.Length)"

    $reqLoginRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/settings/require-login"
    Write-Host "6. Require Login: requireLogin=$($reqLoginRes.requireLogin), tunnelDashboardAccess=$($reqLoginRes.tunnelDashboardAccess)"

    $tunnelRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/tunnel/status" -WebSession $session
    Write-Host "7. Tunnel Status: tunnel=$($tunnelRes.tunnel.status), tailscale=$($tunnelRes.tailscale.status)"

    $cliToolsRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/cli-tools/all-statuses" -WebSession $session
    Write-Host "8. CLI Tools all-statuses: claude installed=$($cliToolsRes.claude.installed)"

    $presetsRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/combos/presets?source=cursor" -WebSession $session
    Write-Host "9. Combos presets: source=$($presetsRes.source), items=$($presetsRes.items.Length)"

    $availRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/models/availability" -WebSession $session
    Write-Host "10. Models availability: unavailableCount=$($availRes.unavailableCount)"

    $disabledRes = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/models/disabled" -WebSession $session
    Write-Host "11. Disabled models: OK"

    $cursorOauth = Invoke-RestMethod -Uri "http://127.0.0.1:20128/api/oauth/cursor/auto-import" -WebSession $session
    Write-Host "12. Cursor Auto-Import: found=$($cursorOauth.found)"

    Write-Host "`nALL 12 FRONTEND-TO-GO ENDPOINTS VERIFIED SUCCESSFULLY!"
} catch {
    Write-Host "TEST FAILED: $_"
    exit 1
}
