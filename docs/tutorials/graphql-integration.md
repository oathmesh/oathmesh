← [Back to Index](../INDEX.md)

# Tutorial: GraphQL Integration (Node + Python)

This tutorial uses the GraphQL middleware added in Phase 2 and references the repo’s runnable examples/tests.

## Prereq (local issuer token for manual calls)

```bash
TOKEN=$(curl -s http://localhost:4000/v1/token \
  -X POST \
  -H "Authorization: Bearer development_secret_do_not_use_in_prod" \
  -H "Content-Type: application/json" \
  -d '{"sub":"svc://local/graphql-client","aud":"https://api.example.com","act":"graphql.query","ttl_hint":120}' | jq -r '.token')
```

---

## Node (Apollo middleware)

References:

- Middleware: `sdk/node/src/middleware/graphql.ts`
- Example app: `sdk/node/src/middleware/examples/apollo-server.ts`
- Practical behavior tests: `sdk/node/test/graphql.middleware.test.ts`

Run middleware tests:

```bash
cd sdk/node
npm install
npm test -- graphql.middleware.test.ts
```

Expected: tests pass for token extraction, verification, context injection, and rate limits.

Resolver access pattern:

```ts
const oathmesh = getOathMeshContext(context);
const subject = oathmesh?.claims.principal.subject;
```

---

## Python (Strawberry middleware)

References:

- Middleware: `sdk/python/src/oathmesh/middleware/graphql.py`
- Example app: `sdk/python/src/oathmesh/middleware/examples/strawberry_server.py`
- Practical behavior tests: `sdk/python/tests/test_graphql_middleware.py`

Run middleware tests:

```bash
cd sdk/python
pip install -e .[test]
pytest tests/test_graphql_middleware.py -q
```

Expected: tests pass for auth header handling, verified context injection, and rate limiting.

Resolver access pattern:

```py
oathmesh = get_oathmesh_context(info.context)
subject = oathmesh.claims.principal.subject
```

---

## Manual protected GraphQL request shape

```bash
curl -X POST http://localhost:8000/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: OathMesh $TOKEN" \
  -d '{"query":"{ currentUser { id name } }"}'
```

Use your GraphQL server URL (`/graphql`) and set `aud` in minting to match that server’s configured audience.
