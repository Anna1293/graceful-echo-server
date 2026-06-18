#!/bin/sh
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CERTS="$ROOT/certs"
mkdir -p "$CERTS"

KEY="$CERTS/server.key"
CRT="$CERTS/server.crt"

if ! command -v openssl >/dev/null 2>&1; then
	echo "openssl not found. Install openssl and retry." >&2
	exit 1
fi

openssl req -x509 -newkey rsa:2048 \
	-keyout "$KEY" -out "$CRT" \
	-days 365 -nodes \
	-subj "/CN=localhost" \
	-addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

echo "Wrote $CRT and $KEY"
