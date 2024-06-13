FROM node:lts-alpine AS base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

FROM base as builder

# Create app directory
WORKDIR /usr/src/app

RUN mkdir -p sdk

# Install app dependencies
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./

COPY sdk/package.json ./sdk

COPY ./sdk/src/clients/blockchain/abis/*.json ./sdk/src/clients/blockchain/abis/

RUN --mount=type=cache,id=pnpm,target=/pnpm/store apk add --virtual build-dependencies --no-cache python3 make g++ \
    && pnpm install --frozen-lockfile \
    && apk del build-dependencies

COPY ./sdk ./sdk

RUN pnpm run -F ./sdk build

FROM base as production

ENV NODE_ENV production

# Create app directory
WORKDIR /usr/src/app

RUN mkdir -p sdk

# Install app dependencies
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml  ./

COPY sdk/package.json ./sdk

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --ignore-scripts \
    && pnpm add -g ts-node \
    && pnpm store prune

USER node

COPY --from=builder /usr/src/app/sdk/dist ./sdk/src
COPY --from=builder /usr/src/app/sdk/scripts ./sdk/scripts

CMD [ "ts-node", "./sdk/scripts/runPostman.ts" ]