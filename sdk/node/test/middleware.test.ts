import { describe, it, expect } from 'vitest';
import express, { type Request, type Response } from 'express';
import request from 'supertest';
import { verifyToken, OathMeshError } from '../src/index';

describe('OathMesh Express Middleware', () => {
  const config = {
    audience: 'https://inventory.internal',
    trustedIssuers: ['http://issuer.local'],
  };

  function createApp() {
    const app = express();
    app.use(verifyToken(config));
    app.get('/', (_req: Request, res: Response) => {
      res.json({ ok: true, caller: _req.oathmeshContext });
    });
    return app;
  }

  it('rejects requests without Authorization header', async () => {
    const app = createApp();
    const res = await request(app).get('/');

    expect(res.status).toBe(401);
    expect(res.body.error).toBe('claim_missing:token');
    expect(res.body.fix).toBeTruthy();
  });

  it('rejects requests with Bearer instead of OathMesh prefix', async () => {
    const app = createApp();
    const res = await request(app).get('/').set('Authorization', 'Bearer fake');

    expect(res.status).toBe(401);
    expect(res.body.error).toBe('claim_missing:token');
  });

  it('rejects malformed tokens', async () => {
    const app = createApp();
    const res = await request(app).get('/').set('Authorization', 'OathMesh not-a-real-token');

    expect(res.status).toBe(401);
    // Should fail at header decode or issuer check
    expect(res.body.error).toBeTruthy();
  });

  it('OathMeshError has correct shape', () => {
    const err = new OathMeshError('audience_mismatch', 'wrong audience', 'fix it');

    expect(err.code).toBe('audience_mismatch');
    expect(err.message).toBe('wrong audience');
    expect(err.fix).toBe('fix it');
    expect(err.name).toBe('OathMeshError');
    expect(err instanceof Error).toBe(true);
  });
});
