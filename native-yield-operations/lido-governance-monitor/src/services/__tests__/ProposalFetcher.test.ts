import { jest, describe, it, expect, beforeEach } from "@jest/globals";

import { CreateProposalInput } from "../../core/entities/Proposal.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";
import { IProposalFetcher } from "../../core/services/IProposalFetcher.js";
import { ILidoGovernanceMonitorLogger } from "../../utils/logging/index.js";
import { ProposalFetcher } from "../ProposalFetcher.js";

const createLoggerMock = (): jest.Mocked<ILidoGovernanceMonitorLogger> => ({
  name: "test-logger",
  critical: jest.fn(),
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

const createSourceFetcherMock = (): jest.Mocked<IProposalFetcher> => ({
  getLatestProposals: jest.fn(),
});

const createProposalInput = (overrides: Partial<CreateProposalInput> = {}): CreateProposalInput => ({
  source: ProposalSource.DISCOURSE,
  sourceId: "100",
  url: "https://research.lido.fi/t/proposal/100",
  title: "Proposal 100",
  author: "author",
  sourceCreatedAt: new Date("2024-01-15"),
  rawProposalText: "Content",
  ...overrides,
});

describe("ProposalFetcher", () => {
  let fetcher: ProposalFetcher;
  let logger: jest.Mocked<ILidoGovernanceMonitorLogger>;
  let proposalRepository: jest.Mocked<IProposalRepository>;
  let sourceFetcherA: jest.Mocked<IProposalFetcher>;
  let sourceFetcherB: jest.Mocked<IProposalFetcher>;

  beforeEach(() => {
    logger = createLoggerMock();
    proposalRepository = {
      findBySourceAndSourceId: jest.fn(),
      findByStateForAnalysis: jest.fn(),
      findByStateForNotification: jest.fn(),
      create: jest.fn(),
      updateState: jest.fn(),
      saveAnalysis: jest.fn(),
      incrementAnalysisAttempt: jest.fn(),
      incrementNotifyAttempt: jest.fn(),
      markNotified: jest.fn(),
    } as jest.Mocked<IProposalRepository>;
    sourceFetcherA = createSourceFetcherMock();
    sourceFetcherB = createSourceFetcherMock();

    fetcher = new ProposalFetcher(logger, [sourceFetcherA, sourceFetcherB], proposalRepository);
  });

  describe("getLatestProposals", () => {
    it("calls getLatestProposals on all fetchers", async () => {
      // Arrange
      sourceFetcherA.getLatestProposals.mockResolvedValue([]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(sourceFetcherA.getLatestProposals).toHaveBeenCalled();
      expect(sourceFetcherB.getLatestProposals).toHaveBeenCalled();
    });

    it("deduplicates proposals already in DB", async () => {
      // Arrange
      const proposal = createProposalInput();
      sourceFetcherA.getLatestProposals.mockResolvedValue([proposal]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue({ id: "existing-uuid" } as any);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.findBySourceAndSourceId).toHaveBeenCalledWith(ProposalSource.DISCOURSE, "100");
      expect(proposalRepository.create).not.toHaveBeenCalled();
    });

    it("creates new proposals not in DB", async () => {
      // Arrange
      const proposal = createProposalInput();
      sourceFetcherA.getLatestProposals.mockResolvedValue([proposal]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue(null);
      proposalRepository.create.mockResolvedValue({ id: "new-uuid" } as any);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.create).toHaveBeenCalledWith(proposal);
    });

    it("handles mixed new and existing proposals", async () => {
      // Arrange
      const existingProposal = createProposalInput({ sourceId: "100" });
      const newProposal = createProposalInput({ sourceId: "200" });
      sourceFetcherA.getLatestProposals.mockResolvedValue([existingProposal, newProposal]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);
      proposalRepository.findBySourceAndSourceId
        .mockResolvedValueOnce({ id: "existing-uuid" } as any)
        .mockResolvedValueOnce(null);
      proposalRepository.create.mockResolvedValue({ id: "new-uuid" } as any);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.create).toHaveBeenCalledTimes(1);
      expect(proposalRepository.create).toHaveBeenCalledWith(newProposal);
    });

    it("continues when one fetcher rejects", async () => {
      // Arrange
      const proposal = createProposalInput();
      sourceFetcherA.getLatestProposals.mockRejectedValue(new Error("Network error"));
      sourceFetcherB.getLatestProposals.mockResolvedValue([proposal]);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue(null);
      proposalRepository.create.mockResolvedValue({ id: "new-uuid" } as any);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(logger.critical).toHaveBeenCalledWith(
        "Source fetcher failed",
        expect.objectContaining({ error: expect.any(Error) }),
      );
      expect(proposalRepository.create).toHaveBeenCalledWith(proposal);
    });

    it("handles DB errors during create", async () => {
      // Arrange
      const proposal = createProposalInput();
      sourceFetcherA.getLatestProposals.mockResolvedValue([proposal]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue(null);
      proposalRepository.create.mockRejectedValue(new Error("DB error"));

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(logger.critical).toHaveBeenCalledWith("Failed to create proposal", expect.any(Object));
    });

    it("works with empty fetcher array", async () => {
      // Arrange
      const emptyFetcher = new ProposalFetcher(logger, [], proposalRepository);

      // Act
      await emptyFetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.create).not.toHaveBeenCalled();
      expect(logger.info).toHaveBeenCalledWith("Proposal polling completed", expect.any(Object));
    });

    it("works when all fetchers return empty arrays", async () => {
      // Arrange
      sourceFetcherA.getLatestProposals.mockResolvedValue([]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.findBySourceAndSourceId).not.toHaveBeenCalled();
      expect(proposalRepository.create).not.toHaveBeenCalled();
    });
  });
});
