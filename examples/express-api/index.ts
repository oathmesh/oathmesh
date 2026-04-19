// examples/express-api — Express.js example with OathMesh verification.
//
// Usage:
//   cd examples/express-api
//   npm install
//   OATHMESH_AUDIENCE=https://inventory.internal \
//   OATHMESH_TRUSTED_ISSUERS=https://issuer.oathmesh.tech \
//   npx ts-node index.ts

import express from 'express';
import rateLimit from 'express-rate-limit';
import { verifyToken } from '@oathmesh/sdk';

const app = express();

const audience = process.env.OATHMESH_AUDIENCE || 'https://inventory.internal';
const issuers = (process.env.OATHMESH_TRUSTED_ISSUERS || 'http://localhost:4000').split(',');

// Rate limiter: 100 requests per 15 minutes per IP
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100,
  standardHeaders: true,
  legacyHeaders: false,
});
app.use(limiter);

// Mount OathMesh verification on all routes below
app.use(verifyToken({
  audience,
  trustedIssuers: issuers,
  onDenied: (err) => {
    console.warn(`[oathmesh] denied: ${err.code} — ${err.message}`);
  },
  onVerified: (ctx) => {
    console.log(`[oathmesh] allowed: ${ctx.principal.subject} → ${ctx.action}`);
  },
}));

// Protected endpoint
app.get('/inventory', (req, res) => {
  const caller = req.oathmeshContext!;
  res.json({
    status: 'success',
    data: ['widget-a', 'widget-b', 'widget-c'],
    caller: {
      subject: caller.principal.subject,
      action: caller.action,
      tokenId: caller.tokenId,
      environment: caller.environment,
    },
  });
});

// Health check (no auth required — mounted before middleware in real apps)
app.get('/healthz', (_req, res) => {
  res.send('OK');
});

const port = process.env.PORT || 3000;
app.listen(port, () => {
  console.log(`express-api listening on :${port}`);
});
