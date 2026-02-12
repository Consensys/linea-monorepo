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

    const results = await Promise.allSettled(this.sourceFetchers.map((fetcher) => fetcher.getLatestProposals()));

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
        const { proposal: persisted, isNew } = await this.proposalRepository.upsert(proposal);
        if (isNew) {
          this.logger.info("Created new proposal", { id: persisted.id, title: proposal.title });
          created++;
        } else {
          this.logger.debug("Proposal already exists, skipping", {
            source: proposal.source,
            sourceId: proposal.sourceId,
          });
        }
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
}
