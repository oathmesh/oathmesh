import { RequestHandler, Request } from 'express';

export interface VerifiedCallerContext {
  principal: {
    subject: string;
    issuer: string;
  };
  action: string;
  tokenId: string;
  environment: string;
}

declare global {
  namespace Express {
    interface Request {
      oathmeshContext?: VerifiedCallerContext;
    }
  }
}

export interface VerifierConfig {
  audience: string;
  trustedIssuers: string[];
}

export function verifyToken(config: VerifierConfig): RequestHandler;
