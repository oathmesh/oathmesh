FROM node:20-alpine

WORKDIR /app

# Copy SDK first for layer caching
COPY sdk/node/ /sdk/node/
WORKDIR /sdk/node
RUN npm install && npx tsc

# Copy example
WORKDIR /app
COPY examples/express-api/package.json ./
RUN npm install
COPY examples/express-api/ ./

EXPOSE 3000
CMD ["npx", "ts-node", "index.ts"]
