import { IApplication, ILogger, IMetricsService } from "@consensys/linea-shared-utils";
import { DataSource } from "typeorm";
import { createPublicClient, createWalletClient, defineChain, http } from "viem";

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
  createSignerClient,
  contractSignerToViemAccount,
} from "../../../infrastructure/blockchain/viem";
import { MessageMetricsUpdater } from "../../../infrastructure/metrics/MessageMetricsUpdater";
import { PostmanMetricsService } from "../../../infrastructure/metrics/PostmanMetricsService";
import { SponsorshipMetricsUpdater } from "../../../infrastructure/metrics/SponsorshipMetricsUpdater";
import { TransactionMetricsUpdater } from "../../../infrastructure/metrics/TransactionMetricsUpdater";
import { DB } from "../../../infrastructure/persistence/dataSource";
import { TypeOrmMessageRepository } from "../../../infrastructure/persistence/repositories/TypeOrmMessageRepository";
import { MessageStatusSubscriber } from "../../../infrastructure/persistence/subscribers/MessageStatusSubscriber";
import { DatabaseCleaner } from "../../../services/persistence";
import { DatabaseCleaningPoller } from "../../../services/pollers";
import { ErrorParser } from "../../../utils/ErrorParser";
import { PostmanWinstonLogger } from "../../../utils/PostmanWinstonLogger";

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
    this.logger = new PostmanWinstonLogger(PostmanApp.name, this.config.loggerOptions);
    this.logger.info("Postman configuration:\n  %s", this.toLogfmt(this.redactConfig(this.config)));

    this.db = DB.create(this.config.databaseOptions);
    this.postmanMetricsService = new PostmanMetricsService();
    this.messageMetricsUpdater = new MessageMetricsUpdater(this.db.manager, this.postmanMetricsService);
    this.sponsorshipMetricsUpdater = new SponsorshipMetricsUpdater(this.postmanMetricsService);
    this.transactionMetricsUpdater = new TransactionMetricsUpdater(this.postmanMetricsService);
  }

  public async start(): Promise<void> {
    const { l1Config, l2Config, loggerOptions } = this.config;

    const l1PublicClient = createPublicClient({ transport: http(l1Config.rpcUrl) });
    const l2PublicClient = createPublicClient({ transport: http(l2Config.rpcUrl) });
    const [l1ChainId, l2ChainId] = await Promise.all([l1PublicClient.getChainId(), l2PublicClient.getChainId()]);

    const mkChain = (id: number, rpcUrl: string) =>
      defineChain({
        id,
        name: "custom",
        nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
        rpcUrls: { default: { http: [rpcUrl] } },
      });

    const l1Signer = createSignerClient(
      l1Config.claiming.signer,
      this.logger,
      l1Config.rpcUrl,
      mkChain(l1ChainId, l1Config.rpcUrl),
    );
    const l2Signer = createSignerClient(
      l2Config.claiming.signer,
      this.logger,
      l2Config.rpcUrl,
      mkChain(l2ChainId, l2Config.rpcUrl),
    );
    const l1Account = contractSignerToViemAccount(l1Signer);
    const l2Account = contractSignerToViemAccount(l2Signer);

    const l1WalletClient = createWalletClient({
      account: l1Account,
      transport: http(l1Config.rpcUrl),
      chain: mkChain(l1ChainId, l1Config.rpcUrl),
    });
    const l2WalletClient = createWalletClient({
      account: l2Account,
      transport: http(l2Config.rpcUrl),
      chain: mkChain(l2ChainId, l2Config.rpcUrl),
    });

    const l1GasProvider = new ViemEthereumGasProvider(l1PublicClient, {
      maxFeePerGasCap: l1Config.claiming.maxFeePerGasCap,
      gasEstimationPercentile: l1Config.claiming.gasEstimationPercentile,
      enforceMaxGasFee: l1Config.claiming.isMaxGasFeeEnforced,
    });
    const l2GasProvider = new ViemLineaGasProvider(l2PublicClient, {
      maxFeePerGasCap: l2Config.claiming.maxFeePerGasCap,
      enforceMaxGasFee: l2Config.claiming.isMaxGasFeeEnforced,
    });

    const lineaRollupClient = new ViemLineaRollupClient(
      l1PublicClient,
      l1WalletClient,
      l1Config.messageServiceContractAddress,
      l2PublicClient,
      l2Config.messageServiceContractAddress,
      l1GasProvider,
    );
    const l2MessageServiceClient = new ViemL2MessageServiceClient(
      l2PublicClient,
      l2WalletClient,
      l2Config.messageServiceContractAddress,
      l2GasProvider,
      l2Account.address,
    );

    const messageRepository = new TypeOrmMessageRepository(this.db);

    const calldataDecoder = new ViemCalldataDecoder();
    const transactionSigner = new ViemTransactionSigner(l2Signer, l2ChainId);
    const errorParser = new ErrorParser();
    const sharedMetrics = {
      sponsorshipMetricsUpdater: this.sponsorshipMetricsUpdater,
      transactionMetricsUpdater: this.transactionMetricsUpdater,
    };

    if (this.config.l1L2AutoClaimEnabled) {
      this.l1ToL2App = new L1ToL2App({
        l1LogClient: new ViemLineaRollupLogClient(l1PublicClient, l1Config.messageServiceContractAddress),
        l1Provider: new ViemProvider(l1PublicClient),
        l2MessageServiceClient,
        l2Provider: new ViemLineaProvider(l2PublicClient),
        l2PublicClient,
        l2SignerAddress: l2Account.address,
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
      this.l2ToL1App = new L2ToL1App({
        l2LogClient: new ViemL2MessageServiceLogClient(l2PublicClient, l2Config.messageServiceContractAddress),
        l2Provider: new ViemProvider(l2PublicClient),
        lineaRollupClient,
        l1Provider: new ViemProvider(l1PublicClient),
        l1PublicClient,
        l1SignerAddress: l1Account.address,
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

    const databaseCleaner = new DatabaseCleaner(
      messageRepository,
      new PostmanWinstonLogger(DatabaseCleaner.name, loggerOptions),
    );
    this.databaseCleaningPoller = new DatabaseCleaningPoller(
      databaseCleaner,
      new PostmanWinstonLogger(DatabaseCleaningPoller.name, loggerOptions),
      {
        enabled: this.config.databaseCleanerConfig.enabled,
        daysBeforeNowToDelete: this.config.databaseCleanerConfig.daysBeforeNowToDelete,
        cleaningInterval: this.config.databaseCleanerConfig.cleaningInterval,
      },
    );

    await this.db.initialize();
    this.logger.info("Database initialized.");

    await this.messageMetricsUpdater.initialize();
    this.db.subscribers.push(
      new MessageStatusSubscriber(this.messageMetricsUpdater, new PostmanWinstonLogger(MessageStatusSubscriber.name)),
    );
    this.api = createPostmanApi(
      this.config.apiConfig.port,
      this.postmanMetricsService,
      new PostmanWinstonLogger("ExpressApiApplication"),
    );

    this.l1ToL2App?.start();
    this.l2ToL1App?.start();
    this.databaseCleaningPoller.start();
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
