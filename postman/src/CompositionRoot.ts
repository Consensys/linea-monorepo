import {
  createPublicClient,
  createWalletClient,
  defineChain,
  http,
  type Address,
  type Chain,
  type Hex,
  type PublicClient,
  type WalletClient,
} from "viem";
import { privateKeyToAccount } from "viem/accounts";

import { MessageSentEventPoller } from "./application/pollers/MessageSentEventPoller";
import { Poller } from "./application/pollers/Poller";
import { NonceCoordinator } from "./application/services/NonceCoordinator";
import { AnchorMessages } from "./application/use-cases/AnchorMessages";
import { ClaimMessages } from "./application/use-cases/ClaimMessages";
import { CleanDatabase } from "./application/use-cases/CleanDatabase";
import { ComputeTransactionSize } from "./application/use-cases/ComputeTransactionSize";
import { MonitorClaimReceipts } from "./application/use-cases/MonitorClaimReceipts";
import { ProcessMessageSentEvents } from "./application/use-cases/ProcessMessageSentEvents";
import { RetryStuckClaims } from "./application/use-cases/RetryStuckClaims";
import { MessageEventFilter } from "./domain/services/MessageEventFilter";
import { Direction } from "./domain/types/enums";
import { EthereumTransactionValidator } from "./infrastructure/blockchain/l1/EthereumTransactionValidator";
import { ViemEthereumGasProvider } from "./infrastructure/blockchain/l1/ViemEthereumGasProvider";
import { ViemL1ContractClient } from "./infrastructure/blockchain/l1/ViemL1ContractClient";
import { ViemL1LogClient } from "./infrastructure/blockchain/l1/ViemL1LogClient";
import { ViemProvider } from "./infrastructure/blockchain/l1/ViemProvider";
import { LineaTransactionValidator } from "./infrastructure/blockchain/l2/LineaTransactionValidator";
import { ViemL2ContractClient } from "./infrastructure/blockchain/l2/ViemL2ContractClient";
import { ViemL2LogClient } from "./infrastructure/blockchain/l2/ViemL2LogClient";
import { ViemLineaGasProvider } from "./infrastructure/blockchain/l2/ViemLineaGasProvider";
import { ViemLineaProvider } from "./infrastructure/blockchain/l2/ViemLineaProvider";
import { ViemTransactionSizeCalculator } from "./infrastructure/blockchain/l2/ViemTransactionSizeCalculator";
import { ViemCalldataDecoder } from "./infrastructure/blockchain/shared/ViemCalldataDecoder";
import { ViemErrorParser } from "./infrastructure/blockchain/shared/ViemErrorParser";
import { ViemNonceManager } from "./infrastructure/blockchain/shared/ViemNonceManager";

import type { PostmanConfig } from "./application/config/PostmanConfig";
import type { IPoller } from "./application/pollers/Poller";
import type { ILogger } from "./domain/ports/ILogger";
import type { IMessageRepository } from "./domain/ports/IMessageRepository";
import type { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "./domain/ports/IMetrics";

export type ViemClients = {
  l1PublicClient: PublicClient;
  l2PublicClient: PublicClient;
  l1WalletClient: WalletClient;
  l2WalletClient: WalletClient;
  l1SignerAddress: Address;
  l2SignerAddress: Address;
  l1Chain: Chain;
  l2Chain: Chain;
};

export type MetricsUpdaters = {
  sponsorship: ISponsorshipMetricsUpdater;
  transaction: ITransactionMetricsUpdater;
};

export type LoggerFactory = (name: string) => ILogger;

export async function createViemClients(config: PostmanConfig): Promise<ViemClients> {
  const tempL1 = createPublicClient({ transport: http(config.l1Config.rpcUrl) });
  const tempL2 = createPublicClient({ transport: http(config.l2Config.rpcUrl) });

  const [l1ChainId, l2ChainId] = await Promise.all([tempL1.getChainId(), tempL2.getChainId()]);

  const l1Chain = defineChain({
    id: l1ChainId,
    name: `l1-${l1ChainId}`,
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: [config.l1Config.rpcUrl] } },
  });

  const l2Chain = defineChain({
    id: l2ChainId,
    name: `l2-${l2ChainId}`,
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: [config.l2Config.rpcUrl] } },
  });

  const l1Account = privateKeyToAccount(config.l1Config.claiming.signerPrivateKey as Hex);
  const l2Account = privateKeyToAccount(config.l2Config.claiming.signerPrivateKey as Hex);

  const l1PublicClient = createPublicClient({ chain: l1Chain, transport: http(config.l1Config.rpcUrl) });
  const l2PublicClient = createPublicClient({ chain: l2Chain, transport: http(config.l2Config.rpcUrl) });

  const l1WalletClient = createWalletClient({
    chain: l1Chain,
    transport: http(config.l1Config.rpcUrl),
    account: l1Account,
  });

  const l2WalletClient = createWalletClient({
    chain: l2Chain,
    transport: http(config.l2Config.rpcUrl),
    account: l2Account,
  });

  return {
    l1PublicClient,
    l2PublicClient,
    l1WalletClient,
    l2WalletClient,
    l1SignerAddress: l1Account.address,
    l2SignerAddress: l2Account.address,
    l1Chain,
    l2Chain,
  };
}

