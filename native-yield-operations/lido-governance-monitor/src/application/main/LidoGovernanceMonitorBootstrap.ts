import Anthropic from "@anthropic-ai/sdk";
import { ExponentialBackoffRetryService, ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
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
    const logger = new WinstonLogger("LidoGovernanceMonitor");

    // Database
    const prisma = new PrismaClient({ datasourceUrl: config.database.url });

    // Repositories
    const proposalRepository = new ProposalRepository(prisma);

    // Shared services
    const retryService = new ExponentialBackoffRetryService(logger);

    // Clients
    const discourseClient = new DiscourseClient(logger, retryService, config.discourse.proposalsUrl);

    const anthropicClient = new Anthropic({ apiKey: config.anthropic.apiKey });
    const aiClient = new ClaudeAIClient(logger, anthropicClient, config.anthropic.model, systemPrompt);

    const slackClient = new SlackClient(logger, config.slack.webhookUrl, config.slack.auditWebhookUrl);

    // Services
    const normalizationService = new NormalizationService(logger, discourseClient.getBaseUrl());

    const proposalPoller = new ProposalPoller(logger, discourseClient, normalizationService, proposalRepository);

    const proposalProcessor = new ProposalProcessor(
      logger,
      aiClient,
      proposalRepository,
      config.riskAssessment.threshold,
      config.riskAssessment.promptVersion,
    );

    const notificationService = new NotificationService(
      logger,
      slackClient,
      proposalRepository,
      config.riskAssessment.threshold,
    );

    return new LidoGovernanceMonitorBootstrap(logger, prisma, proposalPoller, proposalProcessor, notificationService);
  }

  async start(): Promise<void> {
    this.logger.info("Starting Lido Governance Monitor");

    await this.prisma.$connect();
    this.logger.info("Database connected");

    await this.proposalPoller.pollOnce();
    await this.proposalProcessor.processOnce();
    await this.notificationService.notifyOnce();

    this.logger.info("Lido Governance Monitor execution completed");
  }

  async stop(): Promise<void> {
    this.logger.info("Stopping Lido Governance Monitor");
    await this.prisma.$disconnect();
    this.logger.info("Database disconnected");
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
