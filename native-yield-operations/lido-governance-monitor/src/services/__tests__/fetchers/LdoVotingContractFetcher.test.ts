import { jest, describe, it, expect, beforeEach } from "@jest/globals";

import { ProposalSource } from "../../../core/entities/ProposalSource.js";
import { IProposalRepository } from "../../../core/repositories/IProposalRepository.js";
import { ILidoGovernanceMonitorLogger } from "../../../utils/logging/index.js";
import { LdoVotingContractFetcher } from "../../fetchers/LdoVotingContractFetcher.js";

const createLoggerMock = (): jest.Mocked<ILidoGovernanceMonitorLogger> => ({
  name: "test-logger",
  critical: jest.fn(),
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

interface MockPublicClient {
  getLogs: jest.Mock;
  getBlock: jest.Mock;
}

const createPublicClientMock = (): MockPublicClient => ({
  getLogs: jest.fn(),
  getBlock: jest.fn(),
});

const createProposalRepositoryMock = (): jest.Mocked<Pick<IProposalRepository, "findLatestSourceIdBySource">> => ({
  findLatestSourceIdBySource: jest.fn(),
});

describe("LdoVotingContractFetcher", () => {
  let logger: jest.Mocked<ILidoGovernanceMonitorLogger>;
  let publicClient: MockPublicClient;
  let proposalRepository: jest.Mocked<Pick<IProposalRepository, "findLatestSourceIdBySource">>;
  const contractAddress = "0x2e59a20f205bb85a89c53f1936454680651e618e";
  const initialEventScanBlock = 11473216n;

  beforeEach(() => {
    logger = createLoggerMock();
    publicClient = createPublicClientMock();
    proposalRepository = createProposalRepositoryMock();
  });

  const createFetcher = (): LdoVotingContractFetcher =>
    new LdoVotingContractFetcher(
      logger,
      publicClient as never,
      contractAddress,
      initialEventScanBlock,
      proposalRepository as unknown as IProposalRepository,
    );

  describe("getLatestProposals", () => {
    it("fetches StartVote events and returns CreateProposalInput[]", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue(null);
      publicClient.getLogs.mockResolvedValue([
        {
          args: { voteId: 180n, creator: "0xabc123", metadata: "Vote metadata text" },
          blockNumber: 100n,
        },
      ]);
      publicClient.getBlock.mockResolvedValue({ timestamp: 1700000000n });

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toEqual([
        {
          source: ProposalSource.LDO_VOTING_CONTRACT,
          sourceId: "180",
          url: "https://vote.lido.fi/vote/180",
          title: "LDO Contract vote 180",
          author: "0xabc123",
          sourceCreatedAt: new Date(1700000000 * 1000),
          text: "Vote metadata text",
          sourceBlockNumber: 100n,
        },
      ]);
      expect(publicClient.getBlock).toHaveBeenCalledWith({ blockNumber: 100n });
    });

    it("returns empty array when getLogs returns no events", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue(null);
      publicClient.getLogs.mockResolvedValue([]);

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toEqual([]);
      expect(publicClient.getBlock).not.toHaveBeenCalled();
    });

    it("returns empty array and logs warning when getLogs throws", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue(null);
      publicClient.getLogs.mockRejectedValue(new Error("RPC connection failed"));

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toEqual([]);
      expect(logger.warn).toHaveBeenCalledWith(
        expect.stringContaining("Failed to fetch LDO voting contract events"),
        expect.objectContaining({ error: "RPC connection failed" }),
      );
    });

    it("maps all fields correctly from event data", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue(null);
      publicClient.getLogs.mockResolvedValue([
        {
          args: {
            voteId: 42n,
            creator: "0x1234567890abcdef1234567890abcdef12345678",
            metadata: "Omnibus vote: 1) Fund xyz, 2) Update oracle",
          },
          blockNumber: 999n,
        },
      ]);
      publicClient.getBlock.mockResolvedValue({ timestamp: 1609459200n }); // 2021-01-01T00:00:00Z

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toHaveLength(1);
      const proposal = result[0];
      expect(proposal.source).toBe(ProposalSource.LDO_VOTING_CONTRACT);
      expect(proposal.sourceId).toBe("42");
      expect(proposal.url).toBe("https://vote.lido.fi/vote/42");
      expect(proposal.title).toBe("LDO Contract vote 42");
      expect(proposal.author).toBe("0x1234567890abcdef1234567890abcdef12345678");
      expect(proposal.sourceCreatedAt).toEqual(new Date("2021-01-01T00:00:00.000Z"));
      expect(proposal.text).toBe("Omnibus vote: 1) Fund xyz, 2) Update oracle");
      expect(proposal.sourceBlockNumber).toBe(999n);
    });

    it("skips events where getBlock fails and continues with others", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue(null);
      publicClient.getLogs.mockResolvedValue([
        { args: { voteId: 10n, creator: "0xaaa", metadata: "vote 10" }, blockNumber: 100n },
        { args: { voteId: 11n, creator: "0xbbb", metadata: "vote 11" }, blockNumber: 200n },
        { args: { voteId: 12n, creator: "0xccc", metadata: "vote 12" }, blockNumber: 300n },
      ]);
      publicClient.getBlock
        .mockResolvedValueOnce({ timestamp: 1700000000n })
        .mockRejectedValueOnce(new Error("block not found"))
        .mockResolvedValueOnce({ timestamp: 1700000200n });

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toHaveLength(2);
      expect(result[0].sourceId).toBe("10");
      expect(result[1].sourceId).toBe("12");
      expect(logger.warn).toHaveBeenCalledWith(
        expect.stringContaining("Failed to fetch block for vote 11"),
        expect.objectContaining({ error: "block not found" }),
      );
    });

    it("calls getLogs with initialEventScanBlock when no DB entry exists", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue(null);
      publicClient.getLogs.mockResolvedValue([]);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.findLatestSourceIdBySource).toHaveBeenCalledWith(ProposalSource.LDO_VOTING_CONTRACT);
      expect(publicClient.getLogs).toHaveBeenCalledWith(expect.objectContaining({ fromBlock: initialEventScanBlock }));
    });

    it("uses block number from DB lookup as fromBlock when DB has latest voteId", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue("180");
      // First call: lookup for voteId 180
      publicClient.getLogs.mockResolvedValueOnce([{ args: { voteId: 180n }, blockNumber: 100n }]);
      // Second call: main fetch from block 100
      publicClient.getLogs.mockResolvedValueOnce([
        {
          args: { voteId: 180n, creator: "0xabc", metadata: "vote 180" },
          blockNumber: 100n,
        },
        {
          args: { voteId: 181n, creator: "0xdef", metadata: "vote 181" },
          blockNumber: 200n,
        },
      ]);
      publicClient.getBlock.mockResolvedValue({ timestamp: 1700000000n });

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(publicClient.getLogs).toHaveBeenCalledTimes(2);
      // First call: lookup by voteId
      expect(publicClient.getLogs).toHaveBeenNthCalledWith(
        1,
        expect.objectContaining({
          args: { voteId: 180n },
          fromBlock: initialEventScanBlock,
          toBlock: "latest",
        }),
      );
      // Second call: main fetch using discovered block
      expect(publicClient.getLogs).toHaveBeenNthCalledWith(
        2,
        expect.objectContaining({
          fromBlock: 100n,
          toBlock: "latest",
        }),
      );
      expect(result).toHaveLength(2);
      expect(result[0].sourceId).toBe("180");
      expect(result[1].sourceId).toBe("181");
    });

    it("falls back to 'earliest' when initialEventScanBlock is undefined", async () => {
      // Arrange
      const fetcher = new LdoVotingContractFetcher(
        logger,
        publicClient as never,
        contractAddress,
        undefined,
        proposalRepository as unknown as IProposalRepository,
      );
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue(null);
      publicClient.getLogs.mockResolvedValue([]);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(publicClient.getLogs).toHaveBeenCalledWith(expect.objectContaining({ fromBlock: "earliest" }));
    });

    it("falls back to initialEventScanBlock when lookup RPC call fails", async () => {
      // Arrange
      const fetcher = createFetcher();
      proposalRepository.findLatestSourceIdBySource.mockResolvedValue("180");
      // First call: lookup fails
      publicClient.getLogs.mockRejectedValueOnce(new Error("RPC lookup failed"));
      // Second call: fallback main fetch
      publicClient.getLogs.mockResolvedValueOnce([
        {
          args: { voteId: 180n, creator: "0xabc", metadata: "vote 180" },
          blockNumber: 100n,
        },
      ]);
      publicClient.getBlock.mockResolvedValue({ timestamp: 1700000000n });

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(logger.warn).toHaveBeenCalledWith(
        expect.stringContaining("Failed to look up block for latest voteId 180"),
        expect.objectContaining({ error: "RPC lookup failed" }),
      );
      expect(publicClient.getLogs).toHaveBeenNthCalledWith(2, expect.objectContaining({ fromBlock: initialEventScanBlock }));
      expect(result).toHaveLength(1);
    });
  });
});
