import { beforeEach, describe, expect, it, vi } from 'vitest';
import { verifyOathToken } from '../src/index';
import { createOathMeshMiddleware, getOathMeshContext } from '../src/middleware/graphql';
import { checkPermission, wrapWithOathMeshDirective } from '../src/middleware/directive';
import type { OathMeshGraphQLConfig, OathMeshGraphQLContext } from '../src/middleware/types';

vi.mock('../src/index', async () => {
  const actual = await vi.importActual<typeof import('../src/index')>('../src/index');
  return { ...actual, verifyOathToken: vi.fn() };
});

const mockedVerifyOathToken = vi.mocked(verifyOathToken);

function makeClaims(subject: string, scope: string[] = ['action:read']) {
  return {
    principal: { issuer: 'https://issuer.test.local', subject },
    action: 'read',
    tokenId: `jti-${subject}`,
    environment: 'test',
    scope,
  };
}

function makeRequestContext(authorization: string, operation: 'query' | 'mutation' = 'query') {
  return {
    request: { http: { headers: new Map([['authorization', authorization]]) } },
    contextValue: {} as Record<string, unknown>,
    operation: { operation },
  };
}

describe('GraphQL middleware practical behavior', () => {
  let config: OathMeshGraphQLConfig;

  beforeEach(() => {
    mockedVerifyOathToken.mockReset();
    config = {
      verifier: {
        audience: 'https://api.test.local',
        trustedIssuers: ['https://issuer.test.local'],
      },
    };
  });

  it('rejects missing or non-OathMesh authorization headers', async () => {
    const middleware = createOathMeshMiddleware(config);
    const missing = { request: { http: { headers: new Map() } }, contextValue: {}, operation: { operation: 'query' } };

    await expect(middleware.didResolveOperation(missing)).rejects.toMatchObject({ code: 'claim_missing:token' });
    await expect(middleware.didResolveOperation(makeRequestContext('Bearer abc'))).rejects.toMatchObject({
      code: 'claim_missing:token',
    });
    expect(mockedVerifyOathToken).not.toHaveBeenCalled();
  });

  it('extracts OathMesh token, verifies it, and injects context', async () => {
    mockedVerifyOathToken.mockResolvedValueOnce(makeClaims('svc://graphql/injected'));
    const middleware = createOathMeshMiddleware(config);
    const requestContext = makeRequestContext('OathMesh token-123');

    await middleware.didResolveOperation(requestContext);

    expect(mockedVerifyOathToken).toHaveBeenCalledWith(
      'token-123',
      expect.objectContaining({ audience: 'https://api.test.local' })
    );
    const oathmesh = getOathMeshContext(requestContext.contextValue);
    expect(oathmesh?.verified).toBe(true);
    expect(oathmesh?.claims.principal.subject).toBe('svc://graphql/injected');
  });

  it('enforces query rate limits and invokes callback', async () => {
    const onRateLimitExceeded = vi.fn();
    config.rateLimits = { queriesPerMinute: 1 };
    config.onRateLimitExceeded = onRateLimitExceeded;
    mockedVerifyOathToken.mockResolvedValue(makeClaims('svc://graphql/ratelimit'));

    const middleware = createOathMeshMiddleware(config);
    await middleware.didResolveOperation(makeRequestContext('OathMesh rate-limit-token', 'query'));
    await expect(
      middleware.didResolveOperation(makeRequestContext('OathMesh rate-limit-token', 'query'))
    ).rejects.toMatchObject({ code: 'rate_limit_exceeded' });

    expect(onRateLimitExceeded).toHaveBeenCalledWith('svc://graphql/ratelimit', 'query');
  });

  it('tracks mutation and query limits separately', async () => {
    config.rateLimits = { queriesPerMinute: 1, mutationsPerMinute: 1 };
    mockedVerifyOathToken.mockResolvedValue(makeClaims('svc://graphql/separate-limits'));
    const middleware = createOathMeshMiddleware(config);

    await middleware.didResolveOperation(makeRequestContext('OathMesh shared-token', 'query'));
    await middleware.didResolveOperation(makeRequestContext('OathMesh shared-token', 'mutation'));
    await expect(
      middleware.didResolveOperation(makeRequestContext('OathMesh shared-token', 'query'))
    ).rejects.toMatchObject({ code: 'rate_limit_exceeded' });
  });
});

describe('directive permission behavior', () => {
  it('chains middleware auth context into directive permission checks', async () => {
    mockedVerifyOathToken.mockResolvedValueOnce(makeClaims('svc://graphql/allowed', ['action:read:user']));
    const middleware = createOathMeshMiddleware({
      verifier: { audience: 'https://api.test.local', trustedIssuers: ['https://issuer.test.local'] },
    });
    const requestContext = makeRequestContext('OathMesh auth-token');

    await middleware.didResolveOperation(requestContext);

    const resolver = vi.fn().mockResolvedValue('allowed');
    const wrapped = wrapWithOathMeshDirective(resolver, { require: 'action:read:user' }, {});
    const result = await wrapped({}, {}, requestContext.contextValue, {});

    expect(result).toBe('allowed');
    expect(resolver).toHaveBeenCalledOnce();
  });

  it('checkPermission enforces verified scope and wildcard', () => {
    const ctx: OathMeshGraphQLContext = {
      claims: makeClaims('svc://graphql/permissions', ['action:read:user', 'action:write:user']),
      verified: true,
    };
    expect(checkPermission(ctx, 'action:read:user')).toBe(true);
    expect(checkPermission(ctx, 'action:*')).toBe(true);
    expect(checkPermission(ctx, 'action:delete:user')).toBe(false);
    expect(checkPermission({ ...ctx, verified: false }, 'action:read:user')).toBe(false);
  });

  it('wrapWithOathMeshDirective blocks unauthorized resolver execution', async () => {
    const resolver = vi.fn().mockResolvedValue('secret');
    const wrapped = wrapWithOathMeshDirective(resolver, { require: 'action:admin' }, {});
    const context = { oathmesh: { claims: makeClaims('svc://graphql/denied', ['action:read']), verified: true } };

    const result = await wrapped({}, {}, context, {});
    expect(result).toBeNull();
    expect(resolver).not.toHaveBeenCalled();
  });
});
