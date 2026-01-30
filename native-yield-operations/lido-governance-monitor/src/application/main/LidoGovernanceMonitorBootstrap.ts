import Anthropic from "@anthropic-ai/sdk";
import { createLogger, ILogger } from "@consensys/linea-shared-utils";
import { PrismaClient } from "@prisma/client";

import { Config } from "./config/index.js";
import { ClaudeAIClient } from "../../clients/ClaudeAIClient.js";
import { ProposalRepository } from "../../clients/db/ProposalRepository.js";
import { DiscourseClient } from "../../clients/DiscourseClient.js";
import { SlackClient } from "../../clients/SlackClient.js";
import { NormalizationService } from "../../services/NormalizationService.js";
import { NotificationService } from "../../services/NotificationService.js";
import { ProposalPoller } from "../../services/ProposalPoller.js";
import { ProposalProcessor } from "../../services/ProposalProcessor.js";

export class LidoGovernanceMonitorBootstrap {
  private constructor(
    private readonly logger: ILogger,
    private readonly prisma: PrismaClient,
    private readonly proposalPoller: ProposalPoller,
    private readonly proposalProcessor: ProposalProcessor,
    private readonly notificationService: NotificationService,
  ) {}

  static create(config: Config, systemPrompt: string): LidoGovernanceMonitorBootstrap {
    const logger = createLogger("LidoGovernanceMonitor");

    // Database
    const prisma = new PrismaClient({ datasourceUrl: config.database.url });

    // Repositories
    const proposalRepository = new ProposalRepository(prisma);

    // Clients
    const discourseClient = new DiscourseClient(logger, config.discourse.proposalsUrl);

    const anthropicClient = new Anthropic({ apiKey: config.anthropic.apiKey });
    const aiClient = new ClaudeAIClient(logger, anthropicClient, config.anthropic.model, systemPrompt);

    const slackClient = new SlackClient(logger, config.slack.webhookUrl);

    // Services
    const normalizationService = new NormalizationService(logger, discourseClient.getBaseUrl());

    const proposalPoller = new ProposalPoller(
      logger,
      discourseClient,
      normalizationService,
      proposalRepository,
      config.discourse.pollingIntervalMs,
    );

    const proposalProcessor = new ProposalProcessor(
      logger,
      aiClient,
      proposalRepository,
      config.riskAssessment.threshold,
      config.riskAssessment.promptVersion,
      config.riskAssessment.domainContext,
      config.processing.intervalMs,
    );

    const notificationService = new NotificationService(
      logger,
      slackClient,
      proposalRepository,
      config.processing.intervalMs,
    );

    return new LidoGovernanceMonitorBootstrap(logger, prisma, proposalPoller, proposalProcessor, notificationService);
  }

  // NOTE: start() and stop() are deliberately not unit tested.
  // These orchestration methods call real Prisma $connect/$disconnect and start
  // service intervals that interact with mocked dependencies in unpredictable ways,
  // leading to flaky tests. The wiring is verified via create() and getter tests.
  // Integration/E2E tests should cover the full lifecycle.

  async start(): Promise<void> {
    this.logger.info("Starting Lido Governance Monitor");

    await this.prisma.$connect();
    this.logger.info("Database connected");

    this.proposalPoller.start();
    this.proposalProcessor.start();
    this.notificationService.start();

    this.logger.info("All services started");
  }

  async stop(): Promise<void> {
    this.logger.info("Stopping Lido Governance Monitor");

    this.proposalPoller.stop();
    this.proposalProcessor.stop();
    this.notificationService.stop();

    await this.prisma.$disconnect();
    this.logger.info("Database disconnected");

    this.logger.info("All services stopped");
  }

  getProposalPoller(): ProposalPoller {
    return this.proposalPoller;
  }

  getProposalProcessor(): ProposalProcessor {
    return this.proposalProcessor;
  }

  getNotificationService(): NotificationService {
    return this.notificationService;
  }
}
