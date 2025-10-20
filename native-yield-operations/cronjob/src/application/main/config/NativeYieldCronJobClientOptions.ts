import { Config } from "./config.schema";
import { transports } from "winston";

export const toClientOptions = (env: Config) => ({
  yieldOptions: {
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
  },
  apiOptions: env.API_PORT,
  trigger: {
    triggerEventPollingTimeSeconds: env.TRIGGER_EVENT_POLLING_TIME_SECONDS,
    triggerFallbackDelaySeconds: env.TRIGGER_FALLBACK_DELAY_SECONDS,
  },
  loggerOptions: {
    level: "info",
    transports: [new transports.Console()],
  },
});

export type NativeYieldCronJobClientOptions = ReturnType<typeof toClientOptions>;
