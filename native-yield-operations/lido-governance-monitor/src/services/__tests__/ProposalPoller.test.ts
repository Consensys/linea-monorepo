import { ILogger } from "@consensys/linea-shared-utils";
import { jest, describe, it, expect, beforeEach, afterEach } from "@jest/globals";

import { IDiscourseClient } from "../../core/clients/IDiscourseClient.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { RawDiscourseProposal, RawDiscourseProposalList } from "../../core/entities/RawDiscourseProposal.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";
import { INormalizationService } from "../../core/services/INormalizationService.js";
import { ProposalPoller } from "../ProposalPoller.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

describe("ProposalPoller", () => {
  let poller: ProposalPoller;
  let logger: jest.Mocked<ILogger>;
  let discourseClient: jest.Mocked<IDiscourseClient>;
  let normalizationService: jest.Mocked<INormalizationService>;
  let proposalRepository: jest.Mocked<IProposalRepository>;

  const createMockProposalList = (topics: Array<{ id: number; slug: string }>): RawDiscourseProposalList => ({
    topic_list: { topics },
  });

  const createMockProposal = (id: number, slug: string): RawDiscourseProposal => ({
    id,
    slug,
    title: `Proposal ${id}`,
    created_at: "2024-01-15T10:00:00.000Z",
    post_stream: {
      posts: [{ id: 1, username: "author", cooked: "<p>Content</p>", post_url: "/t/1", created_at: "2024-01-15" }],
    },
  });

  beforeEach(() => {
    logger = createLoggerMock();
    discourseClient = {
      fetchLatestProposals: jest.fn(),
      fetchProposalDetails: jest.fn(),
    } as jest.Mocked<IDiscourseClient>;
    normalizationService = {
      normalizeDiscourseProposal: jest.fn(),
      stripHtml: jest.fn(),
    } as jest.Mocked<INormalizationService>;
    proposalRepository = {
      findBySourceAndSourceId: jest.fn(),
      findByState: jest.fn(),
      create: jest.fn(),
      updateState: jest.fn(),
      saveAnalysis: jest.fn(),
      incrementAnalysisAttempt: jest.fn(),
      incrementNotifyAttempt: jest.fn(),
      markNotified: jest.fn(),
    } as jest.Mocked<IProposalRepository>;

    poller = new ProposalPoller(logger, discourseClient, normalizationService, proposalRepository);
  });

  afterEach(() => {
  });

  describe("pollOnce", () => {
    it("fetches latest proposals from Discourse", async () => {
      // Arrange
      discourseClient.fetchLatestProposals.mockResolvedValue(createMockProposalList([]));

      // Act
      await poller.pollOnce();

      // Assert
      expect(discourseClient.fetchLatestProposals).toHaveBeenCalled();
    });

    it("skips processing when fetch returns undefined", async () => {
      // Arrange
      discourseClient.fetchLatestProposals.mockResolvedValue(undefined);

      // Act
      await poller.pollOnce();

      // Assert
      expect(discourseClient.fetchProposalDetails).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Failed to fetch latest proposals from Discourse");
    });

    it("fetches details for each topic in the list", async () => {
      // Arrange
      const topicList = createMockProposalList([
        { id: 100, slug: "proposal-a" },
        { id: 101, slug: "proposal-b" },
      ]);
      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValue(undefined);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue(null);

      // Act
      await poller.pollOnce();

      // Assert
      expect(discourseClient.fetchProposalDetails).toHaveBeenCalledWith(100);
      expect(discourseClient.fetchProposalDetails).toHaveBeenCalledWith(101);
    });

    it("skips proposals that already exist in database", async () => {
      // Arrange
      const topicList = createMockProposalList([{ id: 100, slug: "existing-proposal" }]);
      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue({ id: "existing-uuid" } as any);

      // Act
      await poller.pollOnce();

      // Assert
      expect(discourseClient.fetchProposalDetails).not.toHaveBeenCalled();
      expect(logger.debug).toHaveBeenCalledWith("Proposal already exists, skipping", expect.any(Object));
    });

    it("normalizes and creates new proposals", async () => {
      // Arrange
      const topicList = createMockProposalList([{ id: 100, slug: "new-proposal" }]);
      const proposalDetails = createMockProposal(100, "new-proposal");
      const normalizedInput = {
        source: ProposalSource.DISCOURSE,
        sourceId: "100",
        url: "https://research.lido.fi/t/new-proposal/100",
        title: "Proposal 100",
        author: "author",
        sourceCreatedAt: new Date("2024-01-15"),
        text: "Content",
      };

      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValue(proposalDetails);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue(null);
      normalizationService.normalizeDiscourseProposal.mockReturnValue(normalizedInput);
      proposalRepository.create.mockResolvedValue({ id: "new-uuid" } as any);

      // Act
      await poller.pollOnce();

      // Assert
      expect(normalizationService.normalizeDiscourseProposal).toHaveBeenCalledWith(proposalDetails);
      expect(proposalRepository.create).toHaveBeenCalledWith(normalizedInput);
      expect(logger.info).toHaveBeenCalledWith("Created new proposal", expect.any(Object));
    });

    it("skips when proposal details fetch fails", async () => {
      // Arrange
      const topicList = createMockProposalList([{ id: 100, slug: "proposal" }]);
      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValue(undefined);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue(null);

      // Act
      await poller.pollOnce();

      // Assert
      expect(normalizationService.normalizeDiscourseProposal).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Failed to fetch proposal details", expect.any(Object));
    });

    it("handles errors during proposal creation gracefully", async () => {
      // Arrange
      const topicList = createMockProposalList([{ id: 100, slug: "proposal" }]);
      const proposalDetails = createMockProposal(100, "proposal");
      const normalizedInput = {
        source: ProposalSource.DISCOURSE,
        sourceId: "100",
        url: "url",
        title: "Title",
        author: "author",
        sourceCreatedAt: new Date(),
        text: "Content",
      };

      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValue(proposalDetails);
      proposalRepository.findBySourceAndSourceId.mockResolvedValue(null);
      normalizationService.normalizeDiscourseProposal.mockReturnValue(normalizedInput);
      proposalRepository.create.mockRejectedValue(new Error("Database error"));

      // Act
      await poller.pollOnce();

      // Assert
      expect(logger.error).toHaveBeenCalledWith("Failed to create proposal", expect.any(Object));
    });

    it("catches and logs errors without throwing", async () => {
      // Arrange
      discourseClient.fetchLatestProposals.mockRejectedValue(new Error("Network error"));

      // Act
      await poller.pollOnce();

      // Assert
      expect(logger.error).toHaveBeenCalledWith("Proposal polling failed", expect.any(Error));
    });
  });
});