export function createL1ToL2Flow(
  clients: ViemClients,
  repository: IMessageRepository,
  metricsUpdaters: MetricsUpdaters,
  config: PostmanConfig,
  loggerFactory: LoggerFactory,
): IPoller[] {
  const errorParser = new ViemErrorParser();

  const l1Provider = new ViemProvider(clients.l1PublicClient);
  const l2Provider = new ViemLineaProvider(clients.l2PublicClient);

  const l2GasProvider = config.l2Config.enableLineaEstimateGas
    ? new ViemLineaGasProvider(
        clients.l2PublicClient,
        clients.l2SignerAddress,
        config.l2Config.messageServiceContractAddress as Address,
        {
          enforceMaxGasFee: config.l2Config.claiming.isMaxGasFeeEnforced,
          maxFeePerGasCap: config.l2Config.claiming.maxFeePerGasCap,
        },
      )
    : new ViemEthereumGasProvider(clients.l2PublicClient, {
        gasEstimationPercentile: config.l2Config.claiming.gasEstimationPercentile,
        maxFeePerGasCap: config.l2Config.claiming.maxFeePerGasCap,
        enforceMaxGasFee: config.l2Config.claiming.isMaxGasFeeEnforced,
      });

  const l1LogClient = new ViemL1LogClient(
    clients.l1PublicClient,
    config.l1Config.messageServiceContractAddress as Address,
  );

  const l2ContractClient = new ViemL2ContractClient(
    clients.l2PublicClient,
    clients.l2WalletClient,
    config.l2Config.messageServiceContractAddress as Address,
    l2GasProvider as ViemLineaGasProvider,
  );

  const l1CalldataDecoder = config.l1Config.listener.eventFilters?.calldataFilter?.calldataFunctionInterface
    ? new ViemCalldataDecoder([config.l1Config.listener.eventFilters.calldataFilter.calldataFunctionInterface])
    : null;

  const l1MessageFilter = new MessageEventFilter(
    l1CalldataDecoder,
    { isEOAEnabled: config.l1Config.isEOAEnabled, isCalldataEnabled: config.l1Config.isCalldataEnabled },
    loggerFactory("L1MessageEventFilter"),
  );

  const l2NonceCoordinator = new NonceCoordinator(
    repository,
    new ViemNonceManager(clients.l2PublicClient, clients.l2SignerAddress),
    config.l2Config.claiming.maxNonceDiff,
    loggerFactory("L2NonceCoordinator"),
  );

  const l2TransactionValidator = new LineaTransactionValidator(
    {
      profitMargin: config.l2Config.claiming.profitMargin,
      maxClaimGasLimit: config.l2Config.claiming.maxClaimGasLimit,
      isPostmanSponsorshipEnabled: config.l2Config.claiming.isPostmanSponsorshipEnabled,
      maxPostmanSponsorGasLimit: config.l2Config.claiming.maxPostmanSponsorGasLimit,
    },
    l2Provider,
    l2ContractClient,
    l2ContractClient,
    loggerFactory("L2TransactionValidator"),
  );

  const transactionSizeCalculator = new ViemTransactionSizeCalculator(l2ContractClient);

  const processMessageSentEvents = new ProcessMessageSentEvents(
    repository,
    l1LogClient,
    l1Provider,
    l1MessageFilter,
    {
      direction: Direction.L1_TO_L2,
      maxBlocksToFetchLogs: config.l1Config.listener.maxBlocksToFetchLogs,
      blockConfirmation: config.l1Config.listener.blockConfirmation,
      eventFilters: config.l1Config.listener.eventFilters,
    },
    loggerFactory("L1ProcessMessageSentEvents"),
  );

  const anchorMessages = new AnchorMessages(
    l2ContractClient,
    repository,
    errorParser,
    {
      direction: Direction.L1_TO_L2,
      maxFetchMessagesFromDb: config.l1Config.listener.maxFetchMessagesFromDb,
      originContractAddress: config.l1Config.messageServiceContractAddress,
    },
    loggerFactory("L2AnchorMessages"),
  );

  const computeTransactionSize = new ComputeTransactionSize(
    repository,
    l2ContractClient,
    transactionSizeCalculator,
    errorParser,
    {
      direction: Direction.L1_TO_L2,
      originContractAddress: config.l1Config.messageServiceContractAddress,
    },
    loggerFactory("L2ComputeTransactionSize"),
  );

  const claimMessages = new ClaimMessages(
    l2ContractClient,
    l2ContractClient,
    l2NonceCoordinator,
    repository,
    l2TransactionValidator,
    errorParser,
    {
      direction: Direction.L1_TO_L2,
      originContractAddress: config.l1Config.messageServiceContractAddress,
      feeRecipientAddress: config.l2Config.claiming.feeRecipientAddress,
      profitMargin: config.l2Config.claiming.profitMargin,
      maxNumberOfRetries: config.l2Config.claiming.maxNumberOfRetries,
      retryDelayInSeconds: config.l2Config.claiming.retryDelayInSeconds,
      maxClaimGasLimit: config.l2Config.claiming.maxClaimGasLimit,
      claimViaAddress: config.l2Config.claiming.claimViaAddress,
    },
    loggerFactory("L2ClaimMessages"),
  );

  const retryStuckClaims = new RetryStuckClaims(
    l2ContractClient,
    l2ContractClient,
    l2Provider,
    repository,
    errorParser,
    loggerFactory("L2RetryStuckClaims"),
    config.l2Config.claiming.maxTxRetries,
  );

  const monitorClaimReceipts = new MonitorClaimReceipts(
    repository,
    l2ContractClient,
    metricsUpdaters.sponsorship,
    metricsUpdaters.transaction,
    l2Provider,
    retryStuckClaims,
    errorParser,
    {
      direction: Direction.L1_TO_L2,
      messageSubmissionTimeout: config.l2Config.claiming.messageSubmissionTimeout,
    },
    loggerFactory("L2MonitorClaimReceipts"),
  );

  const messageSentEventPoller = new MessageSentEventPoller(
    processMessageSentEvents,
    l1Provider,
    repository,
    {
      direction: Direction.L1_TO_L2,
      pollingInterval: config.l1Config.listener.pollingInterval,
      initialFromBlock: config.l1Config.listener.initialFromBlock,
      originContractAddress: config.l1Config.messageServiceContractAddress,
    },
    loggerFactory("L1MessageSentEventPoller"),
  );

  const anchoringPoller = new Poller(
    "L2MessageAnchoringPoller",
    () => anchorMessages.process(),
    config.l2Config.listener.pollingInterval,
    loggerFactory("L2MessageAnchoringPoller"),
  );

  const transactionSizePoller = new Poller(
    "L2ClaimTransactionSizePoller",
    () => computeTransactionSize.process(),
    config.l2Config.listener.pollingInterval,
    loggerFactory("L2ClaimTransactionSizePoller"),
  );

  const claimingPoller = new Poller(
    "L2MessageClaimingPoller",
    () => claimMessages.process(),
    config.l2Config.listener.pollingInterval,
    loggerFactory("L2MessageClaimingPoller"),
  );

  const persistingPoller = new Poller(
    "L2MessagePersistingPoller",
    () => monitorClaimReceipts.process(),
    config.l2Config.listener.pollingInterval,
    loggerFactory("L2MessagePersistingPoller"),
  );

  return [messageSentEventPoller, anchoringPoller, transactionSizePoller, claimingPoller, persistingPoller];
}

