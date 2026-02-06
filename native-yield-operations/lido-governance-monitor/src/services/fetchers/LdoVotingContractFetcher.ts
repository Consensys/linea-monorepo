import { type Address, type PublicClient, parseAbiItem } from "viem";

import { CreateProposalInput } from "../../core/entities/Proposal.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";
import { IProposalFetcher } from "../../core/services/IProposalFetcher.js";
import { ILidoGovernanceMonitorLogger } from "../../utils/logging/index.js";

const START_VOTE_EVENT = parseAbiItem(
  "event StartVote(uint256 indexed voteId, address indexed creator, string metadata)",
);

export class LdoVotingContractFetcher implements IProposalFetcher {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly publicClient: PublicClient,
    private readonly contractAddress: Address,
    private readonly maxVotesPerPoll: number,
    private readonly proposalRepository: IProposalRepository,
  ) {}

  async getLatestProposals(): Promise<CreateProposalInput[]> {
    const fromBlock = await this.resolveFromBlock();

    let logs;
    try {
      logs = await this.publicClient.getLogs({
        address: this.contractAddress,
        event: START_VOTE_EVENT,
        fromBlock,
        toBlock: "latest",
      });
    } catch (error) {
      this.logger.warn("Failed to fetch LDO voting contract events", {
        error: error instanceof Error ? error.message : String(error),
      });
      return [];
    }

    const recentLogs = logs.slice(-this.maxVotesPerPoll);

    const proposals: CreateProposalInput[] = [];
    for (const log of recentLogs) {
      const { voteId, creator, metadata } = log.args;
      try {
        const block = await this.publicClient.getBlock({ blockNumber: log.blockNumber! });
        proposals.push({
          source: ProposalSource.LDO_VOTING_CONTRACT,
          sourceId: String(voteId),
          url: `https://vote.lido.fi/vote/${voteId}`,
          title: `LDO Contract vote ${voteId}`,
          author: creator ?? null,
          sourceCreatedAt: new Date(Number(block.timestamp) * 1000),
          text: metadata ?? "",
          sourceBlockNumber: log.blockNumber!,
        });
      } catch (error) {
        this.logger.warn(`Failed to fetch block for vote ${voteId}`, {
          error: error instanceof Error ? error.message : String(error),
        });
      }
    }

    return proposals;
  }

  private async resolveFromBlock(): Promise<bigint | "earliest"> {
    const latestSourceId = await this.proposalRepository.findLatestSourceIdBySource(ProposalSource.LDO_VOTING_CONTRACT);

    if (!latestSourceId) {
      return "earliest";
    }

    try {
      const lookupLogs = await this.publicClient.getLogs({
        address: this.contractAddress,
        event: START_VOTE_EVENT,
        args: { voteId: BigInt(latestSourceId) },
        fromBlock: "earliest",
        toBlock: "latest",
      });

      if (lookupLogs.length > 0 && lookupLogs[0].blockNumber != null) {
        return lookupLogs[0].blockNumber;
      }
    } catch (error) {
      this.logger.warn(`Failed to look up block for latest voteId ${latestSourceId}`, {
        error: error instanceof Error ? error.message : String(error),
      });
    }

    return "earliest";
  }
}
