# 🚦 Quick Start

One minimal happy path: clone → build → run → mint → protected API success.

## Prerequisites

- Go toolchain that satisfies `go.mod` (currently `go 1.26.2`)
- Docker + Docker Compose
- OpenSSL
- curl

Canonical local endpoints from `docker-compose.yml`:
- Issuer: `http://localhost:4000`
- Protected chi API: `http://localhost:8081/inventory`

## Happy Path (bash / zsh)

```bash
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh

openssl genpkey -algorithm Ed25519 -out private.pem
go build -o bin/oathmesh ./cmd/oathmesh
docker compose up -d --build

TOKEN=$(OATHMESH_ISSUER=http://localhost:4000 OATHMESH_PRIVATE_KEY_FILE=./private.pem ./bin/oathmesh mint \
  --sub "agent://repo/acme/deploy-bot" \
  --aud "https://inventory.internal" \
  --act "deploy" \
  --quiet)

curl -i -H "Authorization: OathMesh $TOKEN" http://localhost:8081/inventory
```

Expected success: `HTTP/1.1 200 OK` with JSON from `chi-api`.

## Happy Path (PowerShell)

```powershell
openssl genpkey -algorithm Ed25519 -out private.pem
go build -o bin/oathmesh.exe ./cmd/oathmesh
docker compose up -d --build

$env:OATHMESH_ISSUER = "http://localhost:4000"
$env:OATHMESH_PRIVATE_KEY_FILE = ".\private.pem"
$TOKEN = & .\bin\oathmesh.exe mint --sub "agent://repo/acme/deploy-bot" --aud "https://inventory.internal" --act "deploy" --quiet
Invoke-WebRequest -Headers @{ Authorization = "OathMesh $TOKEN" } -Uri "http://localhost:8081/inventory"
```

Expected success: `StatusCode : 200`.

## Quick fixes

- Ports `4000` or `8081` busy: `docker compose down` then rerun `docker compose up -d --build`
- Minted token expired (max 300s): mint again

Next step (Step 2, guided onboarding): [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)  
Then continue to Step 3 (full docs index): [docs/INDEX.md](docs/INDEX.md)
