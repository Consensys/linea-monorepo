import { CreateProposalInput } from "../../core/entities/Proposal.js";
import { IProposalFetcher } from "../../core/services/IProposalFetcher.js";
import { ILidoGovernanceMonitorLogger } from "../../utils/logging/index.js";

export class LdoVotingContractFetcher implements IProposalFetcher {
  constructor(private readonly logger: ILidoGovernanceMonitorLogger) {}

  async getLatestProposals(): Promise<CreateProposalInput[]> {
    this.logger.debug("LdoVotingContractFetcher: not yet implemented");
    return [];
  }
}
