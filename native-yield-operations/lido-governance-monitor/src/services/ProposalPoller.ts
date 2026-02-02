import { ILogger } from "@consensys/linea-shared-utils";

import { IDiscourseClient } from "../core/clients/IDiscourseClient.js";
import { ProposalSource } from "../core/entities/ProposalSource.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { INormalizationService } from "../core/services/INormalizationService.js";
import { IProposalPoller } from "../core/services/IProposalPoller.js";

export class ProposalPoller implements IProposalPoller {
  constructor(
    private readonly logger: ILogger,
    private readonly discourseClient: IDiscourseClient,
    private readonly normalizationService: INormalizationService,
    private readonly proposalRepository: IProposalRepository,
  ) {}

  async pollOnce(): Promise<void> {
    try {
      this.logger.info("Starting proposal polling");

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

      this.logger.info("Proposal polling completed");
    } catch (error) {
      this.logger.error("Proposal polling failed", error);
    }
  }

  private async processTopic(topicId: number): Promise<void> {
    // Check if proposal already exists
    const existing = await this.proposalRepository.findBySourceAndSourceId(
      ProposalSource.DISCOURSE,
      topicId.toString(),
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
