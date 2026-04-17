import { OathMeshError } from './types';

export interface MintRequest {
  sub: string;
  aud: string;
  act: string;
  ttl_hint?: number;
  nbf_hint?: number;
  scope?: string[];
  reason?: string;
  env?: string;
  rqh?: string;
}

export interface MintResponse {
  token: string;
  expires_in: number;
  token_type: string;
}

export interface OathMeshClientConfig {
  /** The issuer URL (e.g. https://issuer.oathmesh.dev) */
  issuer: string;
  /** API key for authenticating with the issuer */
  apiKey: string;
}

/**
 * OathMeshClient auto-manages token minting and caching securely.
 */
export class OathMeshClient {
  private config: OathMeshClientConfig;
  private tokenCache: Map<string, { token: string; expiresAt: number }> = new Map();

  constructor(config: OathMeshClientConfig) {
    this.config = config;
  }

  /**
   * Mint a new token, caching it automatically until near its TTL.
   */
  async mint(req: MintRequest): Promise<string> {
    const cacheKey = JSON.stringify(req);
    const cached = this.tokenCache.get(cacheKey);

    // Auto-refresh: return cache if valid for at least 15 more secs
    if (cached && cached.expiresAt > Date.now() + 15000) {
      return cached.token;
    }

    const url = new URL('/v1/token', this.config.issuer);
    const res = await fetch(url.toString(), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.config.apiKey}`
      },
      body: JSON.stringify(req)
    });

    if (!res.ok) {
      throw new OathMeshError('verification_failed', `Failed to mint token: ${res.status} ${res.statusText}`);
    }

    const data = await res.json() as MintResponse;
    
    this.tokenCache.set(cacheKey, {
      token: data.token,
      expiresAt: Date.now() + data.expires_in * 1000
    });

    return data.token;
  }
}
