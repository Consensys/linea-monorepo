import { transports, LoggerOptions } from "winston";

import { validateEventsFilters } from "./ConfigValidator";

import type { PostmanOptions } from "../../application/config/PostmanConfig";
import type { DBOptions } from "../persistence/config/types";

function parseOptionalInt(envVar: string | undefined): number | undefined {
  if (envVar === undefined || envVar === "") return undefined;
  const parsed = parseInt(envVar);
  return isNaN(parsed) ? undefined : parsed;
}

function parseOptionalBigInt(envVar: string | undefined): bigint | undefined {
  if (envVar === undefined || envVar === "") return undefined;
  return BigInt(envVar);
}

function parseOptionalFloat(envVar: string | undefined): number | undefined {
  if (envVar === undefined || envVar === "") return undefined;
  const parsed = parseFloat(envVar);
  return isNaN(parsed) ? undefined : parsed;
}

function buildEventFilters(
  fromAddressEnv: string | undefined,
  toAddressEnv: string | undefined,
  calldataEnv: string | undefined,
  calldataFunctionInterfaceEnv: string | undefined,
): PostmanOptions["l1Options"]["listener"]["eventFilters"] | undefined {
  if (!fromAddressEnv && !toAddressEnv && !(calldataEnv && calldataFunctionInterfaceEnv)) {
    return undefined;
  }

  return {
    fromAddressFilter: fromAddressEnv,
    toAddressFilter: toAddressEnv,
    ...(calldataEnv && calldataFunctionInterfaceEnv
      ? {
          calldataFilter: {
            criteriaExpression: calldataEnv,
            calldataFunctionInterface: calldataFunctionInterfaceEnv,
          },
        }
      : {}),
  };
}

