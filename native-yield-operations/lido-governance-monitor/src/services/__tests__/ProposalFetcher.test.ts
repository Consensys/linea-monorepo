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
      upsert: jest.fn(),
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
      proposalRepository.upsert.mockResolvedValue({ proposal: { id: "existing-uuid" } as any, isNew: false });

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.upsert).toHaveBeenCalledWith(proposal);
      expect(logger.debug).toHaveBeenCalledWith("Proposal already exists, skipping", expect.any(Object));
    });

    it("creates new proposals not in DB", async () => {
      // Arrange
      const proposal = createProposalInput();
      sourceFetcherA.getLatestProposals.mockResolvedValue([proposal]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);
      proposalRepository.upsert.mockResolvedValue({ proposal: { id: "new-uuid" } as any, isNew: true });

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.upsert).toHaveBeenCalledWith(proposal);
      expect(logger.info).toHaveBeenCalledWith("Created new proposal", { id: "new-uuid", title: proposal.title });
    });

    it("handles mixed new and existing proposals", async () => {
      // Arrange
      const existingProposal = createProposalInput({ sourceId: "100" });
      const newProposal = createProposalInput({ sourceId: "200" });
      sourceFetcherA.getLatestProposals.mockResolvedValue([existingProposal, newProposal]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);
      proposalRepository.upsert
        .mockResolvedValueOnce({ proposal: { id: "existing-uuid" } as any, isNew: false })
        .mockResolvedValueOnce({ proposal: { id: "new-uuid" } as any, isNew: true });

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.upsert).toHaveBeenCalledTimes(2);
      expect(proposalRepository.upsert).toHaveBeenCalledWith(existingProposal);
      expect(proposalRepository.upsert).toHaveBeenCalledWith(newProposal);
    });

    it("continues when one fetcher rejects", async () => {
      // Arrange
      const proposal = createProposalInput();
      sourceFetcherA.getLatestProposals.mockRejectedValue(new Error("Network error"));
      sourceFetcherB.getLatestProposals.mockResolvedValue([proposal]);
      proposalRepository.upsert.mockResolvedValue({ proposal: { id: "new-uuid" } as any, isNew: true });

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(logger.critical).toHaveBeenCalledWith(
        "Source fetcher failed",
        expect.objectContaining({ error: expect.any(Error) }),
      );
      expect(proposalRepository.upsert).toHaveBeenCalledWith(proposal);
    });

    it("handles DB errors during upsert", async () => {
      // Arrange
      const proposal = createProposalInput();
      sourceFetcherA.getLatestProposals.mockResolvedValue([proposal]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);
      proposalRepository.upsert.mockRejectedValue(new Error("DB error"));

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
      expect(proposalRepository.upsert).not.toHaveBeenCalled();
      expect(logger.info).toHaveBeenCalledWith("Proposal polling completed", expect.any(Object));
    });

    it("works when all fetchers return empty arrays", async () => {
      // Arrange
      sourceFetcherA.getLatestProposals.mockResolvedValue([]);
      sourceFetcherB.getLatestProposals.mockResolvedValue([]);

      // Act
      await fetcher.getLatestProposals();

      // Assert
      expect(proposalRepository.upsert).not.toHaveBeenCalled();
    });
  });
});
