/**
 * Apollo Server Example with OathMesh GraphQL Middleware
 *
 * This example demonstrates how to:
 * 1. Set up Apollo Server v4
 * 2. Register OathMesh middleware for authentication
 * 3. Use directives for field-level authorization
 * 4. Access verified claims in resolvers
 *
 * Usage:
 * ```bash
 * # Install dependencies (if not already installed)
 * npm install @apollo/server graphql @oathmesh/sdk
 *
 * # Run the server
 * npx ts-node examples/apollo-server.ts
 * ```
 *
 * Testing:
 * ```bash
 * # Query without token → 401 error
 * curl -X POST http://localhost:4000/graphql \
 *   -H "Content-Type: application/json" \
 *   -d '{"query": "{ currentUser { id name } }"}'
 *
 * # Query with token
 * curl -X POST http://localhost:4000/graphql \
 *   -H "Content-Type: application/json" \
 *   -H "Authorization: OathMesh <token>" \
 *   -d '{"query": "{ currentUser { id name email } }"}'
 * ```
 */

import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { createOathMeshMiddleware, getOathMeshContext } from '../graphql';
import type { OathMeshGraphQLConfig } from '../types';

// ─── GraphQL Schema ──────────────────────────────────────────────────────────

const typeDefs = `#graphql
  """
  User represents an authenticated principal in the system.
  The email field is restricted to authorized callers.
  """
  type User {
    id: ID!
    name: String!
    email: String @oathmesh(require: "action:read:user:email")
  }

  type Query {
    """Get the currently authenticated user from the token."""
    currentUser: User!
  }

  type Mutation {
    """Update the current user's profile."""
    updateUser(name: String!): User!
  }

  directive @oathmesh(require: String!) on FIELD_DEFINITION
`;

// ─── Resolvers ───────────────────────────────────────────────────────────────

const resolvers = {
  Query: {
    currentUser: async (_: any, __: any, context: any) => {
      // Get verified claims from OathMesh middleware
      const oathmesh = getOathMeshContext(context);

      if (!oathmesh || !oathmesh.verified) {
        throw new Error('Not authenticated');
      }

      // In a real app, look up user from database using subject
      const subject = oathmesh.claims.principal.subject;
      return {
        id: '1',
        name: 'Test User',
        email: 'user@example.com', // Will be null if not authorized
      };
    },
  },

  Mutation: {
    updateUser: async (_: any, { name }: { name: string }, context: any) => {
      const oathmesh = getOathMeshContext(context);

      if (!oathmesh || !oathmesh.verified) {
        throw new Error('Not authenticated');
      }

      // Verify mutation permission
      if (!oathmesh.claims.scope?.includes('action:write:user:profile')) {
        throw new Error('Insufficient permissions to update user');
      }

      return {
        id: '1',
        name,
        email: 'user@example.com',
      };
    },
  },

  User: {
    email: (parent: any, _: any, context: any) => {
      const oathmesh = getOathMeshContext(context);

      // Check if caller has permission to read email
      if (!oathmesh?.verified || !oathmesh.claims.scope?.includes('action:read:user:email')) {
        // Return null for unauthorized access
        return null;
      }

      return parent.email;
    },
  },
};

// ─── Server Setup ────────────────────────────────────────────────────────────

async function startServer() {
  /**
   * OathMesh configuration.
   * In production, load from environment variables.
   */
  const oathMeshConfig: OathMeshGraphQLConfig = {
    verifier: {
      // The audience URL this API expects
      audience: 'https://api.example.com',

      // Trusted issuer(s) — add your OathMesh issuer URLs here
      trustedIssuers: [
        'https://issuer.oathmesh.tech',
        // 'https://your-issuer.example.com',
      ],

      // Optional: Enforce request binding for mutations
      requireRequestBinding: false,
    },

    // Rate limiting configuration
    rateLimits: {
      queriesPerMinute: 100,  // Default: 100
      mutationsPerMinute: 10, // Default: 10
    },

    // Optional: Log rate limit violations
    onRateLimitExceeded: (subject: string, operationType: string) => {
      console.warn(`⚠️  Rate limit exceeded for ${subject} (${operationType})`);
    },
  };

  // Create Apollo Server
  const server = new ApolloServer({
    typeDefs,
    resolvers,
    plugins: [
      createOathMeshMiddleware(oathMeshConfig) as any,
    ],
  });

  // Start the server
  const { url } = await startStandaloneServer(server, {
    listen: { port: 4000 },
    context: async () => ({}), // Initialize empty context
  });

  console.log(`🚀 Server ready at ${url}`);
  console.log(`📚 GraphQL Playground: ${url}`);
  console.log(`
💡 Usage:

  # Without token → 401 error
  curl -X POST ${url} \\
    -H "Content-Type: application/json" \\
    -d '{"query": "{ currentUser { id name } }"}'

  # With token → access allowed (if valid and authorized)
  curl -X POST ${url} \\
    -H "Content-Type: application/json" \\
    -H "Authorization: OathMesh <token>" \\
    -d '{"query": "{ currentUser { id name email } }"}'
  `);
}

// Start server
startServer().catch(err => {
  console.error('Failed to start server:', err);
  process.exit(1);
});
