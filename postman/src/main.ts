import { ExpressApiApplication, type IApplication } from "@consensys/linea-shared-utils";

import { getConfig } from "./application/config/getConfig";
import { createViemClients, createL1ToL2Flow, createL2ToL1Flow, createDatabaseCleaningPoller } from "./CompositionRoot";
import { PostmanLogger } from "./infrastructure/logging/PostmanLogger";
import { MessageMetricsUpdater } from "./infrastructure/metrics/MessageMetricsUpdater";
import { PostmanMetricsService } from "./infrastructure/metrics/PostmanMetricsService";
import { SponsorshipMetricsUpdater } from "./infrastructure/metrics/SponsorshipMetricsUpdater";
import { TransactionMetricsUpdater } from "./infrastructure/metrics/TransactionMetricsUpdater";
import { DB } from "./infrastructure/persistence/DataSource";
import { TypeOrmMessageRepository } from "./infrastructure/persistence/repositories/TypeOrmMessageRepository";
import { MessageStatusSubscriber } from "./infrastructure/persistence/subscribers/MessageStatusSubscriber";

import type { PostmanOptions } from "./application/config/PostmanConfig";
import type { IPoller } from "./application/pollers/Poller";
import type { ILogger } from "./domain/ports/ILogger";
import type { DBOptions } from "./infrastructure/persistence/config/types";
import type { LoggerOptions } from "winston";

export async function startPostman(options: PostmanOptions): Promise<{
  stop: () => void;
}> {
  const config = getConfig(options);
  const loggerFactory = (name: string): ILogger =>
    new PostmanLogger(name, config.loggerOptions as LoggerOptions | undefined);
  const logger = loggerFactory("PostmanMain");

  process.on("unhandledRejection", (reason) => {
    logger.error("Unhandled promise rejection: %s", reason);
  });

  process.on("uncaughtException", (error) => {
    logger.error("Uncaught exception, shutting down: %s", error);
    process.exit(1);
  });

  // 1. Database
  const db = DB.create(config.databaseOptions as DBOptions);

  try {
    await db.initialize();
    logger.info("Database connection established successfully.");
  } catch (error) {
    logger.error("Failed to connect to the database.", error);
    throw error;
  }

  // 2. Repository
  const messageRepository = new TypeOrmMessageRepository(db);

  // 3. Metrics
  const postmanMetricsService = new PostmanMetricsService();
  const messageMetricsUpdater = new MessageMetricsUpdater(db.manager, postmanMetricsService);
  const sponsorshipMetricsUpdater = new SponsorshipMetricsUpdater(postmanMetricsService);
  const transactionMetricsUpdater = new TransactionMetricsUpdater(postmanMetricsService);

  await messageMetricsUpdater.initialize();

  // 4. Register subscriber for metrics
  const messageStatusSubscriber = new MessageStatusSubscriber(
    messageMetricsUpdater,
    loggerFactory("MessageStatusSubscriber"),
  );
  db.subscribers.push(messageStatusSubscriber);

  // 5. Create viem clients
  const clients = await createViemClients(config);
  logger.info("Viem clients initialized for L1 chain=%s and L2 chain=%s.", clients.l1Chain.id, clients.l2Chain.id);

  const metricsUpdaters = {
    sponsorship: sponsorshipMetricsUpdater,
    transaction: transactionMetricsUpdater,
  };

  // 6. Create flow pollers
  const pollers: IPoller[] = [];

  if (config.l1L2AutoClaimEnabled) {
    const l1ToL2Pollers = createL1ToL2Flow(clients, messageRepository, metricsUpdaters, config, loggerFactory);
    pollers.push(...l1ToL2Pollers);
    logger.info("L1→L2 flow enabled with %s pollers.", l1ToL2Pollers.length);
  }

  if (config.l2L1AutoClaimEnabled) {
    const l2ToL1Pollers = createL2ToL1Flow(clients, messageRepository, metricsUpdaters, config, loggerFactory);
    pollers.push(...l2ToL1Pollers);
    logger.info("L2→L1 flow enabled with %s pollers.", l2ToL1Pollers.length);
  }

  // 7. Database cleaner
  const dbCleanerPoller = createDatabaseCleaningPoller(messageRepository, config, loggerFactory);
  if (dbCleanerPoller) {
    pollers.push(dbCleanerPoller);
  }

  // 8. API
  const api: IApplication = new ExpressApiApplication(
    config.apiConfig.port,
    postmanMetricsService,
    loggerFactory("ExpressApiApplication"),
  );

  // 9. Start all
  for (const poller of pollers) {
    poller.start();
  }
  api.start();

  logger.info("All postman services have been started.");

  // 10. Return stop handle
  return {
    stop: () => {
      for (const poller of pollers) {
        poller.stop();
      }
      api.stop();
      logger.info("All postman services have been stopped.");
    },
  };
}
