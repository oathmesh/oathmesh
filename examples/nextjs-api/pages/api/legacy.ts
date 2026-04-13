// Next.js Pages Router API route (legacy pattern)

import { withOathMeshApi } from '@oathmesh/sdk/next';

export default withOathMeshApi(
  {
    audience: process.env.OATHMESH_AUDIENCE || 'https://inventory.internal',
    trustedIssuers: (process.env.OATHMESH_TRUSTED_ISSUERS || 'http://localhost:4000').split(','),
  },
  (req, res) => {
    const caller = (req as any).oathmeshContext;
    res.json({
      status: 'success',
      caller: {
        subject: caller.principal.subject,
        action: caller.action,
      },
      note: 'This uses the Pages Router pattern. Prefer App Router for new projects.',
    });
  }
);
