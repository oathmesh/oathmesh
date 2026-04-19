# fastapi-api Troubleshooting

## Request Path (FastAPI)

```text
Caller --> Authorization header --> Depends(require_oathmesh) --> endpoint
                                 | pass                       | fail
                                 v                            v
                             caller object                HTTP 401 detail
```

## Quick Checks

| Check | Command |
|---|---|
| Install deps | `pip install -r requirements.txt` |
| Run service | `uvicorn main:app --host 0.0.0.0 --port 8000` |
| Test call | `curl -H "Authorization: OathMesh $TOKEN" http://localhost:8000/inventory` |

## Common Issues

| Issue | Cause | Fix |
|---|---|---|
| `issuer_untrusted` | `trusted_issuers` mismatch | Use exact issuer URL |
| `audience_mismatch` | Wrong audience in config/token | Align token `--aud` with verifier config |
| `binding_mismatch` | `rqh` claim does not match request | Remove `rqh` for simple flows or compute canonical hash correctly |
| Import/runtime errors | Wrong Python env or missing package | Use Python 3.9+ and reinstall requirements |
