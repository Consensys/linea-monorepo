import { IDiscourseClient } from "../../core/clients/IDiscourseClient.js";
import { CreateProposalInput } from "../../core/entities/Proposal.js";
import { INormalizationService } from "../../core/services/INormalizationService.js";
import { IProposalFetcher } from "../../core/services/IProposalFetcher.js";
import { ILidoGovernanceMonitorLogger } from "../../utils/logging/index.js";

export class DiscourseFetcher implements IProposalFetcher {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly discourseClient: IDiscourseClient,
    private readonly normalizationService: INormalizationService,
    private readonly maxTopicsPerPoll: number = 20,
  ) {}

  async getLatestProposals(): Promise<CreateProposalInput[]> {
    const proposalList = await this.discourseClient.fetchLatestProposals();

    if (!proposalList) {
      this.logger.warn("Failed to fetch latest proposals from Discourse");
      return [];
    }

    const allTopics = proposalList.topic_list.topics;
    const topics = allTopics.slice(0, this.maxTopicsPerPoll);
    this.logger.debug("Fetched proposal list", { total: allTopics.length, processing: topics.length });

    const results: CreateProposalInput[] = [];

    for (const topic of topics) {
      const result = await this.processTopic(topic.id);
      if (result) {
        results.push(result);
      }
    }

    return results;
  }

  private async processTopic(topicId: number): Promise<CreateProposalInput | null> {
    const proposalDetails = await this.discourseClient.fetchProposalDetails(topicId);

    if (!proposalDetails) {
      this.logger.warn("Failed to fetch proposal details", { topicId });
      return null;
    }

    try {
      return this.normalizationService.normalizeDiscourseProposal(proposalDetails);
    } catch (error) {
      this.logger.error("Failed to normalize proposal", { topicId, error });
      return null;
    }
  }
}
