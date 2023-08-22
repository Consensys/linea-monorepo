import { format, transports } from "winston";
import * as dotenv from "dotenv";
import { PostmanServiceClient } from "../src/lib/postman/PostmanServiceClient";

dotenv.config();

async function main() {
  const client = new PostmanServiceClient({
    l1Config: {
      rpcUrl: process.env.L1_RPC_URL ?? "",
      messageServiceContractAddress: process.env.L1_CONTRACT_ADDRESS ?? "",
      onlyEOATarget: process.env.ONLY_EOA_TARGET === "true",
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
        maxFeePerGas: parseInt(process.env.MAX_FEE_PER_GAS ?? "100000000000"),
        gasEstimationPercentile: parseInt(process.env.GAS_ESTIMATION_PERCENTILE ?? "50"),
        profitMargin: parseFloat(process.env.PROFIT_MARGIN ?? "1.0"),
        maxNumberOfRetries: parseInt(process.env.MAX_NUMBER_OF_RETRIES ?? "100"),
        retryDelayInSeconds: parseInt(process.env.RETRY_DELAY_IN_SECONDS ?? "30"),
      },
    },
    l2Config: {
      rpcUrl: process.env.L2_RPC_URL ?? "",
      messageServiceContractAddress: process.env.L2_CONTRACT_ADDRESS ?? "",
      onlyEOATarget: process.env.ONLY_EOA_TARGET === "true",
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
        maxFeePerGas: parseInt(process.env.MAX_FEE_PER_GAS ?? "100000000000"),
        gasEstimationPercentile: parseInt(process.env.GAS_ESTIMATION_PERCENTILE ?? "50"),
        profitMargin: parseFloat(process.env.PROFIT_MARGIN ?? "1.0"),
        maxNumberOfRetries: parseInt(process.env.MAX_NUMBER_OF_RETRIES ?? "100"),
        retryDelayInSeconds: parseInt(process.env.RETRY_DELAY_IN_SECONDS ?? "30"),
        maxClaimGasLimit: parseInt(process.env.MAX_CLAIM_GAS_LIMIT ?? "100000"),
      },
    },
    loggerOptions: {
      level: "info",
      format: format.combine(
        format.colorize({ all: true }),
        format.timestamp({
          format: "YYYY-MM-DD hh:mm:ss.SSS A",
        }),
        format.align(),
        format.printf(({ module, timestamp, level, message }) => `[${timestamp}] ${module}: ${level} ${message}`),
      ),
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
