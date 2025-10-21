import { FlattenedConfigSchema } from "./config.schema";
import { transports } from "winston";

export const toClientConfig = (env: FlattenedConfigSchema) => ({
    dataSources: {
      l1RpcUrl: env.L1_RPC_URL,
      stakingGraphQLUrl: env.STAKING_GRAPHQL_URL,
      ipfsBaseUrl: env.IPFS_BASE_URL,
    },
    contractAddresses: {
      lineaRollupContractAddress: env.LINEA_ROLLUP_ADDRESS,
      lazyOracleAddress: env.LAZY_ORACLE_ADDRESS,
      yieldManagerAddress: env.YIELD_MANAGER_ADDRESS,
      lidoYieldProviderAddress: env.LIDO_YIELD_PROVIDER_ADDRESS,
      l2YieldRecipientAddress: env.L2_YIELD_RECIPIENT,
    },
  apiOptions: env.API_PORT,
  timing: {
    // How often we poll for the trigger event
    triggerEventPollingTimeSeconds: env.TRIGGER_EVENT_POLLING_TIME_SECONDS,
    // Max tolerated time for inaction if trigger event polling doesn't find the trigger event
    triggerMaxInactionTimeoutSeconds: env.TRIGGER_MAX_INACTION_TIMEOUT_SECONDS,
    contractReadRetryTimeSeconds: env.CONTRACT_READ_RETRY_TIME_SECONDS
  },
  loggerOptions: {
    level: "info",
    transports: [new transports.Console()],
  },
});

export type NativeYieldCronJobClientConfig = ReturnType<typeof toClientConfig>;
