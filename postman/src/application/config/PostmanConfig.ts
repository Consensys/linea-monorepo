import { Direction } from "../../domain/types/Direction";

type DeepRequired<T> = {
  [P in keyof T]-?: T[P] extends object ? DeepRequired<T[P]> : T[P];
};

export type PostmanOptions = {
  l1Options: L1NetworkOptions;
  l2Options: L2NetworkOptions;
  l1L2AutoClaimEnabled: boolean;
  l2L1AutoClaimEnabled: boolean;
  databaseOptions: unknown;
  databaseCleanerOptions?: DBCleanerOptions;
  loggerOptions?: unknown;
  apiOptions?: ApiOptions;
};

export type PostmanConfig = {
  l1Config: L1NetworkConfig;
  l2Config: L2NetworkConfig;
  l1L2AutoClaimEnabled: boolean;
  l2L1AutoClaimEnabled: boolean;
  databaseOptions: unknown;
  databaseCleanerConfig: DBCleanerConfig;
  loggerOptions?: unknown;
  apiConfig: ApiConfig;
};

type NetworkOptions = {
  claiming: ClaimingOptions;
  listener: ListenerOptions;
  rpcUrl: string;
  messageServiceContractAddress: string;
  isEOAEnabled?: boolean;
  isCalldataEnabled?: boolean;
};

type NetworkConfig = Omit<DeepRequired<NetworkOptions>, "claiming" | "listener"> & {
  claiming: ClaimingConfig;
  listener: ListenerConfig;
};

export type L1NetworkOptions = NetworkOptions;

export type L1NetworkConfig = NetworkConfig;

export type L2NetworkOptions = NetworkOptions & {
  l2MessageTreeDepth?: number;
  enableLineaEstimateGas?: boolean;
};

export type L2NetworkConfig = NetworkConfig & {
  l2MessageTreeDepth: number;
  enableLineaEstimateGas: boolean;
};

export type ClaimingOptions = {
  signerPrivateKey: string;
  messageSubmissionTimeout?: number;
  feeRecipientAddress?: string;
  maxNonceDiff?: number;
  maxFeePerGasCap?: bigint;
  gasEstimationPercentile?: number;
  isMaxGasFeeEnforced?: boolean;
  profitMargin?: number;
  maxNumberOfRetries?: number;
  retryDelayInSeconds?: number;
  maxClaimGasLimit?: bigint;
  maxTxRetries?: number;
  isPostmanSponsorshipEnabled?: boolean;
  maxPostmanSponsorGasLimit?: bigint;
  claimViaAddress?: string;
};

export type ClaimingConfig = Omit<Required<ClaimingOptions>, "feeRecipientAddress" | "claimViaAddress"> & {
  feeRecipientAddress?: string;
  claimViaAddress?: string;
};

export type ListenerOptions = {
  pollingInterval?: number;
  receiptPollingInterval?: number;
  initialFromBlock?: number;
  blockConfirmation?: number;
  maxFetchMessagesFromDb?: number;
  maxBlocksToFetchLogs?: number;
  eventFilters?: {
    fromAddressFilter?: string;
    toAddressFilter?: string;
    calldataFilter?: {
      criteriaExpression: string;
      calldataFunctionInterface: string;
    };
  };
};

export type ListenerConfig = Required<Omit<ListenerOptions, "eventFilters">> &
  Partial<Pick<ListenerOptions, "eventFilters">>;

export type ApiOptions = {
  port?: number;
};

export type ApiConfig = Required<ApiOptions>;

export type DBCleanerOptions = {
  enabled: boolean;
  cleaningInterval?: number;
  daysBeforeNowToDelete?: number;
};

export type DBCleanerConfig = Required<DBCleanerOptions>;

export type MessageSentEventProcessorConfig = {
  direction: Direction;
  maxBlocksToFetchLogs: number;
  blockConfirmation: number;
  isEOAEnabled: boolean;
  isCalldataEnabled: boolean;
  eventFilters?: {
    fromAddressFilter?: string;
    toAddressFilter?: string;
    calldataFilter?: {
      criteriaExpression: string;
      calldataFunctionInterface: string;
    };
  };
};

export type MessageAnchoringProcessorConfig = {
  maxFetchMessagesFromDb: number;
  originContractAddress: string;
};

export type MessageClaimingProcessorConfig = {
  maxNonceDiff: number;
  feeRecipientAddress?: string;
  profitMargin: number;
  maxNumberOfRetries: number;
  retryDelayInSeconds: number;
  maxClaimGasLimit: bigint;
  direction: Direction;
  originContractAddress: string;
  claimViaAddress?: string;
};

export type MessageClaimingPersisterConfig = {
  direction: Direction;
  messageSubmissionTimeout: number;
  maxTxRetries: number;
};

export type L2ClaimMessageTransactionSizeProcessorConfig = {
  direction: Direction;
  originContractAddress: string;
};
