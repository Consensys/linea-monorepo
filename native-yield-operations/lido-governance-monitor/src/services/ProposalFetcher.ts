import { CreateProposalInput } from "../core/entities/Proposal.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { IProposalFetcher } from "../core/services/IProposalFetcher.js";
import { ILidoGovernanceMonitorLogger } from "../utils/logging/index.js";

export class ProposalFetcher implements IProposalFetcher {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly sourceFetchers: IProposalFetcher[],
    private readonly proposalRepository: IProposalRepository,
  ) {}

  async getLatestProposals(): Promise<CreateProposalInput[]> {
    this.logger.info("Starting proposal polling");

    const results = await Promise.allSettled(
      this.sourceFetchers.map((fetcher) => fetcher.getLatestProposals()),
    );

    const proposals: CreateProposalInput[] = [];
    for (const result of results) {
      if (result.status === "fulfilled") {
        proposals.push(...result.value);
      } else {
        this.logger.critical("Source fetcher failed", { error: result.reason });
      }
    }

    let created = 0;
    for (const proposal of proposals) {
      try {
        const isNew = await this.persistIfNew(proposal);
        if (isNew) created++;
      } catch (error) {
        this.logger.critical("Failed to create proposal", {
          source: proposal.source,
          sourceId: proposal.sourceId,
          error,
        });
      }
    }

    this.logger.info("Proposal polling completed", {
      sources: this.sourceFetchers.length,
      fetched: proposals.length,
      created,
    });

    return proposals;
  }

  private async persistIfNew(proposal: CreateProposalInput): Promise<boolean> {
    const existing = await this.proposalRepository.findBySourceAndSourceId(
      proposal.source,
      proposal.sourceId,
    );

    if (existing) {
      this.logger.debug("Proposal already exists, skipping", {
        source: proposal.source,
        sourceId: proposal.sourceId,
      });
      return false;
    }

    const result = await this.proposalRepository.create(proposal);
    this.logger.info("Created new proposal", { id: result.id, title: proposal.title });
    return true;
  }
}
