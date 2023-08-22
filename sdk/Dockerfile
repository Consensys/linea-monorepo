FROM node:lts-alpine AS builder

# Create app directory
WORKDIR /usr/src/app

# Install app dependencies
COPY package*.json ./

COPY src/lib/abis/*.json ./src/lib/abis/

RUN npm ci

COPY . .

RUN npm run build

FROM node:slim

ENV NODE_ENV production

# Create app directory
WORKDIR /usr/src/app

# Install app dependencies
COPY package*.json ./

RUN npm ci --production

RUN npm install -g ts-node

USER node

COPY --from=builder /usr/src/app/dist ./src
COPY --from=builder /usr/src/app/scripts ./scripts

CMD [ "ts-node", "./scripts/runPostman.ts" ]