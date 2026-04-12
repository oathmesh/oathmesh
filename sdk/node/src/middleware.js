const { createRemoteJWKSet, jwtVerify } = require('jose');

/**
 * Creates an Express middleware to verify OathMesh tokens
 */
function verifyToken(config) {
  const { audience, trustedIssuers } = config;
  
  if (!audience || !trustedIssuers) {
    throw new Error('audience and trustedIssuers are required');
  }

  // We maintain a map of JWKS providers for each trusted issuer
  const jwksMap = new Map();
  for (const issuer of trustedIssuers) {
    // OathMesh canonical jwks endpoint
    const url = new URL('/.well-known/jwks.json', issuer);
    jwksMap.set(issuer, createRemoteJWKSet(url, { cacheMaxAge: 300000 })); // 5 min
  }

  return async (req, res, next) => {
    try {
      const authHeader = req.headers['authorization'];
      if (!authHeader || !authHeader.startsWith('OathMesh ')) {
        return res.status(401).json({
          code: 'claim_missing:token',
          message: 'missing or invalid Authorization header'
        });
      }

      const token = authHeader.replace('OathMesh ', '');

      // In jose jwtVerify, if we don't know the issuer yet, we can't select the right JWKS.
      // Easiest path: try verify without fetching JWKS to decode the unverified issuer, 
      // or decode it manually, then fetch JWKS.
      // Fortunately jose jwtVerify exports decodeJwt as well if needed.
      const { decodeJwt } = require('jose');
      let payload;
      try {
        payload = decodeJwt(token);
      } catch (e) {
        return res.status(401).json({ code: 'verification_failed', message: 'malformed token' });
      }

      const iss = payload.iss;
      if (!iss || !trustedIssuers.includes(iss)) {
        return res.status(401).json({ code: 'unknown_issuer', message: 'issuer not trusted' });
      }

      const jwks = jwksMap.get(iss);
      const { payload: verifiedPayload } = await jwtVerify(token, jwks, {
        audience: audience,
        clockTolerance: 10,
        algorithms: ['EdDSA']
      });

      // Strict enforcement to match spec
      if (!verifiedPayload.act) throw new Error('missing act');
      if (!verifiedPayload.sub) throw new Error('missing sub');
      if (!verifiedPayload.jti) throw new Error('missing jti');

      req.oathmeshContext = {
        principal: {
          subject: verifiedPayload.sub,
          issuer: iss
        },
        action: verifiedPayload.act,
        tokenId: verifiedPayload.jti,
        environment: verifiedPayload.env || ''
      };

      // Strip auth
      delete req.headers['authorization'];
      
      next();
    } catch (error) {
      if (error.code === 'ERR_JWT_EXPIRED') {
        return res.status(401).json({ code: 'token_expired', message: 'token expired' });
      }
      if (error.code === 'ERR_JWT_AUDIENCE_INVALID' || error.message.includes('audience')) {
        return res.status(401).json({ code: 'audience_mismatch', message: 'wrong audience' });
      }
      return res.status(401).json({ 
        code: 'verification_failed', 
        message: error.message 
      });
    }
  };
}

module.exports = {
  verifyToken
};
