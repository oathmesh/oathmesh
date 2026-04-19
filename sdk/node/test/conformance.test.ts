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
      code: 'verification_failed',
    });
  });

  it('issuer_check_untrusted', async () => {
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValueOnce({ iss: 'https://evil.local' });

    await expect(verifyOathToken('token', config)).rejects.toMatchObject({
      code: 'issuer_untrusted',
    });
  });

  it('audience_check_mismatch', async () => {
    mockDecodeProtectedHeader.mockReturnValueOnce({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValueOnce({ iss: 'https://issuer.local' });
    mockJwtVerify.mockRejectedValueOnce(new Error('audience mismatch'));

    await expect(verifyOathToken('token', config)).rejects.toMatchObject({
      code: 'audience_mismatch',
    });
  });

  it('replay_detection_semantics', async () => {
    mockDecodeProtectedHeader.mockReturnValue({ typ: 'om+jwt', alg: 'EdDSA', kid: 'k1' });
    mockDecodeJwt.mockReturnValue({ iss: 'https://issuer.local' });
    mockJwtVerify.mockResolvedValue({
      payload: {
        iss: 'https://issuer.local',
        sub: 'svc://node/conformance',
        aud: 'https://inventory.internal',
        act: 'read',
        jti: 'jti-node-conformance-1',
      },
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

    await verifyOathToken('token', replayConfig);
    await expect(verifyOathToken('token', replayConfig)).rejects.toMatchObject({
      code: 'replay_detected',
    });
  });

  it('middleware_auth_header_handling_semantics', () => {
    expect(extractToken(undefined)).toBeNull();
    expect(extractToken('Bearer abc')).toBeNull();
    expect(extractToken('OathMesh abc.def.ghi')).toBe('abc.def.ghi');
  });
});
