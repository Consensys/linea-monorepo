import { IApplication, ILogger, IMetricsService, WinstonLogger } from "@consensys/linea-shared-utils";
import { DataSource } from "typeorm";

import { PostmanConfig, PostmanOptions } from "./config/config";
import { getConfig } from "./config/utils";
import { L1ToL2App } from "./L1ToL2App";
import { L2ToL1App } from "./L2ToL1App";
import {
  IMessageMetricsUpdater,
  ISponsorshipMetricsUpdater,
  ITransactionMetricsUpdater,
  LineaPostmanMetrics,
} from "../../../core/metrics";
import { IPoller } from "../../../core/services/pollers/IPoller";
import { createPostmanApi } from "../../../infrastructure/api/PostmanApi";
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
  InMemoryNonceManager,
  createChainContext,
  ViemTransactionRetrier,
  ViemReceiptPoller,
} from "../../../infrastructure/blockchain/viem";
import { ViemErrorParser } from "../../../infrastructure/blockchain/viem";
import { MessageMetricsUpdater } from "../../../infrastructure/metrics/MessageMetricsUpdater";
import { PostmanMetricsService } from "../../../infrastructure/metrics/PostmanMetricsService";
import { SponsorshipMetricsUpdater } from "../../../infrastructure/metrics/SponsorshipMetricsUpdater";
import { TransactionMetricsUpdater } from "../../../infrastructure/metrics/TransactionMetricsUpdater";
import { DB } from "../../../infrastructure/persistence/dataSource";
import { TypeOrmMessageRepository } from "../../../infrastructure/persistence/repositories/TypeOrmMessageRepository";
import { MessageStatusSubscriber } from "../../../infrastructure/persistence/subscribers/MessageStatusSubscriber";
import { IntervalPoller } from "../../../services/pollers";
import { DatabaseCleanerProcessor } from "../../../services/processors/DatabaseCleanerProcessor";

export class PostmanApp {
  private readonly config: PostmanConfig;
  private readonly db: DataSource;
  private readonly logger: ILogger;
  private readonly postmanMetricsService: IMetricsService<LineaPostmanMetrics>;
  private readonly messageMetricsUpdater: IMessageMetricsUpdater;
  private readonly sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater;
  private readonly transactionMetricsUpdater: ITransactionMetricsUpdater;
  private api?: IApplication;
  private l1ToL2App?: L1ToL2App;
  private l2ToL1App?: L2ToL1App;
  private databaseCleaningPoller?: IPoller;

  constructor(options: PostmanOptions) {
    this.config = getConfig(options);
    this.logger = new WinstonLogger(PostmanApp.name, this.config.loggerOptions);
    this.logger.info("Postman configuration:\n  %s", this.toLogfmt(this.redactConfig(this.config)));

    this.db = DB.create(this.config.databaseOptions);
    this.postmanMetricsService = new PostmanMetricsService();
    this.messageMetricsUpdater = new MessageMetricsUpdater(this.db.manager, this.postmanMetricsService);
    this.sponsorshipMetricsUpdater = new SponsorshipMetricsUpdater(this.postmanMetricsService);
    this.transactionMetricsUpdater = new TransactionMetricsUpdater(this.postmanMetricsService);
  }

  public async start(): Promise<void> {
    const { l1Config, l2Config, loggerOptions } = this.config;

    const [l1, l2] = await Promise.all([
      createChainContext(l1Config.rpcUrl, l1Config.claiming.signer, this.logger),
      createChainContext(l2Config.rpcUrl, l2Config.claiming.signer, this.logger),
    ]);

    const l1GasProvider = new ViemEthereumGasProvider(l1.publicClient, {
      maxFeePerGasCap: l1Config.claiming.maxFeePerGasCap,
      gasEstimationPercentile: l1Config.claiming.gasEstimationPercentile,
      enforceMaxGasFee: l1Config.claiming.isMaxGasFeeEnforced,
    });
    const l2GasProvider = new ViemLineaGasProvider(l2.publicClient, {
      maxFeePerGasCap: l2Config.claiming.maxFeePerGasCap,
      enforceMaxGasFee: l2Config.claiming.isMaxGasFeeEnforced,
    });

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
      l2.account.address,
    );

    const messageRepository = new TypeOrmMessageRepository(this.db);

    const l1Provider = new ViemProvider(l1.publicClient);
    const l2Provider = new ViemProvider(l2.publicClient);
    const calldataDecoder = new ViemCalldataDecoder();
    const transactionSigner = new ViemTransactionSigner(l2.signer, l2.chainId);
    const errorParser = new ViemErrorParser();
    const sharedMetrics = {
      sponsorshipMetricsUpdater: this.sponsorshipMetricsUpdater,
      transactionMetricsUpdater: this.transactionMetricsUpdater,
    };

