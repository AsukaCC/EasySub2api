[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^[0-9]+\.[0-9]+\.[0-9]+$')]
    [string]$Version,

    [string]$NotesFile
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$versionPath = Join-Path $repoRoot 'backend/cmd/server/VERSION'
$releaseTag = "v$Version"

function Invoke-Checked {
    param(
        [Parameter(Mandatory = $true)][string]$Command,
        [Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments
    )

    $commandOutput = & $Command @Arguments 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "$Command $($Arguments -join ' ') failed:`n$($commandOutput -join [Environment]::NewLine)"
    }
    return ($commandOutput -join [Environment]::NewLine).Trim()
}

function Get-LatestWorkflowRuns {
    param([Parameter(Mandatory = $true)][string]$Commit)

    $json = Invoke-Checked -Command gh -Arguments @(
        'run', 'list', '--commit', $Commit, '--event', 'push', '--limit', '50',
        '--json', 'databaseId,workflowName,status,conclusion,url'
    )
    if ([string]::IsNullOrWhiteSpace($json)) {
        return @()
    }
    return @($json | ConvertFrom-Json)
}

function Wait-RequiredChecks {
    param([Parameter(Mandatory = $true)][string]$Commit)

    $requiredWorkflows = @('CI', 'Security Scan')
    $deadline = (Get-Date).AddMinutes(60)

    Write-Host "Waiting for GitHub checks on $Commit ..."
    while ((Get-Date) -lt $deadline) {
        $runs = Get-LatestWorkflowRuns -Commit $Commit
        $allComplete = $true

        foreach ($workflowName in $requiredWorkflows) {
            $run = $runs |
                Where-Object { $_.workflowName -eq $workflowName } |
                Sort-Object databaseId -Descending |
                Select-Object -First 1

            if ($null -eq $run) {
                $allComplete = $false
                continue
            }
            if ($run.status -ne 'completed') {
                $allComplete = $false
                continue
            }
            if ($run.conclusion -ne 'success') {
                throw "$workflowName failed for $Commit. Inspect $($run.url)"
            }
        }

        if ($allComplete) {
            Write-Host 'CI and security checks passed.' -ForegroundColor Green
            return
        }
        Start-Sleep -Seconds 10
    }

    throw "Timed out waiting for GitHub checks on $Commit."
}

function Get-ReleaseRun {
    $json = Invoke-Checked -Command gh -Arguments @(
        'run', 'list', '--workflow', 'release.yml', '--event', 'workflow_dispatch', '--limit', '30',
        '--json', 'databaseId,displayTitle,status,conclusion,url'
    )
    if ([string]::IsNullOrWhiteSpace($json)) {
        return $null
    }
    return @($json | ConvertFrom-Json) |
        Where-Object { $_.displayTitle -eq "Release $releaseTag" } |
        Sort-Object databaseId -Descending |
        Select-Object -First 1
}

Push-Location $repoRoot
try {
    foreach ($tool in @('git', 'gh')) {
        if ($null -eq (Get-Command $tool -ErrorAction SilentlyContinue)) {
            if ($tool -eq 'gh') {
                throw 'GitHub CLI is required. Install it with: winget install GitHub.cli'
            }
            throw "$tool is required but was not found in PATH."
        }
    }

    Invoke-Checked -Command gh -Arguments @('auth', 'status') | Out-Null

    $branch = Invoke-Checked -Command git -Arguments @('branch', '--show-current')
    if ($branch -ne 'main') {
        throw "Releases must be prepared from main; current branch is '$branch'."
    }

    $worktreeState = Invoke-Checked -Command git -Arguments @('status', '--porcelain')
    if (-not [string]::IsNullOrWhiteSpace($worktreeState)) {
        throw "The worktree must be clean before a release:`n$worktreeState"
    }

    Invoke-Checked -Command git -Arguments @('fetch', '--prune', 'origin', 'main', '--tags') | Out-Null
    $headCommit = Invoke-Checked -Command git -Arguments @('rev-parse', 'HEAD')
    $originCommit = Invoke-Checked -Command git -Arguments @('rev-parse', 'origin/main')
    if ($headCommit -ne $originCommit) {
        throw "Local main ($headCommit) is not synchronized with origin/main ($originCommit)."
    }

    & gh release view $releaseTag *> $null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "GitHub Release $releaseTag already exists; nothing to do." -ForegroundColor Green
        exit 0
    }

    $currentVersionText = (Get-Content -Raw $versionPath).Trim()
    try {
        $currentVersion = [version]$currentVersionText
        $targetVersion = [version]$Version
    }
    catch {
        throw "VERSION and -Version must both use X.Y.Z format. Current value: '$currentVersionText'."
    }

    if ($currentVersion -gt $targetVersion) {
        throw "Requested version $Version is older than repository version $currentVersionText."
    }

    if ($currentVersion -lt $targetVersion) {
        [IO.File]::WriteAllText($versionPath, "$Version`n", [Text.UTF8Encoding]::new($false))
        Invoke-Checked -Command git -Arguments @('add', '--', 'backend/cmd/server/VERSION') | Out-Null
        Invoke-Checked -Command git -Arguments @('commit', '-m', "chore(release): $releaseTag") | Out-Null
        Invoke-Checked -Command git -Arguments @('push', 'origin', 'main') | Out-Null
        $headCommit = Invoke-Checked -Command git -Arguments @('rev-parse', 'HEAD')
    }
    else {
        $subject = Invoke-Checked -Command git -Arguments @('log', '-1', '--pretty=%s')
        if ($subject -ne "chore(release): $releaseTag") {
            throw "VERSION is already $Version, but HEAD is not the matching release commit."
        }
    }

    Wait-RequiredChecks -Commit $headCommit

    & git rev-parse --verify --quiet "refs/tags/$releaseTag" *> $null
    $localTagExists = $LASTEXITCODE -eq 0
    if ($localTagExists) {
        $tagCommit = Invoke-Checked -Command git -Arguments @('rev-list', '-n', '1', $releaseTag)
        if ($tagCommit -ne $headCommit) {
            throw "Local tag $releaseTag points to $tagCommit instead of $headCommit."
        }
    }
    else {
        $notes = $null
        if (-not [string]::IsNullOrWhiteSpace($NotesFile)) {
            $resolvedNotesPath = (Resolve-Path $NotesFile).Path
            $notes = (Get-Content -Raw $resolvedNotesPath).Trim()
        }
        else {
            $tagList = (Invoke-Checked -Command git -Arguments @(
                'tag', '--sort=-version:refname', '--merged', 'HEAD', '-l', 'v[0-9]*'
            )) -split "`r?`n"
            $previousTag = $tagList |
                Where-Object { $_ -and $_ -ne $releaseTag } |
                Select-Object -First 1
            $range = if ($previousTag) { "$previousTag..HEAD" } else { 'HEAD' }
            $changes = Invoke-Checked -Command git -Arguments @('log', $range, '--pretty=format:- %s (%h)')
            $notes = "EasySub2api $Version`n`n$changes"
        }

        if ([string]::IsNullOrWhiteSpace($notes)) {
            throw 'Release notes cannot be empty.'
        }

        $temporaryNotes = Join-Path ([IO.Path]::GetTempPath()) "easysub2api-$Version-release-notes.md"
        try {
            [IO.File]::WriteAllText($temporaryNotes, "$notes`n", [Text.UTF8Encoding]::new($false))
            Invoke-Checked -Command git -Arguments @('tag', '-a', $releaseTag, '-F', $temporaryNotes, $headCommit) | Out-Null
        }
        finally {
            Remove-Item -LiteralPath $temporaryNotes -Force -ErrorAction SilentlyContinue
        }
    }

    $remoteTagLine = (& git ls-remote --tags origin "refs/tags/$releaseTag" 2>$null) -join ''
    if ([string]::IsNullOrWhiteSpace($remoteTagLine)) {
        Invoke-Checked -Command git -Arguments @('push', 'origin', $releaseTag) | Out-Null
    }

    $releaseRun = Get-ReleaseRun
    if ($null -eq $releaseRun -or $releaseRun.status -eq 'completed') {
        Invoke-Checked -Command gh -Arguments @('workflow', 'run', 'release.yml', '--ref', 'main', '-f', "tag=$releaseTag") | Out-Null
        $deadline = (Get-Date).AddMinutes(5)
        do {
            Start-Sleep -Seconds 5
            $releaseRun = Get-ReleaseRun
        } while ($null -eq $releaseRun -and (Get-Date) -lt $deadline)
    }

    if ($null -eq $releaseRun) {
        throw 'The release workflow was dispatched, but its run could not be located.'
    }
    if ($releaseRun.status -eq 'completed' -and $releaseRun.conclusion -ne 'success') {
        throw "The latest release run already failed. Inspect $($releaseRun.url)"
    }

    Write-Host "Watching $($releaseRun.url)"
    Invoke-Checked -Command gh -Arguments @('run', 'watch', [string]$releaseRun.databaseId, '--exit-status') | Out-Null
    Write-Host "EasySub2api $Version was published successfully." -ForegroundColor Green
}
finally {
    Pop-Location
}
