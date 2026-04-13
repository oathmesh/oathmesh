FROM node:20-alpine AS base

# ── Dependencies ─────────────────────────────────────────────────────────────
FROM base AS deps
WORKDIR /app

# Copy and build the SDK first (layer cache)
COPY sdk/node/ /sdk/node/
WORKDIR /sdk/node
RUN npm install && npx tsc

# Copy example dependencies
WORKDIR /app
COPY examples/nextjs-api/package.json ./
RUN npm install

# ── Build ────────────────────────────────────────────────────────────────────
FROM base AS builder
WORKDIR /app

COPY --from=deps /sdk/node /sdk/node
COPY --from=deps /app/node_modules ./node_modules
COPY examples/nextjs-api/ ./

ENV NEXT_TELEMETRY_DISABLED=1
RUN npx next build

# ── Runtime ──────────────────────────────────────────────────────────────────
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3001
ENV PORT=3001
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
