import { Hex } from "viem";
import { FlattenedConfigSchema } from "./config.schema";
import { transports } from "winston";

export const toClientConfig = (env: FlattenedConfigSchema) => ({
  dataSources: {
    chainId: env.CHAIN_ID,
    l1RpcUrl: env.L1_RPC_URL,
    beaconChainRpcUrl: env.BEACON_CHAIN_RPC_URL,
    stakingGraphQLUrl: env.STAKING_GRAPHQL_URL,
    ipfsBaseUrl: env.IPFS_BASE_URL,
  },
  consensysStakingOAuth2: {
    tokenEndpoint: env.CONSENSYS_STAKING_OAUTH2_TOKEN_ENDPOINT,
    clientId: env.CONSENSYS_STAKING_OAUTH2_CLIENT_ID,
    clientSecret: env.CONSENSYS_STAKING_OAUTH2_CLIENT_SECRET,
    audience: env.CONSENSYS_STAKING_OAUTH2_AUDIENCE,
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
    trigger: {
      // How often we poll for the trigger event
      pollIntervalMs: env.TRIGGER_EVENT_POLL_INTERVAL_MS,
      // Max tolerated time for inaction if trigger event polling doesn't find the trigger event
      maxInactionMs: env.TRIGGER_MAX_INACTION_MS,
    },
    contractReadRetryTimeMs: env.CONTRACT_READ_RETRY_TIME_MS,
  },
  rebalanceToleranceBps: env.REBALANCE_TOLERANCE_BPS,
  maxValidatorWithdrawalRequestsPerTransaction: env.MAX_VALIDATOR_WITHDRAWAL_REQUESTS_PER_TRANSACTION,
  minWithdrawalThresholdEth: env.MIN_WITHDRAWAL_THRESHOLD_ETH,
  web3signer: {
    url: env.WEB3SIGNER_URL,
    publicKey: env.WEB3SIGNER_PUBLIC_KEY as Hex, // address or secp pubkey (compressed/uncompressed)
    keystore: {
      path: env.WEB3SIGNER_KEYSTORE_PATH,
      passphrase: env.WEB3SIGNER_KEYSTORE_PASSPHRASE,
    },
    truststore: {
      path: env.WEB3SIGNER_TRUSTSTORE_PATH,
      passphrase: env.WEB3SIGNER_TRUSTSTORE_PASSPHRASE,
    },
    tlsEnabled: env.WEB3SIGNER_TLS_ENABLED,
  },
  loggerOptions: {
    level: "info",
    transports: [new transports.Console()],
  },
});

export type NativeYieldCronJobClientConfig = ReturnType<typeof toClientConfig>;
