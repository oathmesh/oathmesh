# FastAPI Example

OathMesh-protected FastAPI service using the `oathmesh` Python SDK.

## Run

```bash
cd examples/fastapi-api
pip install -r requirements.txt
OATHMESH_AUDIENCE=https://inventory.internal \
OATHMESH_TRUSTED_ISSUERS=http://localhost:4000 \
uvicorn main:app --host 0.0.0.0 --port 8000
```

## Test

```bash
TOKEN=$(oathmesh mint --sub "job://ci/nightly" \
  --aud "https://inventory.internal" --act "read" --quiet)

curl -H "Authorization: OathMesh $TOKEN" http://localhost:8000/inventory
```
