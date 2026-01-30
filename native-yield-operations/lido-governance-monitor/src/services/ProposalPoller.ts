import { ILogger } from "@consensys/linea-shared-utils";
import { IProposalPoller } from "../core/services/IProposalPoller.js";
import { IDiscourseClient } from "../core/clients/IDiscourseClient.js";
import { INormalizationService } from "../core/services/INormalizationService.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { ProposalSource } from "../core/entities/ProposalSource.js";

export class ProposalPoller implements IProposalPoller {
  private intervalId: NodeJS.Timeout | null = null;

  constructor(
    private readonly logger: ILogger,
    private readonly discourseClient: IDiscourseClient,
    private readonly normalizationService: INormalizationService,
    private readonly proposalRepository: IProposalRepository,
    private readonly pollingIntervalMs: number
  ) {}

  start(): void {
    this.logger.info("ProposalPoller started", { intervalMs: this.pollingIntervalMs });

    // Initial poll
    void this.pollOnce();

    // Schedule subsequent polls
    this.intervalId = setInterval(() => {
      void this.pollOnce();
    }, this.pollingIntervalMs);
  }

  stop(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
    this.logger.info("ProposalPoller stopped");
  }

  async pollOnce(): Promise<void> {
    const proposalList = await this.discourseClient.fetchLatestProposals();

    if (!proposalList) {
      this.logger.warn("Failed to fetch latest proposals from Discourse");
      return;
    }

    const topics = proposalList.topic_list.topics;
    this.logger.debug("Fetched proposal list", { count: topics.length });

    for (const topic of topics) {
      await this.processTopic(topic.id);
    }
  }

  private async processTopic(topicId: number): Promise<void> {
    // Check if proposal already exists
    const existing = await this.proposalRepository.findBySourceAndSourceId(
      ProposalSource.DISCOURSE,
      topicId.toString()
    );

    if (existing) {
      this.logger.debug("Proposal already exists, skipping", { topicId });
      return;
    }

    // Fetch full proposal details
    const proposalDetails = await this.discourseClient.fetchProposalDetails(topicId);

    if (!proposalDetails) {
      this.logger.warn("Failed to fetch proposal details", { topicId });
      return;
    }

    // Normalize and create
    try {
      const normalizedInput = this.normalizationService.normalizeDiscourseProposal(proposalDetails);
      const created = await this.proposalRepository.create(normalizedInput);
      this.logger.info("Created new proposal", { id: created.id, title: normalizedInput.title });
    } catch (error) {
      this.logger.error("Failed to create proposal", { topicId, error });
    }
  }
}
