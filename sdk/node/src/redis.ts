import type { RevocationList } from './types';

/**
 * Redis client interface expected by the cache (compatible with node-redis).
 */
export interface RedisLike {
  get(key: string): Promise<string | null>;
}

export class RedisRevocationCache implements RevocationList {
  private client: RedisLike;
  private prefix: string;

  constructor(client: RedisLike, prefix: string = 'om:rev:') {
    this.client = client;
    this.prefix = prefix;
  }

  async isRevoked(subject: string): Promise<boolean> {
    const val = await this.client.get(`${this.prefix}${subject}`);
    return val !== null;
  }
}
