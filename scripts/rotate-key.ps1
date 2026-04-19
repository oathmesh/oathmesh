[CmdletBinding()]
param(
    [ValidateSet("Generate", "Check")]
    [string]$Mode = "Generate",

    [string]$IssuerUrl,

    [string]$OutputDir = ".\artifacts\key-rotation",

    [string]$Kid,

    [string]$PreviousKids
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-GeneratedKid {
    $now = Get-Date
    $bytes = New-Object byte[] 2
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $randHex = [System.BitConverter]::ToString($bytes).Replace("-", "").ToLowerInvariant()
    return "issuer-key-{0}-{1}-{2}" -f $now.ToString("yyyy"), $now.ToString("MM"), $randHex
}

function Get-JwksKids {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Issuer
    )

    $base = $Issuer.TrimEnd("/")
    $jwksUri = "$base/.well-known/jwks.json"
    $jwks = Invoke-RestMethod -Method Get -Uri $jwksUri -TimeoutSec 10
    if (-not $jwks.keys) {
        throw "JWKS payload has no 'keys' array: $jwksUri"
    }
    return @($jwks.keys | ForEach-Object { $_.kid } | Where-Object { $_ -and $_.Trim() -ne "" })
}

function Require-OpenSSL {
    $openssl = Get-Command openssl -ErrorAction SilentlyContinue
    if (-not $openssl) {
        throw "OpenSSL is required. Install OpenSSL and ensure 'openssl' is on PATH."
    }
}

if ($Mode -eq "Check") {
    if (-not $IssuerUrl) {
        throw "-IssuerUrl is required in Check mode."
    }

    $kids = Get-JwksKids -Issuer $IssuerUrl
    Write-Host "JWKS kids at $IssuerUrl"
    $kids | ForEach-Object { Write-Host " - $_" }

    if ($PreviousKids) {
        $previous = @($PreviousKids.Split(",") | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne "" })
        $newKids = @($kids | Where-Object { $previous -notcontains $_ })
        if ($newKids.Count -gt 0) {
            Write-Host ""
            Write-Host "Detected new kid(s):"
            $newKids | ForEach-Object { Write-Host " + $_" }
        }
        else {
            Write-Warning "No new kids detected compared to -PreviousKids."
        }
    }
    exit 0
}

Require-OpenSSL

if (-not $Kid) {
    $Kid = Get-GeneratedKid
}

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runDir = Join-Path $OutputDir $timestamp
New-Item -ItemType Directory -Path $runDir -Force | Out-Null

$privateKeyPath = Join-Path $runDir "private.pem"
$publicKeyPath = Join-Path $runDir "public.pem"
$metaPath = Join-Path $runDir "rotation-metadata.txt"

& openssl genpkey -algorithm Ed25519 -out $privateKeyPath | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "Failed to generate Ed25519 private key."
}

& openssl pkey -in $privateKeyPath -pubout -out $publicKeyPath | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "Failed to derive Ed25519 public key."
}

$pemText = Get-Content -Path $privateKeyPath -Raw

"kid=$Kid" | Set-Content -Path $metaPath
"generated_at_utc=$([DateTime]::UtcNow.ToString("o"))" | Add-Content -Path $metaPath
"private_key_path=$privateKeyPath" | Add-Content -Path $metaPath
"public_key_path=$publicKeyPath" | Add-Content -Path $metaPath

Write-Host "Generated new Ed25519 keypair:"
Write-Host " - Private key: $privateKeyPath"
Write-Host " - Public key : $publicKeyPath"
Write-Host " - Suggested kid label: $Kid"
Write-Host ""

if ($IssuerUrl) {
    try {
        $currentKids = Get-JwksKids -Issuer $IssuerUrl
        Write-Host "Current JWKS kids at ${IssuerUrl}:"
        $currentKids | ForEach-Object { Write-Host " - $_" }
        $currentKidCsv = ($currentKids -join ",")
        Write-Host ""
        Write-Host "Post-cutover verification:"
        Write-Host ".\scripts\rotate-key.ps1 -Mode Check -IssuerUrl $IssuerUrl -PreviousKids `"$currentKidCsv`""
    }
    catch {
        Write-Warning "Unable to fetch JWKS from $IssuerUrl. Continue with manual checks. Error: $($_.Exception.Message)"
    }
}

Write-Host ""
Write-Host "Manual operator steps (no direct cloud writes performed):"
Write-Host "1) Update issuer secret with NEW private key."
Write-Host "   - OATHMESH_PRIVATE_KEY      = <contents of private.pem>"
Write-Host "   - or OATHMESH_PRIVATE_KEY_B64 = <base64 PEM>"
Write-Host "2) Roll issuer deployment using your normal rolling restart."
Write-Host "3) Confirm /.well-known/jwks.json publishes a new kid."
Write-Host "4) Monitor verification errors through overlap window (>= TTL_MAX + JWKS cache TTL)."
Write-Host "5) Keep previous key available for rollback."
Write-Host ""
Write-Host "B64 helper command (prints to console only when you run it):"
Write-Host "[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes((Get-Content -Raw `"$privateKeyPath`")))"
