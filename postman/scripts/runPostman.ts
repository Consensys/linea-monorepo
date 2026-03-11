import * as dotenv from "dotenv";
import { transports } from "winston";

import { SignerConfig } from "../src/application/postman/app/config/config";
import { PostmanApp } from "../src/application/postman/app/PostmanApp";

dotenv.config();

function buildSignerConfig(prefix: "L1" | "L2"): SignerConfig {
  const signerType = process.env[`${prefix}_SIGNER_TYPE`] ?? "private-key";

  if (signerType === "web3signer") {
    return {
      type: "web3signer",
      endpoint: process.env[`${prefix}_WEB3_SIGNER_ENDPOINT`] ?? "",
      publicKey: (process.env[`${prefix}_WEB3_SIGNER_PUBLIC_KEY`] ?? "0x") as `0x${string}`,
      ...(process.env[`${prefix}_WEB3_SIGNER_TLS_KEYSTORE_PATH`]
        ? {
            tls: {
              keyStorePath: process.env[`${prefix}_WEB3_SIGNER_TLS_KEYSTORE_PATH`]!,
              keyStorePassword: process.env[`${prefix}_WEB3_SIGNER_TLS_KEYSTORE_PASSWORD`] ?? "",
              trustStorePath: process.env[`${prefix}_WEB3_SIGNER_TLS_TRUSTSTORE_PATH`] ?? "",
              trustStorePassword: process.env[`${prefix}_WEB3_SIGNER_TLS_TRUSTSTORE_PASSWORD`] ?? "",
            },
          }
        : {}),
    };
  }

  return {
    type: "private-key",
    privateKey: (process.env[`${prefix}_SIGNER_PRIVATE_KEY`] ?? "0x") as `0x${string}`,
  };
}

