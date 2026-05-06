$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$certDir = Join-Path $projectRoot "certs"

New-Item -ItemType Directory -Path $certDir -Force | Out-Null

docker run --rm `
  -v "${certDir}:/out" `
  alpine:3.20 `
  sh -c "apk add --no-cache openssl >/dev/null && openssl req -x509 -nodes -newkey rsa:2048 -keyout /out/server.key -out /out/server.crt -days 365 -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost,IP:127.0.0.1'"

Write-Host "Generated:"
Write-Host " - $certDir\server.crt"
Write-Host " - $certDir\server.key"
