FROM node:18-alpine AS deps

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:18-alpine AS builder

ARG NEXT_PUBLIC_WALLET_CONNECT_ID
ARG NEXT_PUBLIC_INFURA_ID
ENV NEXT_PUBLIC_WALLET_CONNECT_ID=$NEXT_PUBLIC_WALLET_CONNECT_ID
ENV NEXT_PUBLIC_INFURA_ID=$NEXT_PUBLIC_INFURA_ID

ENV NODE_ENV=production
ARG ENV_FILE
WORKDIR /app
COPY --from=deps /app ./
COPY . ./

COPY $ENV_FILE .env.production
RUN npm run build

FROM node:18-alpine AS runner

ARG X_TAG
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app ./

# Execute script
RUN chown -R node:node ./.next/

# Disable next telemetry.
ENV NEXT_TELEMETRY_DISABLED 1

USER node
CMD ["node_modules/.bin/next", "start"]