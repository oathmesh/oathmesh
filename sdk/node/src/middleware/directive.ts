/**
 * GraphQL Field Authorization Directive
 *
 * Implements @oathmesh directive for field-level access control.
 * Resolvers decorated with this directive check permissions before resolving.
 */

import { GraphQLError } from 'graphql';
import type { OathMeshGraphQLContext } from './types';

/**
 * GraphQL schema directive definition for @oathmesh.
 * Add this to your schema:
 *
 * ```graphql
 * directive @oathmesh(require: String!) on FIELD_DEFINITION
 * ```
 *
 * Use it to protect fields:
 *
 * ```graphql
 * type User {
 *   id: ID!
 *   email: String! @oathmesh(require: "action:read:user:email")
 * }
 * ```
 */
export const oathMeshDirectiveTypeDefs = `
  directive @oathmesh(require: String!) on FIELD_DEFINITION
`;

/**
 * Check if a claim satisfies a requirement string.
 * 
 * Requirement formats:
 * - "action:read:user:email" — matches exact action
 * - "role:admin" — matches if scope includes "role:admin"
 * - "scope:*" — any scope allowed
 *
 * For MVP, we check if the requirement appears in token scopes.
 */
export function checkPermission(context: OathMeshGraphQLContext | null, requirement: string): boolean {
  if (!context || !context.verified) {
    return false;
  }

  const claims = context.claims;
  if (!claims.scope) {
    return false;
  }

  // Exact scope match
  if (claims.scope.includes(requirement)) {
    return true;
  }

  // Wildcard match (scope:*)
  if (requirement.includes(':*')) {
    const prefix = requirement.split(':*')[0];
    return claims.scope.some(s => s.startsWith(prefix));
  }

  return false;
}

/**
 * Wrap a GraphQL field resolver with @oathmesh directive enforcement.
 *
 * @param resolve Original field resolver
 * @param directive Directive arguments (contains 'require' string)
 * @param context GraphQL context (should contain oathmesh property)
 * @returns Wrapped resolver that checks permissions
 *
 * @example
 * ```typescript
 * const resolvers = {
 *   User: {
 *     email: wrapWithOathMeshDirective(
 *       (parent) => parent.email,
 *       { require: "action:read:user:email" },
 *       context
 *     )
 *   }
 * };
 * ```
 */
export function wrapWithOathMeshDirective(
  resolve: (parent: any, args: any, context: any, info: any) => any,
  directive: { require: string },
  context: any
) {
  return async (parent: any, args: any, ctx: any, info: any) => {
    const oathmesh = ctx?.oathmesh as OathMeshGraphQLContext | null;

    if (!checkPermission(oathmesh, directive.require)) {
      // Return null and add error to response errors
      return null;
    }

    // Permission granted — call original resolver
    const result = resolve(parent, args, ctx, info);
    return result instanceof Promise ? result : result;
  };
}

/**
 * Create a directive schema mapping for Apollo Server.
 * Use with @apollo/server's built-in directive support.
 *
 * @example
 * ```typescript
 * import { makeDirectivePlugin } from '@apollo/server';
 *
 * const directives = createOathMeshDirectiveMap();
 * const server = new ApolloServer({
 *   typeDefs: [oathMeshDirectiveTypeDefs, ...],
 *   resolvers,
 *   plugins: [
 *     makeDirectivePlugin(directives),
 *   ],
 * });
 * ```
 */
export function createOathMeshDirectiveMap() {
  return {
    oathmesh: {
      // Hook called when directive is encountered
      hook: 'willResolveField',
      handler: (
        _: any,
        _fieldName: string,
        fieldConfig: any,
        objectType: any,
        schemaDirectiveVisitorClass: any
      ) => {
        const originalResolve = fieldConfig.resolve;

        fieldConfig.resolve = async (source: any, args: any, context: any, info: any) => {
          const oathmesh = context?.oathmesh as OathMeshGraphQLContext | null;
          const requirement = (info.parentType.getFields()[info.fieldName] as any)._directives?.find(
            (d: any) => d.name === 'oathmesh'
          )?.args?.require;

          if (!requirement || !checkPermission(oathmesh, requirement)) {
            throw new GraphQLError('Forbidden', {
              nodes: info.fieldNodes,
              extensions: {
                code: 'FORBIDDEN',
              },
            });
          }

          // Call original resolver
          if (originalResolve) {
            return originalResolve(source, args, context, info);
          }
          return source[info.fieldName];
        };
      },
    },
  };
}
