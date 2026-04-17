import { describe, it, expect, vi } from 'vitest';
import express, { type Request, type Response } from 'express';
import request from 'supertest';
import rateLimit from 'express-rate-limit';
import { verifyToken, verifyOathToken, extractToken, OathMeshError } from '../src/index';
import type { VerifierConfig, VerifiedCallerContext } from '../src/types';

// ─── extractToken ────────────────────────────────────────────────────────────

describe('extractToken', () => {
  it('extracts token from OathMesh prefix', () => {
    expect(extractToken('OathMesh abc.def.ghi')).toBe('abc.def.ghi');
  });

  it('returns null for missing header', () => {
    expect(extractToken(null)).toBeNull();
    expect(extractToken(undefined)).toBeNull();
    expect(extractToken('')).toBeNull();
  });

  it('returns null for unknown prefix', () => {
    expect(extractToken('Basic abc123')).toBeNull();
    expect(extractToken('Token abc123')).toBeNull();
  });
});

// ─── OathMeshError ───────────────────────────────────────────────────────────

describe('OathMeshError', () => {
  it('extends Error with code and fix', () => {
    const err = new OathMeshError('audience_mismatch', 'wrong audience', 'fix it');

    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe('OathMeshError');
    expect(err.code).toBe('audience_mismatch');
    expect(err.message).toBe('wrong audience');
    expect(err.fix).toBe('fix it');
  });

  it('serializes to JSON correctly', () => {
    const err = new OathMeshError('token_expired', 'expired', 'mint new');
    const json = err.toJSON();

    expect(json).toEqual({
      error: 'token_expired',
      message: 'expired',
      fix: 'mint new',
    });
  });
});

// ─── Express middleware ──────────────────────────────────────────────────────

describe('Express verifyToken middleware', () => {
  const config: VerifierConfig = {
    audience: 'https://inventory.internal',
    trustedIssuers: ['http://issuer.local'],
  };

  function createApp(cfg?: VerifierConfig) {
    const app = express();
    // Rate limiter for DoS protection (CodeQL js/missing-rate-limiting fix)
    const limiter = rateLimit({
      windowMs: 15 * 60 * 1000,
      max: 100,
    });
    app.use(limiter);
    app.use(verifyToken(cfg ?? config));
    app.get('/', (_req: Request, res: Response) => {
      res.json({ ok: true, caller: _req.oathmeshContext });
    });
    return app;
  }

  it('rejects missing Authorization header with 401', async () => {
    const res = await request(createApp()).get('/');
    expect(res.status).toBe(401);
    expect(res.body.error).toBe('claim_missing:token');
    expect(res.body.fix).toBeTruthy();
  });

  it('rejects non-OathMesh prefix with 401', async () => {
    const res = await request(createApp())
      .get('/')
      .set('Authorization', 'Basic abc123');
    expect(res.status).toBe(401);
    expect(res.body.error).toBe('claim_missing:token');
  });

  it('rejects malformed token with 401', async () => {
    const res = await request(createApp())
      .get('/')
      .set('Authorization', 'OathMesh not-a-real-token');
    expect(res.status).toBe(401);
    expect(res.body.error).toBeTruthy();
  });

  it('calls onDenied hook on rejection', async () => {
    const onDenied = vi.fn();
    const app = createApp({ ...config, onDenied });

    await request(app).get('/');
    expect(onDenied).toHaveBeenCalledOnce();
    expect(onDenied.mock.calls[0][0]).toBeInstanceOf(OathMeshError);
  });
});

// ─── Core verifier ──────────────────────────────────────────────────────────

describe('verifyOathToken', () => {
  const config: VerifierConfig = {
    audience: 'https://inventory.internal',
    trustedIssuers: ['http://issuer.local'],
  };

  it('rejects malformed tokens', async () => {
    await expect(verifyOathToken('garbage', config)).rejects.toThrow(OathMeshError);
  });

  it('rejects empty string', async () => {
    await expect(verifyOathToken('', config)).rejects.toThrow(OathMeshError);
  });
});