async function main() {
  const client = new PostmanApp({
    l1Options: {
      rpcUrl: process.env.L1_RPC_URL ?? "",
      messageServiceContractAddress: (process.env.L1_CONTRACT_ADDRESS ?? "") as `0x${string}`,
      isEOAEnabled: process.env.L1_L2_EOA_ENABLED === "true",
      isCalldataEnabled: process.env.L1_L2_CALLDATA_ENABLED === "true",
      listener: {
        pollingInterval: process.env.L1_LISTENER_INTERVAL ? parseInt(process.env.L1_LISTENER_INTERVAL) : undefined,
        receiptPollingInterval: process.env.L1_LISTENER_RECEIPT_POLLING_INTERVAL
          ? parseInt(process.env.L1_LISTENER_RECEIPT_POLLING_INTERVAL)
          : undefined,
        maxFetchMessagesFromDb: process.env.MAX_FETCH_MESSAGES_FROM_DB
          ? parseInt(process.env.MAX_FETCH_MESSAGES_FROM_DB)
          : undefined,
        maxBlocksToFetchLogs: process.env.L1_MAX_BLOCKS_TO_FETCH_LOGS
          ? parseInt(process.env.L1_MAX_BLOCKS_TO_FETCH_LOGS)
          : undefined,
        ...(parseInt(process.env.L1_LISTENER_INITIAL_FROM_BLOCK ?? "") >= 0
          ? { initialFromBlock: parseInt(process.env.L1_LISTENER_INITIAL_FROM_BLOCK ?? "") }
          : {}),
        ...(parseInt(process.env.L1_LISTENER_BLOCK_CONFIRMATION ?? "") >= 0
          ? { blockConfirmation: parseInt(process.env.L1_LISTENER_BLOCK_CONFIRMATION ?? "") }
          : {}),
        ...(process.env.L1_EVENT_FILTER_FROM_ADDRESS ||
        process.env.L1_EVENT_FILTER_TO_ADDRESS ||
        (process.env.L1_EVENT_FILTER_CALLDATA && process.env.L1_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE)
          ? {
              eventFilters: {
                fromAddressFilter: process.env.L1_EVENT_FILTER_FROM_ADDRESS as `0x${string}` | undefined,
                toAddressFilter: process.env.L1_EVENT_FILTER_TO_ADDRESS as `0x${string}` | undefined,
                ...(process.env.L1_EVENT_FILTER_CALLDATA && process.env.L1_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE
                  ? {
                      calldataFilter: {
                        criteriaExpression: process.env.L1_EVENT_FILTER_CALLDATA,
                        calldataFunctionInterface: process.env.L1_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE,
                      },
                    }
                  : {}),
              },
            }
          : {}),
      },
      claiming: {
        signer: buildSignerConfig("L1"),
        messageSubmissionTimeout: process.env.MESSAGE_SUBMISSION_TIMEOUT
          ? parseInt(process.env.MESSAGE_SUBMISSION_TIMEOUT)
          : undefined,
        maxNonceDiff: process.env.MAX_NONCE_DIFF ? parseInt(process.env.MAX_NONCE_DIFF) : undefined,
        maxFeePerGasCap: process.env.MAX_FEE_PER_GAS_CAP ? BigInt(process.env.MAX_FEE_PER_GAS_CAP) : undefined,
        gasEstimationPercentile: process.env.GAS_ESTIMATION_PERCENTILE
          ? parseInt(process.env.GAS_ESTIMATION_PERCENTILE)
          : undefined,
        profitMargin: process.env.PROFIT_MARGIN ? parseFloat(process.env.PROFIT_MARGIN) : undefined,
        maxNumberOfRetries: process.env.MAX_NUMBER_OF_RETRIES ? parseInt(process.env.MAX_NUMBER_OF_RETRIES) : undefined,
        retryDelayInSeconds: process.env.RETRY_DELAY_IN_SECONDS
          ? parseInt(process.env.RETRY_DELAY_IN_SECONDS)
          : undefined,
        maxClaimGasLimit: process.env.MAX_CLAIM_GAS_LIMIT ? BigInt(process.env.MAX_CLAIM_GAS_LIMIT) : undefined,
        maxBumpsPerCycle: process.env.MAX_BUMPS_PER_CYCLE ? parseInt(process.env.MAX_BUMPS_PER_CYCLE) : undefined,
        maxRetryCycles: process.env.MAX_RETRY_CYCLES ? parseInt(process.env.MAX_RETRY_CYCLES) : undefined,
        isMaxGasFeeEnforced: process.env.L1_MAX_GAS_FEE_ENFORCED === "true",
        isPostmanSponsorshipEnabled: process.env.L2_L1_ENABLE_POSTMAN_SPONSORING === "true",
        maxPostmanSponsorGasLimit: process.env.MAX_POSTMAN_SPONSOR_GAS_LIMIT
          ? BigInt(process.env.MAX_POSTMAN_SPONSOR_GAS_LIMIT)
          : undefined,
        claimViaAddress: process.env.L1_CLAIM_VIA_ADDRESS as `0x${string}` | undefined,
      },
    },
    l2Options: {
      rpcUrl: process.env.L2_RPC_URL ?? "",
      messageServiceContractAddress: (process.env.L2_CONTRACT_ADDRESS ?? "") as `0x${string}`,
      isEOAEnabled: process.env.L2_L1_EOA_ENABLED === "true",
      isCalldataEnabled: process.env.L2_L1_CALLDATA_ENABLED === "true",
      listener: {
        pollingInterval: process.env.L2_LISTENER_INTERVAL ? parseInt(process.env.L2_LISTENER_INTERVAL) : undefined,
        receiptPollingInterval: process.env.L2_LISTENER_RECEIPT_POLLING_INTERVAL
          ? parseInt(process.env.L2_LISTENER_RECEIPT_POLLING_INTERVAL)
          : undefined,
        maxFetchMessagesFromDb: process.env.MAX_FETCH_MESSAGES_FROM_DB
          ? parseInt(process.env.MAX_FETCH_MESSAGES_FROM_DB)
          : undefined,
        maxBlocksToFetchLogs: process.env.L2_MAX_BLOCKS_TO_FETCH_LOGS
          ? parseInt(process.env.L2_MAX_BLOCKS_TO_FETCH_LOGS)
          : undefined,
        ...(parseInt(process.env.L2_LISTENER_INITIAL_FROM_BLOCK ?? "") >= 0
          ? { initialFromBlock: parseInt(process.env.L2_LISTENER_INITIAL_FROM_BLOCK ?? "") }
          : {}),
        ...(parseInt(process.env.L2_LISTENER_BLOCK_CONFIRMATION ?? "") >= 0
          ? { blockConfirmation: parseInt(process.env.L2_LISTENER_BLOCK_CONFIRMATION ?? "") }
          : {}),
        ...(process.env.L2_EVENT_FILTER_FROM_ADDRESS ||
        process.env.L2_EVENT_FILTER_TO_ADDRESS ||
        (process.env.L2_EVENT_FILTER_CALLDATA && process.env.L2_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE)
          ? {
              eventFilters: {
                fromAddressFilter: process.env.L2_EVENT_FILTER_FROM_ADDRESS as `0x${string}` | undefined,
                toAddressFilter: process.env.L2_EVENT_FILTER_TO_ADDRESS as `0x${string}` | undefined,
                ...(process.env.L2_EVENT_FILTER_CALLDATA && process.env.L2_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE
                  ? {
                      calldataFilter: {
                        criteriaExpression: process.env.L2_EVENT_FILTER_CALLDATA,
                        calldataFunctionInterface: process.env.L2_EVENT_FILTER_CALLDATA_FUNCTION_INTERFACE,
                      },
                    }
                  : {}),
              },
            }
          : {}),
      },
      claiming: {
        signer: buildSignerConfig("L2"),
        messageSubmissionTimeout: process.env.MESSAGE_SUBMISSION_TIMEOUT
          ? parseInt(process.env.MESSAGE_SUBMISSION_TIMEOUT)
          : undefined,
        maxNonceDiff: process.env.MAX_NONCE_DIFF ? parseInt(process.env.MAX_NONCE_DIFF) : undefined,
        maxFeePerGasCap: process.env.MAX_FEE_PER_GAS_CAP ? BigInt(process.env.MAX_FEE_PER_GAS_CAP) : undefined,
        gasEstimationPercentile: process.env.GAS_ESTIMATION_PERCENTILE
          ? parseInt(process.env.GAS_ESTIMATION_PERCENTILE)
          : undefined,
        profitMargin: process.env.PROFIT_MARGIN ? parseFloat(process.env.PROFIT_MARGIN) : undefined,
        maxNumberOfRetries: process.env.MAX_NUMBER_OF_RETRIES ? parseInt(process.env.MAX_NUMBER_OF_RETRIES) : undefined,
        retryDelayInSeconds: process.env.RETRY_DELAY_IN_SECONDS
          ? parseInt(process.env.RETRY_DELAY_IN_SECONDS)
          : undefined,
        maxClaimGasLimit: process.env.MAX_CLAIM_GAS_LIMIT ? BigInt(process.env.MAX_CLAIM_GAS_LIMIT) : undefined,
        maxBumpsPerCycle: process.env.MAX_BUMPS_PER_CYCLE ? parseInt(process.env.MAX_BUMPS_PER_CYCLE) : undefined,
        maxRetryCycles: process.env.MAX_RETRY_CYCLES ? parseInt(process.env.MAX_RETRY_CYCLES) : undefined,
        isMaxGasFeeEnforced: process.env.L2_MAX_GAS_FEE_ENFORCED === "true",
        isPostmanSponsorshipEnabled: process.env.L1_L2_ENABLE_POSTMAN_SPONSORING === "true",
        maxPostmanSponsorGasLimit: process.env.MAX_POSTMAN_SPONSOR_GAS_LIMIT
          ? BigInt(process.env.MAX_POSTMAN_SPONSOR_GAS_LIMIT)
          : undefined,
        claimViaAddress: process.env.L2_CLAIM_VIA_ADDRESS as `0x${string}` | undefined,
      },
      l2MessageTreeDepth: process.env.L2_MESSAGE_TREE_DEPTH ? parseInt(process.env.L2_MESSAGE_TREE_DEPTH) : undefined,
      enableLineaEstimateGas: process.env.ENABLE_LINEA_ESTIMATE_GAS === "true",
    },
    l1L2AutoClaimEnabled: process.env.L1_L2_AUTO_CLAIM_ENABLED === "true",
    l2L1AutoClaimEnabled: process.env.L2_L1_AUTO_CLAIM_ENABLED === "true",
    loggerOptions: {
      level: process.env.LOG_LEVEL ?? "info",
      transports: [new transports.Console()],
    },
    databaseOptions: {
      type: "postgres",
      host: process.env.POSTGRES_HOST ?? "127.0.0.1",
      port: parseInt(process.env.POSTGRES_PORT ?? "5432"),
      username: process.env.POSTGRES_USER ?? "postgres",
      password: process.env.POSTGRES_PASSWORD ?? "postgres",
      database: process.env.POSTGRES_DB ?? "postman_db",
      ...(process.env.POSTGRES_SSL === "true"
        ? {
            ssl: {
              rejectUnauthorized: process.env.POSTGRES_SSL_REJECT_UNAUTHORIZED === "true",
              ca: process.env.POSTGRES_SSL_CA_PATH ?? undefined,
            },
          }
        : {}),
    },
    databaseCleanerOptions: {
      enabled: process.env.DB_CLEANER_ENABLED === "true",
      cleaningInterval: process.env.DB_CLEANING_INTERVAL ? parseInt(process.env.DB_CLEANING_INTERVAL) : undefined,
      daysBeforeNowToDelete: process.env.DB_DAYS_BEFORE_NOW_TO_DELETE
        ? parseInt(process.env.DB_DAYS_BEFORE_NOW_TO_DELETE)
        : undefined,
    },
    apiOptions: {
      port: process.env.API_PORT ? parseInt(process.env.API_PORT) : undefined,
    },
  });
  await client.start();
}

main()
  .then()
  .catch((error) => {
    console.error("", error);
    process.exit(1);
  });

process.on("SIGINT", () => {
  process.exit(0);
});

process.on("SIGTERM", () => {
  process.exit(0);
});
