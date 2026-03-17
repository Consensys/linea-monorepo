import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { DataSource } from "typeorm";

import { PostmanConfig } from "./config/config";
import { L1ToL2App, L1ToL2Deps } from "./L1ToL2App";
import { L2ToL1App, L2ToL1Deps } from "./L2ToL1App";
import { Direction } from "../../../core/enums";
import { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "../../../core/metrics";
import { IPoller } from "../../../core/services/pollers/IPoller";
import {
  ViemCalldataDecoder,
  ViemEthereumGasProvider,
  ViemL2MessageServiceClient,
  ViemLineaGasProvider,
  ViemLineaProvider,
  ViemLineaRollupClient,
  ViemLineaRollupLogClient,
  ViemL2MessageServiceLogClient,
  ViemProvider,
  ViemTransactionSigner,
  NonceManager,
  createChainContext,
  ViemTransactionRetrier,
  ViemReceiptPoller,
} from "../../../infrastructure/blockchain/viem";
import { ViemErrorParser } from "../../../infrastructure/blockchain/viem";
import { TypeOrmMessageRepository } from "../../../infrastructure/persistence/repositories/TypeOrmMessageRepository";
import { IntervalPoller } from "../../../services/pollers";
import { DatabaseCleanerProcessor } from "../../../services/processors/DatabaseCleanerProcessor";

export type PostmanServices = {
  l1ToL2App?: L1ToL2App;
  l2ToL1App?: L2ToL1App;
  databaseCleaningPoller?: IPoller;
};

export async function buildPostmanServices(
  config: PostmanConfig,
  db: DataSource,
  sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater,
  transactionMetricsUpdater: ITransactionMetricsUpdater,
  logger: ILogger,
): Promise<PostmanServices> {
  const { l1Config, l2Config, loggerOptions } = config;

  const [l1, l2] = await Promise.all([
    createChainContext(l1Config.rpcUrl, l1Config.claiming.signer, logger),
    createChainContext(l2Config.rpcUrl, l2Config.claiming.signer, logger),
  ]);

  const l1SignerAddress = l1.account.address;
  const l2SignerAddress = l2.account.address;

  const l1GasProvider = new ViemEthereumGasProvider(
    l1.publicClient,
    {
      maxFeePerGasCap: l1Config.claiming.maxFeePerGasCap,
      gasEstimationPercentile: l1Config.claiming.gasEstimationPercentile,
      enforceMaxGasFee: l1Config.claiming.isMaxGasFeeEnforced,
    },
    new WinstonLogger("L1GasProvider", loggerOptions),
  );
  const l2GasProvider = new ViemLineaGasProvider(
    l2.publicClient,
    {
      maxFeePerGasCap: l2Config.claiming.maxFeePerGasCap,
      enforceMaxGasFee: l2Config.claiming.isMaxGasFeeEnforced,
    },
    new WinstonLogger("L2GasProvider", loggerOptions),
  );

  const lineaRollupClient = new ViemLineaRollupClient(
    l1.publicClient,
    l1.walletClient,
    l1Config.messageServiceContractAddress,
    l2.publicClient,
    l2Config.messageServiceContractAddress,
    l1GasProvider,
  );
  const l2MessageServiceClient = new ViemL2MessageServiceClient(
    l2.publicClient,
    l2.walletClient,
    l2Config.messageServiceContractAddress,
    l2GasProvider,
    l2SignerAddress,
  );

  const messageRepository = new TypeOrmMessageRepository(db);

  const l1Provider = new ViemProvider(l1.publicClient, new WinstonLogger("L1Provider", loggerOptions));
  const l2Provider = new ViemProvider(l2.publicClient, new WinstonLogger("L2Provider", loggerOptions));
  const calldataDecoder = new ViemCalldataDecoder();
  const errorParser = new ViemErrorParser();
  const sharedMetrics = { sponsorshipMetricsUpdater, transactionMetricsUpdater };

  const services: PostmanServices = {};

  if (config.l1L2AutoClaimEnabled) {
    const l2NonceManager = new NonceManager(
      l2Provider,
      { getMaxPendingNonce: () => messageRepository.getMaxPendingNonce(Direction.L1_TO_L2) },
      l2SignerAddress,
      l2Config.claiming.maxNonceDiff,
      new WinstonLogger("L2NonceManager", loggerOptions),
    );
    await l2NonceManager.initialize();

    const l2TransactionRetrier = new ViemTransactionRetrier(
      l2.publicClient,
      l2.walletClient,
      l2SignerAddress,
      l2Config.claiming.maxFeePerGasCap,
      new WinstonLogger("L2TransactionRetrier", loggerOptions),
    );
    const l2ReceiptPoller = new ViemReceiptPoller(l2Provider, new WinstonLogger("L2ReceiptPoller", loggerOptions));
    const transactionSigner = new ViemTransactionSigner(l2.signer, l2.chainId);

    const deps: L1ToL2Deps = {
      l1LogClient: new ViemLineaRollupLogClient(l1.publicClient, l1Config.messageServiceContractAddress),
      l1Provider,
      l2MessageServiceClient,
      l2Provider: new ViemLineaProvider(l2.publicClient, new WinstonLogger("L2LineaProvider", loggerOptions)),
      l2NonceManager,
      l2TransactionRetrier,
      l2ReceiptPoller,
      messageRepository,
      calldataDecoder,
      transactionSigner,
      errorParser,
      l1Config,
      l2Config,
      loggerOptions,
      ...sharedMetrics,
    };

    services.l1ToL2App = new L1ToL2App(deps);
  }

  if (config.l2L1AutoClaimEnabled) {
    const l1NonceManager = new NonceManager(
      l1Provider,
      { getMaxPendingNonce: () => messageRepository.getMaxPendingNonce(Direction.L2_TO_L1) },
      l1SignerAddress,
      l1Config.claiming.maxNonceDiff,
      new WinstonLogger("L1NonceManager", loggerOptions),
    );
    await l1NonceManager.initialize();

    const l1TransactionRetrier = new ViemTransactionRetrier(
      l1.publicClient,
      l1.walletClient,
      l1SignerAddress,
      l1Config.claiming.maxFeePerGasCap,
      new WinstonLogger("L1TransactionRetrier", loggerOptions),
    );
    const l1ReceiptPoller = new ViemReceiptPoller(l1Provider, new WinstonLogger("L1ReceiptPoller", loggerOptions));

    const deps: L2ToL1Deps = {
      l2LogClient: new ViemL2MessageServiceLogClient(l2.publicClient, l2Config.messageServiceContractAddress),
      l2Provider,
      lineaRollupClient,
      l1Provider,
      l1NonceManager,
      l1TransactionRetrier,
      l1ReceiptPoller,
      messageRepository,
      l1GasProvider,
      calldataDecoder,
      errorParser,
      l1Config,
      l2Config,
      loggerOptions,
      ...sharedMetrics,
    };

    services.l2ToL1App = new L2ToL1App(deps);
  }

  if (config.databaseCleanerConfig.enabled) {
    const databaseCleanerProcessor = new DatabaseCleanerProcessor(
      messageRepository,
      { daysBeforeNowToDelete: config.databaseCleanerConfig.daysBeforeNowToDelete },
      new WinstonLogger(DatabaseCleanerProcessor.name, loggerOptions),
    );

    services.databaseCleaningPoller = new IntervalPoller(
      databaseCleanerProcessor,
      { pollingInterval: config.databaseCleanerConfig.cleaningInterval },
      new WinstonLogger("DatabaseCleaningPoller", loggerOptions),
    );
  }

  return services;
}
