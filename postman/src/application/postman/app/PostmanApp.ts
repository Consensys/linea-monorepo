import { IApplication, ILogger, IMetricsService, WinstonLogger } from "@consensys/linea-shared-utils";
import { DataSource } from "typeorm";

import { PostmanConfig, PostmanOptions } from "./config/config";
import { getConfig } from "./config/utils";
import { buildPostmanServices, PostmanServices } from "./PostmanContainer";
import {
  IMessageMetricsUpdater,
  ISponsorshipMetricsUpdater,
  ITransactionMetricsUpdater,
  LineaPostmanMetrics,
} from "../../../core/metrics";
import { createPostmanApi } from "../../../infrastructure/api/PostmanApi";
import { MessageMetricsUpdater } from "../../../infrastructure/metrics/MessageMetricsUpdater";
import { PostmanMetricsService } from "../../../infrastructure/metrics/PostmanMetricsService";
import { SponsorshipMetricsUpdater } from "../../../infrastructure/metrics/SponsorshipMetricsUpdater";
import { TransactionMetricsUpdater } from "../../../infrastructure/metrics/TransactionMetricsUpdater";
import { DB } from "../../../infrastructure/persistence/dataSource";
import { MessageStatusSubscriber } from "../../../infrastructure/persistence/subscribers/MessageStatusSubscriber";

export class PostmanApp {
  private readonly config: PostmanConfig;
  private readonly db: DataSource;
  private readonly logger: ILogger;
  private readonly postmanMetricsService: IMetricsService<LineaPostmanMetrics>;
  private readonly messageMetricsUpdater: IMessageMetricsUpdater;
  private readonly sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater;
  private readonly transactionMetricsUpdater: ITransactionMetricsUpdater;
  private api?: IApplication;
  private services?: PostmanServices;

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
    this.services = await buildPostmanServices(
      this.config,
      this.db,
      this.sponsorshipMetricsUpdater,
      this.transactionMetricsUpdater,
      this.logger,
    );

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

    this.services.l1ToL2App?.start();
    this.services.l2ToL1App?.start();
    this.services.databaseCleaningPoller?.start();
    this.api.start();
    this.logger.info("All services started.");
  }

  public async stop(): Promise<void> {
    this.services?.l1ToL2App?.stop();
    this.services?.l2ToL1App?.stop();
    this.services?.databaseCleaningPoller?.stop();
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