export function createL2ToL1Flow(
  clients: ViemClients,
  repository: IMessageRepository,
  metricsUpdaters: MetricsUpdaters,
  config: PostmanConfig,
  loggerFactory: LoggerFactory,
): IPoller[] {
  const errorParser = new ViemErrorParser();

  const l1Provider = new ViemProvider(clients.l1PublicClient);
  const l2Provider = new ViemLineaProvider(clients.l2PublicClient);

  const l1GasProvider = new ViemEthereumGasProvider(clients.l1PublicClient, {
    gasEstimationPercentile: config.l1Config.claiming.gasEstimationPercentile,
    maxFeePerGasCap: config.l1Config.claiming.maxFeePerGasCap,
    enforceMaxGasFee: config.l1Config.claiming.isMaxGasFeeEnforced,
  });

  const l2LogClient = new ViemL2LogClient(
    clients.l2PublicClient,
    config.l2Config.messageServiceContractAddress as Address,
  );

  const l1ContractClient = new ViemL1ContractClient(
    clients.l1PublicClient,
    clients.l1WalletClient,
    clients.l2PublicClient,
    config.l1Config.messageServiceContractAddress as Address,
    config.l2Config.messageServiceContractAddress as Address,
    l1GasProvider,
  );

  const l2CalldataDecoder = config.l2Config.listener.eventFilters?.calldataFilter?.calldataFunctionInterface
    ? new ViemCalldataDecoder([config.l2Config.listener.eventFilters.calldataFilter.calldataFunctionInterface])
    : null;

  const l2MessageFilter = new MessageEventFilter(
    l2CalldataDecoder,
    { isEOAEnabled: config.l2Config.isEOAEnabled, isCalldataEnabled: config.l2Config.isCalldataEnabled },
    loggerFactory("L2MessageEventFilter"),
  );

  const l1NonceCoordinator = new NonceCoordinator(
    repository,
    new ViemNonceManager(clients.l1PublicClient, clients.l1SignerAddress),
    config.l1Config.claiming.maxNonceDiff,
    loggerFactory("L1NonceCoordinator"),
  );

  const l1TransactionValidator = new EthereumTransactionValidator(
    l1ContractClient,
    l1ContractClient,
    l1GasProvider,
    {
      profitMargin: config.l1Config.claiming.profitMargin,
      maxClaimGasLimit: config.l1Config.claiming.maxClaimGasLimit,
      isPostmanSponsorshipEnabled: config.l1Config.claiming.isPostmanSponsorshipEnabled,
      maxPostmanSponsorGasLimit: config.l1Config.claiming.maxPostmanSponsorGasLimit,
    },
    loggerFactory("L1TransactionValidator"),
  );

  const processMessageSentEvents = new ProcessMessageSentEvents(
    repository,
    l2LogClient,
    l2Provider,
    l2MessageFilter,
    {
      direction: Direction.L2_TO_L1,
      maxBlocksToFetchLogs: config.l2Config.listener.maxBlocksToFetchLogs,
      blockConfirmation: config.l2Config.listener.blockConfirmation,
      eventFilters: config.l2Config.listener.eventFilters,
    },
    loggerFactory("L2ProcessMessageSentEvents"),
  );

  const anchorMessages = new AnchorMessages(
    l1ContractClient,
    repository,
    errorParser,
    {
      direction: Direction.L2_TO_L1,
      maxFetchMessagesFromDb: config.l1Config.listener.maxFetchMessagesFromDb,
      originContractAddress: config.l2Config.messageServiceContractAddress,
    },
    loggerFactory("L1AnchorMessages"),
  );

  const claimMessages = new ClaimMessages(
    l1ContractClient,
    l1ContractClient,
    l1NonceCoordinator,
    repository,
    l1TransactionValidator,
    errorParser,
    {
      direction: Direction.L2_TO_L1,
      originContractAddress: config.l2Config.messageServiceContractAddress,
      feeRecipientAddress: config.l1Config.claiming.feeRecipientAddress,
      profitMargin: config.l1Config.claiming.profitMargin,
      maxNumberOfRetries: config.l1Config.claiming.maxNumberOfRetries,
      retryDelayInSeconds: config.l1Config.claiming.retryDelayInSeconds,
      maxClaimGasLimit: config.l1Config.claiming.maxClaimGasLimit,
      claimViaAddress: config.l1Config.claiming.claimViaAddress,
    },
    loggerFactory("L1ClaimMessages"),
    l1GasProvider,
  );

  const retryStuckClaims = new RetryStuckClaims(
    l1ContractClient,
    l1ContractClient,
    l1Provider,
    repository,
    errorParser,
    loggerFactory("L1RetryStuckClaims"),
    config.l1Config.claiming.maxTxRetries,
  );

  const monitorClaimReceipts = new MonitorClaimReceipts(
    repository,
    l1ContractClient,
    metricsUpdaters.sponsorship,
    metricsUpdaters.transaction,
    l1Provider,
    retryStuckClaims,
    errorParser,
    {
      direction: Direction.L2_TO_L1,
      messageSubmissionTimeout: config.l1Config.claiming.messageSubmissionTimeout,
    },
    loggerFactory("L1MonitorClaimReceipts"),
  );

  const messageSentEventPoller = new MessageSentEventPoller(
    processMessageSentEvents,
    l2Provider,
    repository,
    {
      direction: Direction.L2_TO_L1,
      pollingInterval: config.l2Config.listener.pollingInterval,
      initialFromBlock: config.l2Config.listener.initialFromBlock,
      originContractAddress: config.l2Config.messageServiceContractAddress,
    },
    loggerFactory("L2MessageSentEventPoller"),
  );

  const anchoringPoller = new Poller(
    "L1MessageAnchoringPoller",
    () => anchorMessages.process(),
    config.l1Config.listener.pollingInterval,
    loggerFactory("L1MessageAnchoringPoller"),
  );

  const claimingPoller = new Poller(
    "L1MessageClaimingPoller",
    () => claimMessages.process(),
    config.l1Config.listener.pollingInterval,
    loggerFactory("L1MessageClaimingPoller"),
  );

  const persistingPoller = new Poller(
    "L1MessagePersistingPoller",
    () => monitorClaimReceipts.process(),
    config.l1Config.listener.receiptPollingInterval,
    loggerFactory("L1MessagePersistingPoller"),
  );

  return [messageSentEventPoller, anchoringPoller, claimingPoller, persistingPoller];
}

export function createDatabaseCleaningPoller(
  repository: IMessageRepository,
  config: PostmanConfig,
  loggerFactory: LoggerFactory,
): IPoller | null {
  if (!config.databaseCleanerConfig.enabled) {
    return null;
  }

  const msBeforeNowToDelete = config.databaseCleanerConfig.daysBeforeNowToDelete * 24 * 60 * 60 * 1000;
  const databaseCleaner = new CleanDatabase(repository, loggerFactory("DatabaseCleaner"));

  return new Poller(
    "DatabaseCleaningPoller",
    () => databaseCleaner.databaseCleanerRoutine(msBeforeNowToDelete),
    config.databaseCleanerConfig.cleaningInterval,
    loggerFactory("DatabaseCleaningPoller"),
  );
}
