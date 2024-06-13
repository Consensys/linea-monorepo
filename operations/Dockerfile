# syntax=docker/dockerfile:1.2
FROM node:18-slim AS builder

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /usr/src/app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY ./operations/package.json ./operations/package.json

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile --prefer-offline && npm install -g typescript

COPY . .

RUN rm -rf src/synctx && pnpm run -F operations build

FROM node:18-slim as builder-synctx

ARG TARGETPLATFORM

WORKDIR /opt/synctx

COPY ./operations/src/synctx .

RUN apt-get update \
  && apt-get install --no-install-recommends -y git xz-utils perl \
  && OCLIF_TARGET=$(echo ${TARGETPLATFORM} | sed 's#/#-#;s#amd64#x64#') \
  && yarn global add oclif && yarn && yarn run build && yarn install --production --ignore-scripts --prefer-offline \
  && git init \
  && git config user.email "sre@consensys.net" \
  && git config user.name "cs-sre" \
  && git commit --allow-empty -m "dummy commit" \
  && oclif pack tarballs --targets="${OCLIF_TARGET}" \
  && tar -xvf dist/synctx-*.tar.gz

FROM node:18-slim as release

ENV NODE_ENV production
ENV PATH="${PATH}:/opt/synctx/bin"

WORKDIR /usr/src/app

# Install pnpm
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

COPY package.json pnpm-lock.yaml ./

RUN pnpm install --prod --frozen-lockfile --prefer-offline

COPY --chown=node:node --from=builder /usr/src/app/operations/dist ./dist
COPY --chown=node:node --from=builder-synctx /opt/synctx/synctx /opt/synctx/

USER node:node

ENTRYPOINT ["node"]
