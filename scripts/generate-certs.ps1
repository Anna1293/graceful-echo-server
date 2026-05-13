# Self-signed TLS cert for local HTTPS / HTTP/2 (reverse proxy on :8443).
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$certs = Join-Path $root "certs"
New-Item -ItemType Directory -Force -Path $certs | Out-Null

$key = Join-Path $certs "server.key"
$crt = Join-Path $certs "server.crt"

$openssl = Get-Command openssl -ErrorAction SilentlyContinue
if (-not $openssl) {
    Write-Error "openssl not found. Install Git for Windows (includes openssl) or add openssl to PATH."
}

& openssl req -x509 -newkey rsa:2048 `
    -keyout $key -out $crt `
    -days 365 -nodes `
    -subj "/CN=localhost" `
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

Write-Host "Wrote $crt and $key"
