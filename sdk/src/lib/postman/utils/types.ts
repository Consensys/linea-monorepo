import { LoggerOptions } from "winston";
import { PostgresConnectionOptions } from "typeorm/driver/postgres/PostgresConnectionOptions";
import { BetterSqlite3ConnectionOptions } from "typeorm/driver/better-sqlite3/BetterSqlite3ConnectionOptions";
import { Direction, MessageStatus } from "./enums";

export type PostmanConfig = {
  l1Config: L1NetworkConfig;
  l2Config: L2NetworkConfig;
  databaseOptions: DBConfig;
  loggerOptions?: LoggerOptions;
};

export type DBConfig = PostgresConnectionOptions | BetterSqlite3ConnectionOptions;

type NetworkConfig = {
  claiming: ClaimingConfig;
  listener: ListenerConfig;
  rpcUrl: string;
  messageServiceContractAddress: string;
  onlyEOATarget?: boolean;
};

export type L1NetworkConfig = NetworkConfig;

export type L2NetworkConfig = NetworkConfig;

export type ClaimingConfig = {
  signerPrivateKey: string;
  messageSubmissionTimeout: number;
  feeRecipientAddress?: string;
  maxNonceDiff?: number;
  maxFeePerGas?: number;
  gasEstimationPercentile?: number;
  profitMargin?: number;
  maxNumberOfRetries?: number;
  retryDelayInSeconds?: number;
  maxClaimGasLimit?: number;
};

export type ListenerConfig = {
  pollingInterval?: number;
  initialFromBlock?: number;
  blockConfirmation?: number;
  maxFetchMessagesFromDb?: number;
  maxBlocksToFetchLogs?: number;
};

export type MessageInDb = {
  id?: number;
  messageSender: string;
  destination: string;
  fee: string;
  value: string;
  messageNonce: number;
  calldata: string;
  messageHash: string;
  messageContractAddress: string;
  sentBlockNumber: number;
  direction: Direction;
  status: MessageStatus;
  claimTxCreationDate?: Date;
  claimTxGasLimit?: number;
  claimTxMaxFeePerGas?: bigint;
  claimTxMaxPriorityFeePerGas?: bigint;
  claimTxNonce?: number;
  claimTxHash?: string;
  claimNumberOfRetry: number;
  claimLastRetriedAt?: Date;
  claimGasEstimationThreshold?: number;
  createdAt?: Date;
  updatedAt?: Date;
};
