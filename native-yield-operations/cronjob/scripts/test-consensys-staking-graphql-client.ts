/**
 * Manual integration runner for ConsensysStakingGraphQLClient.
 *
 * Example usage:
 * GRAPHQL_ENDPOINT=https://example/graphql \
 * TOKEN_URL=... \
 * CLIENT_ID=... \
 * CLIENT_SECRET=... \
 * AUDIENCE=... \
 * BEACON_NODE_RPC_URL=https://example/beacon \
 * pnpm --filter @consensys/linea-native-yield-cron-job exec tsx scripts/test-consensys-staking-graphql-client.ts
 */

import { BeaconNodeApiClient, WinstonLogger } from "@consensys/linea-shared-utils";
import { ConsensysStakingGraphQLClient } from "../src/clients/ConsensysStakingGraphQLClient";
import { OAuth2TokenClient } from "ts-libs/linea-shared-utils/src";
import { createApolloClient } from "../src/utils/createApolloClient";

// private readonly apolloClient: ApolloClient,

async function main() {
  const requiredEnvVars = [
    "GRAPHQL_ENDPOINT",
    "BEACON_NODE_RPC_URL",
    "TOKEN_URL",
    "CLIENT_ID",
    "CLIENT_SECRET",
    "AUDIENCE",
  ];
  const missing = requiredEnvVars.filter((key) => !process.env[key]);
  if (missing.length > 0) {
    console.error(`Missing required env vars: ${missing.join(", ")}`);
    process.exitCode = 1;
    return;
  }

  const beaconClient = new BeaconNodeApiClient(process.env.BEACON_NODE_RPC_URL);

  const tokenClient = new OAuth2TokenClient(
    new WinstonLogger("OAuth2TokenClient.integration"),
    process.env.TOKEN_URL!,
    process.env.CLIENT_ID!,
    process.env.CLIENT_SECRET!,
    process.env.AUDIENCE!,
  );

  const apolloClient = createApolloClient(tokenClient, process.env.GRAPHQL_ENDPOINT!);
  const consensysStakingClient = new ConsensysStakingGraphQLClient(
    apolloClient,
    beaconClient,
    new WinstonLogger("ConsensysStakingGraphQLClient.integration"),
  );

  try {
    const validators = await consensysStakingClient.getActiveValidatorsWithPendingWithdrawals();
    console.log(`Fetched ${validators.length} validators with pending withdrawals.`);
    console.log(validators);
    const totalPendingWei = consensysStakingClient.getTotalPendingPartialWithdrawalsWei(validators);
    console.log(`Total pending partial withdrawals (wei): ${totalPendingWei.toString()}`);
  } catch (err) {
    console.error("ConsensysStakingGraphQLClient integration script failed:", err);
    process.exitCode = 1;
  }
}

main();
