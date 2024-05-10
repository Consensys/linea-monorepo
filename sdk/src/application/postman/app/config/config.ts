import { LoggerOptions } from "winston";
import { DBConfig, DBCleanerConfig } from "../../persistence/config/types";

/**
 * Configuration for the Postman service, including network configurations, database options, and logging.
 */
export type PostmanConfig = {
  l1Config: L1NetworkConfig;
  l2Config: L2NetworkConfig;
  l1L2AutoClaimEnabled: boolean;
  l2L1AutoClaimEnabled: boolean;
  databaseOptions: DBConfig;
  databaseCleanerConfig?: DBCleanerConfig;
  loggerOptions?: LoggerOptions;
};

/**
 * Base configuration for a network, including claiming, listener settings, and contract details.
 */
type NetworkConfig = {
  claiming: ClaimingConfig;
  listener: ListenerConfig;
  rpcUrl: string;
  messageServiceContractAddress: string;
  isEOAEnabled?: boolean;
  isCalldataEnabled?: boolean;
};

/**
 * Configuration specific to the L1 network, extending the base NetworkConfig.
 */
export type L1NetworkConfig = NetworkConfig;

/**
 * Configuration specific to the L2 network, extending the base NetworkConfig with additional options.
 */
export type L2NetworkConfig = NetworkConfig & { l2MessageTreeDepth?: number };

/**
 * Configuration for claiming operations, including signer details, fee settings, and retry policies.
 */
export type ClaimingConfig = {
  signerPrivateKey: string;
  messageSubmissionTimeout: number;
  feeRecipientAddress?: string;
  maxNonceDiff?: number;
  maxFeePerGas?: bigint;
  gasEstimationPercentile?: number;
  isMaxGasFeeEnforced?: boolean;
  profitMargin?: number;
  maxNumberOfRetries?: number;
  retryDelayInSeconds?: number;
  maxClaimGasLimit?: number;
  maxTxRetries?: number;
};

/**
 * Configuration for the event listener, including polling settings and block fetching limits.
 */
export type ListenerConfig = {
  pollingInterval?: number;
  initialFromBlock?: number;
  blockConfirmation?: number;
  maxFetchMessagesFromDb?: number;
  maxBlocksToFetchLogs?: number;
};
