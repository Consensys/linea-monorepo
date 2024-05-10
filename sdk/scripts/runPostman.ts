import * as dotenv from "dotenv";
import { transports } from "winston";
import { PostmanServiceClient } from "../src/application/postman/app/PostmanServiceClient";

dotenv.config();

async function main() {
  const client = new PostmanServiceClient({
    l1Config: {
      rpcUrl: process.env.L1_RPC_URL ?? "",
      messageServiceContractAddress: process.env.L1_CONTRACT_ADDRESS ?? "",
      isEOAEnabled: process.env.L1_L2_EOA_ENABLED === "true",
      isCalldataEnabled: process.env.L1_L2_CALLDATA_ENABLED === "true",
      listener: {
        pollingInterval: parseInt(process.env.L1_LISTENER_INTERVAL ?? "4000"),
        maxFetchMessagesFromDb: parseInt(process.env.MAX_FETCH_MESSAGES_FROM_DB ?? "1000"),
        maxBlocksToFetchLogs: parseInt(process.env.L1_MAX_BLOCKS_TO_FETCH_LOGS ?? "1000"),
        ...(parseInt(process.env.L1_LISTENER_INITIAL_FROM_BLOCK ?? "") >= 0
          ? { initialFromBlock: parseInt(process.env.L1_LISTENER_INITIAL_FROM_BLOCK ?? "") }
          : {}),
        ...(parseInt(process.env.L1_LISTENER_BLOCK_CONFIRMATION ?? "") >= 0
          ? { blockConfirmation: parseInt(process.env.L1_LISTENER_BLOCK_CONFIRMATION ?? "") }
          : {}),
      },
      claiming: {
        signerPrivateKey: process.env.L1_SIGNER_PRIVATE_KEY ?? "",
        messageSubmissionTimeout: parseInt(process.env.MESSAGE_SUBMISSION_TIMEOUT ?? "300000"),
        maxNonceDiff: parseInt(process.env.MAX_NONCE_DIFF ?? "10000"),
        maxFeePerGas: BigInt(process.env.MAX_FEE_PER_GAS ?? "100000000000"),
        gasEstimationPercentile: parseInt(process.env.GAS_ESTIMATION_PERCENTILE ?? "50"),
        profitMargin: parseFloat(process.env.PROFIT_MARGIN ?? "1.0"),
        maxNumberOfRetries: parseInt(process.env.MAX_NUMBER_OF_RETRIES ?? "100"),
        retryDelayInSeconds: parseInt(process.env.RETRY_DELAY_IN_SECONDS ?? "30"),
        maxClaimGasLimit: parseInt(process.env.MAX_CLAIM_GAS_LIMIT ?? "100000"),
        maxTxRetries: parseInt(process.env.MAX_TX_RETRIES ?? "20"),
        isMaxGasFeeEnforced: process.env.L1_MAX_GAS_FEE_ENFORCED === "true",
      },
    },
    l2Config: {
      rpcUrl: process.env.L2_RPC_URL ?? "",
      messageServiceContractAddress: process.env.L2_CONTRACT_ADDRESS ?? "",
      isEOAEnabled: process.env.L2_L1_EOA_ENABLED === "true",
      isCalldataEnabled: process.env.L2_L1_CALLDATA_ENABLED === "true",
      listener: {
        pollingInterval: parseInt(process.env.L2_LISTENER_INTERVAL ?? "4000"),
        maxFetchMessagesFromDb: parseInt(process.env.MAX_FETCH_MESSAGES_FROM_DB ?? "1000"),
        maxBlocksToFetchLogs: parseInt(process.env.L2_MAX_BLOCKS_TO_FETCH_LOGS ?? "1000"),
        ...(parseInt(process.env.L2_LISTENER_INITIAL_FROM_BLOCK ?? "") >= 0
          ? { initialFromBlock: parseInt(process.env.L2_LISTENER_INITIAL_FROM_BLOCK ?? "") }
          : {}),
        ...(parseInt(process.env.L2_LISTENER_BLOCK_CONFIRMATION ?? "") >= 0
          ? { blockConfirmation: parseInt(process.env.L2_LISTENER_BLOCK_CONFIRMATION ?? "") }
          : {}),
      },
      claiming: {
        signerPrivateKey: process.env.L2_SIGNER_PRIVATE_KEY ?? "",
        messageSubmissionTimeout: parseInt(process.env.MESSAGE_SUBMISSION_TIMEOUT ?? "300000"),
        maxNonceDiff: parseInt(process.env.MAX_NONCE_DIFF ?? "10000"),
        maxFeePerGas: BigInt(process.env.MAX_FEE_PER_GAS ?? "100000000000"),
        gasEstimationPercentile: parseInt(process.env.GAS_ESTIMATION_PERCENTILE ?? "50"),
        profitMargin: parseFloat(process.env.PROFIT_MARGIN ?? "1.0"),
        maxNumberOfRetries: parseInt(process.env.MAX_NUMBER_OF_RETRIES ?? "100"),
        retryDelayInSeconds: parseInt(process.env.RETRY_DELAY_IN_SECONDS ?? "30"),
        maxClaimGasLimit: parseInt(process.env.MAX_CLAIM_GAS_LIMIT ?? "100000"),
        maxTxRetries: parseInt(process.env.MAX_TX_RETRIES ?? "20"),
        isMaxGasFeeEnforced: process.env.L2_MAX_GAS_FEE_ENFORCED === "true",
      },
      l2MessageTreeDepth: parseInt(process.env.L2_MESSAGE_TREE_DEPTH ?? "5"),
    },
    l1L2AutoClaimEnabled: process.env.L1_L2_AUTO_CLAIM_ENABLED === "true",
    l2L1AutoClaimEnabled: process.env.L2_L1_AUTO_CLAIM_ENABLED === "true",
    loggerOptions: {
      level: "info",
      transports: [new transports.Console()],
    },
    databaseOptions: {
      type: "postgres",
      host: process.env.POSTGRES_HOST ?? "127.0.0.1",
      port: parseInt(process.env.POSTGRES_PORT ?? "5432"),
      username: process.env.POSTGRES_USER ?? "postgres",
      password: process.env.POSTGRES_PASSWORD ?? "postgres",
      database: process.env.POSTGRES_DB ?? "postman_db",
    },
    databaseCleanerConfig: {
      enabled: process.env.DB_CLEANER_ENABLED === "true",
      cleaningInterval: parseInt(process.env.DB_CLEANING_INTERVAL ?? "43200000"),
      daysBeforeNowToDelete: parseInt(process.env.DB_DAYS_BEFORE_NOW_TO_DELETE ?? "14"),
    },
  });
  await client.connectDatabase();
  client.startAllServices();
}

main()
  .then()
  .catch((error) => {
    console.error("", error);
    process.exit(1);
  });
