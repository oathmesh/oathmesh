import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { VerifierConfig } from '../src/types';
import { extractToken, verifyOathToken } from '../src/index';

const mockDecodeProtectedHeader = vi.fn();
const mockDecodeJwt = vi.fn();
const mockJwtVerify = vi.fn();

vi.mock('jose', () => ({
  createRemoteJWKSet: vi.fn(() => 'jwks'),
  decodeProtectedHeader: (token: string) => mockDecodeProtectedHeader(token),
  decodeJwt: (token: string) => mockDecodeJwt(token),
  jwtVerify: (...args: unknown[]) => mockJwtVerify(...args),
}));

const config: VerifierConfig = {
  audience: 'https://inventory.internal',
  trustedIssuers: ['https://issuer.local'],
};

function baseClaims(overrides: Record<string, unknown> = {}) {
  return {
    iss: 'https://issuer.local',
    sub: 'svc://node/conformance',
    aud: 'https://inventory.internal',
    act: 'read',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 120,
    jti: 'jti-node-conformance-1',
    ...overrides,
  };
}

describe('conformance parity cases', () => {
  beforeEach(() => {
    mockDecodeProtectedHeader.mockReset();
    mockDecodeJwt.mockReset();
    mockJwtVerify.mockReset();
  });

  it('token_parsing_validation_failures', async () => {
    mockDecodeProtectedHeader.mockImplementationOnce(() => {
      throw new Error('bad header');
    });

    await expect(verifyOathToken('not-a-token', config)).rejects.toMatchObject({
      code: 'token_malformed',
    });
  });

  it('issuer_check_untrusted', async () => {
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValueOnce(baseClaims({ iss: 'https://evil.local' }));

    await expect(verifyOathToken('token.part.sig', config)).rejects.toMatchObject({
      code: 'issuer_untrusted',
    });
  });

  it('audience_check_mismatch', async () => {
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValueOnce(baseClaims());
    mockJwtVerify.mockRejectedValueOnce(new Error('audience mismatch'));

    await expect(verifyOathToken('token.part.sig', config)).rejects.toMatchObject({
      code: 'audience_mismatch',
    });
  });

  it('replay_detection_semantics', async () => {
    mockDecodeProtectedHeader.mockReturnValue({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValue(baseClaims());
    mockJwtVerify.mockResolvedValue({
      payload: baseClaims(),
    });

    let seen = false;
    const replayConfig: VerifierConfig = {
      ...config,
      replayCache: {
        check: () => seen,
        add: () => {
          seen = true;
        },
      },
    };

    await verifyOathToken('token.part.sig', replayConfig);
    await expect(verifyOathToken('token.part.sig', replayConfig)).rejects.toMatchObject({
      code: 'replay_detected',
    });
  });

  it('middleware_auth_header_handling_semantics', () => {
    expect(extractToken(undefined)).toBeNull();
    expect(extractToken('Bearer abc')).toBeNull();
    expect(extractToken('OathMesh abc.def.ghi')).toBe('abc.def.ghi');
  });

  it('alg_none_rejection', async () => {
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'none', kid: 'k1' });
    await expect(verifyOathToken('token.part.sig', config)).rejects.toMatchObject({
      code: 'algorithm_not_allowed',
    });
  });

  it('subject_format_validation', async () => {
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValueOnce(baseClaims({ sub: 'not-a-valid-subject' }));
    await expect(verifyOathToken('token.part.sig', config)).rejects.toMatchObject({
      code: 'claim_missing:sub',
    });
  });

  it('binding_required_semantics', async () => {
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValueOnce(baseClaims());
    mockJwtVerify.mockResolvedValueOnce({ payload: baseClaims() });
    await expect(verifyOathToken('token.part.sig', { ...config, requireRequestBinding: true })).rejects.toMatchObject({
      code: 'binding_required',
    });
  });

  it('iat_future_rejection', async () => {
    const iat = Math.floor(Date.now() / 1000) + 60;
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValueOnce(baseClaims({ iat }));
    mockJwtVerify.mockResolvedValueOnce({ payload: baseClaims({ iat }) });
    await expect(verifyOathToken('token.part.sig', config)).rejects.toMatchObject({
      code: 'token_expired',
    });
  });
});