export function loadConfigFromEnv(env: NodeJS.ProcessEnv = process.env): PostmanOptions {
  const l1EventFilters = buildEventFilters(
    env.L1_EVENT_FILTER_FROM_ADDRESS,
    env.L1_EVENT_FILTER_TO_ADDRESS,
    env.L1_EVENT_FILTER_CALLDATA,
    env.L1_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE,
  );

  const l2EventFilters = buildEventFilters(
    env.L2_EVENT_FILTER_FROM_ADDRESS,
    env.L2_EVENT_FILTER_TO_ADDRESS,
    env.L2_EVENT_FILTER_CALLDATA,
    env.L2_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE,
  );

  if (l1EventFilters) {
    validateEventsFilters(l1EventFilters);
  }

  if (l2EventFilters) {
    validateEventsFilters(l2EventFilters);
  }

  const databaseOptions: DBOptions = {
    type: "postgres",
    host: env.POSTGRES_HOST ?? "127.0.0.1",
    port: parseInt(env.POSTGRES_PORT ?? "5432"),
    username: env.POSTGRES_USER ?? "postgres",
    password: env.POSTGRES_PASSWORD ?? "postgres",
    database: env.POSTGRES_DB ?? "postman_db",
    ...(env.POSTGRES_SSL === "true"
      ? {
          ssl: {
            rejectUnauthorized: env.POSTGRES_SSL_REJECT_UNAUTHORIZED === "true",
            ca: env.POSTGRES_SSL_CA_PATH ?? undefined,
          },
        }
      : {}),
  };

  const loggerOptions: LoggerOptions = {
    level: env.LOG_LEVEL ?? "info",
    transports: [new transports.Console()],
  };

  return {
    l1Options: {
      rpcUrl: env.L1_RPC_URL ?? "",
      messageServiceContractAddress: env.L1_CONTRACT_ADDRESS ?? "",
      isEOAEnabled: env.L1_L2_EOA_ENABLED === "true",
      isCalldataEnabled: env.L1_L2_CALLDATA_ENABLED === "true",
      listener: {
        pollingInterval: parseOptionalInt(env.L1_LISTENER_INTERVAL),
        receiptPollingInterval: parseOptionalInt(env.L1_LISTENER_RECEIPT_POLLING_INTERVAL),
        maxFetchMessagesFromDb: parseOptionalInt(env.MAX_FETCH_MESSAGES_FROM_DB),
        maxBlocksToFetchLogs: parseOptionalInt(env.L1_MAX_BLOCKS_TO_FETCH_LOGS),
        ...(parseInt(env.L1_LISTENER_INITIAL_FROM_BLOCK ?? "") >= 0
          ? { initialFromBlock: parseInt(env.L1_LISTENER_INITIAL_FROM_BLOCK ?? "") }
          : {}),
        ...(parseInt(env.L1_LISTENER_BLOCK_CONFIRMATION ?? "") >= 0
          ? { blockConfirmation: parseInt(env.L1_LISTENER_BLOCK_CONFIRMATION ?? "") }
          : {}),
        ...(l1EventFilters ? { eventFilters: l1EventFilters } : {}),
      },
      claiming: {
        signerPrivateKey: env.L1_SIGNER_PRIVATE_KEY ?? "",
        messageSubmissionTimeout: parseOptionalInt(env.MESSAGE_SUBMISSION_TIMEOUT),
        maxNonceDiff: parseOptionalInt(env.MAX_NONCE_DIFF),
        maxFeePerGasCap: parseOptionalBigInt(env.MAX_FEE_PER_GAS_CAP),
        gasEstimationPercentile: parseOptionalInt(env.GAS_ESTIMATION_PERCENTILE),
        profitMargin: parseOptionalFloat(env.PROFIT_MARGIN),
        maxNumberOfRetries: parseOptionalInt(env.MAX_NUMBER_OF_RETRIES),
        retryDelayInSeconds: parseOptionalInt(env.RETRY_DELAY_IN_SECONDS),
        maxClaimGasLimit: parseOptionalBigInt(env.MAX_CLAIM_GAS_LIMIT),
        maxTxRetries: parseOptionalInt(env.MAX_TX_RETRIES),
        isMaxGasFeeEnforced: env.L1_MAX_GAS_FEE_ENFORCED === "true",
        isPostmanSponsorshipEnabled: env.L2_L1_ENABLE_POSTMAN_SPONSORING === "true",
        maxPostmanSponsorGasLimit: parseOptionalBigInt(env.MAX_POSTMAN_SPONSOR_GAS_LIMIT),
        claimViaAddress: env.L1_CLAIM_VIA_ADDRESS,
      },
    },
    l2Options: {
      rpcUrl: env.L2_RPC_URL ?? "",
      messageServiceContractAddress: env.L2_CONTRACT_ADDRESS ?? "",
      isEOAEnabled: env.L2_L1_EOA_ENABLED === "true",
      isCalldataEnabled: env.L2_L1_CALLDATA_ENABLED === "true",
      listener: {
        pollingInterval: parseOptionalInt(env.L2_LISTENER_INTERVAL),
        receiptPollingInterval: parseOptionalInt(env.L2_LISTENER_RECEIPT_POLLING_INTERVAL),
        maxFetchMessagesFromDb: parseOptionalInt(env.MAX_FETCH_MESSAGES_FROM_DB),
        maxBlocksToFetchLogs: parseOptionalInt(env.L2_MAX_BLOCKS_TO_FETCH_LOGS),
        ...(parseInt(env.L2_LISTENER_INITIAL_FROM_BLOCK ?? "") >= 0
          ? { initialFromBlock: parseInt(env.L2_LISTENER_INITIAL_FROM_BLOCK ?? "") }
          : {}),
        ...(parseInt(env.L2_LISTENER_BLOCK_CONFIRMATION ?? "") >= 0
          ? { blockConfirmation: parseInt(env.L2_LISTENER_BLOCK_CONFIRMATION ?? "") }
          : {}),
        ...(l2EventFilters ? { eventFilters: l2EventFilters } : {}),
      },
      claiming: {
        signerPrivateKey: env.L2_SIGNER_PRIVATE_KEY ?? "",
        messageSubmissionTimeout: parseOptionalInt(env.MESSAGE_SUBMISSION_TIMEOUT),
        maxNonceDiff: parseOptionalInt(env.MAX_NONCE_DIFF),
        maxFeePerGasCap: parseOptionalBigInt(env.MAX_FEE_PER_GAS_CAP),
        gasEstimationPercentile: parseOptionalInt(env.GAS_ESTIMATION_PERCENTILE),
        profitMargin: parseOptionalFloat(env.PROFIT_MARGIN),
        maxNumberOfRetries: parseOptionalInt(env.MAX_NUMBER_OF_RETRIES),
        retryDelayInSeconds: parseOptionalInt(env.RETRY_DELAY_IN_SECONDS),
        maxClaimGasLimit: parseOptionalBigInt(env.MAX_CLAIM_GAS_LIMIT),
        maxTxRetries: parseOptionalInt(env.MAX_TX_RETRIES),
        isMaxGasFeeEnforced: env.L2_MAX_GAS_FEE_ENFORCED === "true",
        isPostmanSponsorshipEnabled: env.L1_L2_ENABLE_POSTMAN_SPONSORING === "true",
        maxPostmanSponsorGasLimit: parseOptionalBigInt(env.MAX_POSTMAN_SPONSOR_GAS_LIMIT),
        claimViaAddress: env.L2_CLAIM_VIA_ADDRESS,
      },
      l2MessageTreeDepth: parseOptionalInt(env.L2_MESSAGE_TREE_DEPTH),
      enableLineaEstimateGas: env.ENABLE_LINEA_ESTIMATE_GAS === "true",
    },
    l1L2AutoClaimEnabled: env.L1_L2_AUTO_CLAIM_ENABLED === "true",
    l2L1AutoClaimEnabled: env.L2_L1_AUTO_CLAIM_ENABLED === "true",
    loggerOptions,
    databaseOptions,
    databaseCleanerOptions: {
      enabled: env.DB_CLEANER_ENABLED === "true",
      cleaningInterval: parseOptionalInt(env.DB_CLEANING_INTERVAL),
      daysBeforeNowToDelete: parseOptionalInt(env.DB_DAYS_BEFORE_NOW_TO_DELETE),
    },
    apiOptions: {
      port: parseOptionalInt(env.API_PORT),
    },
  };
}
