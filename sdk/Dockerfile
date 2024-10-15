FROM node:lts-slim AS base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

RUN corepack enable

RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 ca-certificates bash curl make g++ \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

FROM base AS builder

WORKDIR /usr/src/app

ARG NATIVE_LIBS_RELEASE_TAG
ENV NATIVE_LIBS_RELEASE_TAG=${NATIVE_LIBS_RELEASE_TAG}

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY ./sdk/package.json ./sdk/package.json
COPY ./ts-libs/linea-native-libs/package.json ./ts-libs/linea-native-libs/package.json

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile --prefer-offline --ignore-scripts

COPY ./sdk ./sdk
COPY ts-libs/linea-native-libs ./ts-libs/linea-native-libs

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm run build \
    && pnpm deploy --filter=./sdk --prod ./prod/sdk

FROM node:lts-slim AS production

ENV NODE_ENV=production

WORKDIR /usr/src/app

USER node

COPY --from=builder /usr/src/app/prod/sdk ./sdk

CMD [ "node", "./sdk/dist/scripts/runPostman.js" ]