    if (this.config.l1L2AutoClaimEnabled) {
      const l2NonceManager = new InMemoryNonceManager(
        l2Provider,
        l2.account.address,
        l2Config.claiming.maxNonceDiff,
        new WinstonLogger("L2NonceManager", loggerOptions),
      );
      await l2NonceManager.initialize();

      const l2TransactionRetrier = new ViemTransactionRetrier(
        l2.publicClient,
        l2.walletClient,
        l2.account.address,
        l2Config.claiming.maxFeePerGasCap,
      );
      const l2ReceiptPoller = new ViemReceiptPoller(l2Provider);

      this.l1ToL2App = new L1ToL2App({
        l1LogClient: new ViemLineaRollupLogClient(l1.publicClient, l1Config.messageServiceContractAddress),
        l1Provider,
        l2MessageServiceClient,
        l2Provider: new ViemLineaProvider(l2.publicClient),
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
      });
    }

    if (this.config.l2L1AutoClaimEnabled) {
      const l1NonceManager = new InMemoryNonceManager(
        l1Provider,
        l1.account.address,
        l1Config.claiming.maxNonceDiff,
        new WinstonLogger("L1NonceManager", loggerOptions),
      );
      await l1NonceManager.initialize();

      const l1TransactionRetrier = new ViemTransactionRetrier(
        l1.publicClient,
        l1.walletClient,
        l1.account.address,
        l1Config.claiming.maxFeePerGasCap,
      );
      const l1ReceiptPoller = new ViemReceiptPoller(l1Provider);

      this.l2ToL1App = new L2ToL1App({
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
      });
    }

    if (this.config.databaseCleanerConfig.enabled) {
      const databaseCleanerProcessor = new DatabaseCleanerProcessor(
        messageRepository,
        { daysBeforeNowToDelete: this.config.databaseCleanerConfig.daysBeforeNowToDelete },
        new WinstonLogger(DatabaseCleanerProcessor.name, loggerOptions),
      );

      this.databaseCleaningPoller = new IntervalPoller(
        databaseCleanerProcessor,
        { pollingInterval: this.config.databaseCleanerConfig.cleaningInterval },
        new WinstonLogger("DatabaseCleaningPoller", loggerOptions),
      );
    }

    await this.db.initialize();
    this.logger.info("Database initialized.");

    await this.messageMetricsUpdater.initialize();
    this.db.subscribers.push(
      new MessageStatusSubscriber(this.messageMetricsUpdater, new WinstonLogger(MessageStatusSubscriber.name)),
    );
    this.api = createPostmanApi(
      this.config.apiConfig.port,
      this.postmanMetricsService,
      new WinstonLogger("ExpressApiApplication"),
    );

    this.l1ToL2App?.start();
    this.l2ToL1App?.start();
    this.databaseCleaningPoller?.start();
    this.api.start();
    this.logger.info("All services started.");
  }

  public async stop(): Promise<void> {
    this.l1ToL2App?.stop();
    this.l2ToL1App?.stop();
    this.databaseCleaningPoller?.stop();
    this.api?.stop();
    await this.db.destroy();
    this.logger.info("All services stopped.");
  }

  private redactConfig(config: PostmanConfig): Record<string, unknown> {
    const redactNetworkConfig = (networkConfig: PostmanConfig["l1Config"]) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { signer: _, ...claimingWithoutSigner } = networkConfig.claiming;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { rpcUrl: __, ...networkWithoutRpcUrl } = networkConfig;
      return {
        ...networkWithoutRpcUrl,
        claiming: claimingWithoutSigner,
      };
    };

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { databaseOptions: _, loggerOptions, ...rest } = config;
    return {
      ...rest,
      l1Config: redactNetworkConfig(config.l1Config),
      l2Config: redactNetworkConfig(config.l2Config),
      loggerOptions: { level: loggerOptions?.level },
    };
  }

  private toLogfmt(obj: Record<string, unknown>, prefix = ""): string {
    const pairs: string[] = [];
    for (const [key, value] of Object.entries(obj)) {
      const fullKey = prefix ? `${prefix}.${key}` : key;
      if (value !== null && typeof value === "object" && !Array.isArray(value)) {
        pairs.push(this.toLogfmt(value as Record<string, unknown>, fullKey));
      } else {
        const str = typeof value === "bigint" ? value.toString() : String(value ?? "");
        pairs.push(str.includes(" ") || str.includes('"') || str === "" ? `${fullKey}="${str}"` : `${fullKey}=${str}`);
      }
    }
    return pairs.join("\n  ");
  }
}
