# Replace a Shared API Key with OathMesh in One Afternoon

## The Problem

You have two services. Service A calls Service B using a shared API key:

```
Authorization: Bearer sk-abc123def456
```

This key:
- Never expires
- Is shared across all callers — you can't tell *who* is calling
- Is stored in environment variables, CI secrets, and config files
- If leaked, gives unlimited access until manually rotated
- Produces no audit trail of individual calls

## The Fix

Replace the static key with short-lived, signed, auditable OathMesh tokens. Each call gets a unique identity that expires in seconds.

## Step 1: Deploy the Issuer (30 minutes)

```bash
# Clone OathMesh
git clone https://github.com/oathmesh/oathmesh.git
cd oathmesh

# Generate a signing key
openssl genpkey -algorithm Ed25519 -out private.pem

# Start the issuer
export OATHMESH_ISSUER=https://issuer.internal
export OATHMESH_PRIVATE_KEY="$(cat private.pem)"
go run ./cmd/oathmesh serve --port 4000
```

The issuer is now running. It publishes its public keys at `/.well-known/jwks.json`.

## Step 2: Update Service A — the Caller (30 minutes)

Before:
```python
response = requests.get("https://service-b.internal/data",
    headers={"Authorization": "Bearer sk-abc123def456"})
```

After:
```python
from oathmesh import verify_token, VerifierConfig

# Mint a token for this specific call
token = requests.post("https://issuer.internal/v1/token", json={
    "sub": "svc://service-a/worker",
    "aud": "https://service-b.internal",
    "act": "data.read",
    "ttl_hint": 60
}).json()["token"]

# Use the token instead of the API key
response = requests.get("https://service-b.internal/data",
    headers={"Authorization": f"OathMesh {token}"})
```

## Step 3: Update Service B — the Receiver (30 minutes)

Before:
```python
API_KEY = os.environ["SERVICE_B_API_KEY"]

def check_auth(request):
    if request.headers.get("Authorization") != f"Bearer {API_KEY}":
        raise HTTPException(401)
```

After:
```python
from oathmesh import verify_token, VerifierConfig, OathMeshError

config = VerifierConfig(
    audience="https://service-b.internal",
    trusted_issuers=["https://issuer.internal"],
)

def check_auth(request):
    try:
        caller = verify_token(request.headers.get("authorization", ""), config)
        return caller  # Contains subject, action, source provenance
    except OathMeshError as e:
        raise HTTPException(401, detail={"error": e.code})
```

## Step 4: Remove the API Key (5 minutes)

1. Delete `SERVICE_B_API_KEY` from all environment variables and CI secrets
2. Remove the key from Service B's validation code
3. Done. There is no static secret to rotate, store, or leak.

## What You Gained

| Before (API Key) | After (OathMesh) |
|---|---|
| Never expires | Expires in ≤300 seconds |
| No caller identity | Full caller identity: subject, action, source |
| No audit trail | NDJSON audit event on every call |
| Shared across all callers | Unique token per call |
| Manual rotation on leak | Automatic expiry, no rotation needed |
| No policy enforcement | Pkl policy rules with default deny |

## Total Time

| Step | Time |
|---|---|
| Deploy issuer | 30 minutes |
| Update caller | 30 minutes |
| Update receiver | 30 minutes |
| Remove API key | 5 minutes |
| **Total** | **~1.5 hours** |
