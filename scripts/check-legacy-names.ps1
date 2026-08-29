[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$allowlistPath = Join-Path $repoRoot 'tools/legacy-name-allowlist.txt'
$targets = @(
    'Dockerfile',
    'deploy/.env.example',
    'deploy/Dockerfile',
    'deploy/docker-entrypoint.sh',
    'deploy/apple-container.sh',
    'deploy/docker-compose.dev.yml',
    'deploy/docker-compose.local.yml',
    'deploy/docker-compose.standalone.yml',
    'deploy/docker-compose.yml',
    'deploy/easysub2api.service',
    'deploy/easysub2api-datamanagementd.service',
    'backend/internal/config/config.go',
    'backend/internal/handler/gateway_key_billing.go',
    'backend/internal/server/routes/gateway.go',
    'backend/internal/server/middleware/api_key_auth.go',
    'backend/internal/handler/admin/ops_ws_handler.go',
    'backend/internal/service/gateway_service.go',
    'backend/internal/service/upstream_billing_probe.go',
    'backend/internal/securityaudit/prompt_types.go',
    'backend/internal/service/data_management_service.go',
    'backend/internal/service/update_service.go',
    'backend/internal/handler/admin/account_data.go',
    'frontend/src/main.ts',
    'frontend/src/utils/migrateLegacyStorage.ts',
    'frontend/src/utils/ipGeoLookup.ts',
    'frontend/src/types/index.ts',
    'frontend/src/api/admin/ops.ts',
    'frontend/src/components/keys/UseKeyModal.vue',
    'frontend/src/components/admin/account/ImportDataModal.vue'
)

$allowedPaths = @{}
foreach ($line in Get-Content $allowlistPath) {
    $trimmed = $line.Trim()
    if (-not $trimmed -or $trimmed.StartsWith('#')) { continue }
    $fields = $trimmed.Split('|', 3)
    if ($fields.Count -ne 3 -or $fields[1] -ne 'next-release') {
        throw "Invalid legacy allowlist entry: $line"
    }
    $allowedPaths[$fields[0].Replace('\\', '/')] = $fields[2]
}

$violations = [Collections.Generic.List[string]]::new()
foreach ($relativePath in $targets) {
    $absolutePath = Join-Path $repoRoot $relativePath
    if (-not (Test-Path -LiteralPath $absolutePath)) {
        $violations.Add("Missing audited file: $relativePath")
        continue
    }

    $lineNumber = 0
    foreach ($line in Get-Content -LiteralPath $absolutePath) {
        $lineNumber += 1
        if ($line -cmatch '(?i)(?<!easy)sub2api') {
            $normalizedPath = $relativePath.Replace('\\', '/')
            if (-not $allowedPaths.ContainsKey($normalizedPath)) {
                $violations.Add("${normalizedPath}:${lineNumber}: $line")
            }
        }
    }
}

if ($violations.Count -gt 0) {
    Write-Error ("Unregistered legacy EasySub2api names:`n" + ($violations -join "`n"))
    exit 1
}

Write-Host "Legacy-name audit passed. $($allowedPaths.Count) compatibility files expire next release."
