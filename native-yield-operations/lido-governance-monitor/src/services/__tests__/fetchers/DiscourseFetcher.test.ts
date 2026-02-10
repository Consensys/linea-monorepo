import { jest, describe, it, expect, beforeEach } from "@jest/globals";

import { IDiscourseClient } from "../../../core/clients/IDiscourseClient.js";
import { ProposalSource } from "../../../core/entities/ProposalSource.js";
import { RawDiscourseProposal, RawDiscourseProposalList } from "../../../core/entities/RawDiscourseProposal.js";
import { INormalizationService } from "../../../core/services/INormalizationService.js";
import { ILidoGovernanceMonitorLogger } from "../../../utils/logging/index.js";
import { DiscourseFetcher } from "../../fetchers/DiscourseFetcher.js";

const createLoggerMock = (): jest.Mocked<ILidoGovernanceMonitorLogger> => ({
  name: "test-logger",
  critical: jest.fn(),
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

describe("DiscourseFetcher", () => {
  let fetcher: DiscourseFetcher;
  let logger: jest.Mocked<ILidoGovernanceMonitorLogger>;
  let discourseClient: jest.Mocked<IDiscourseClient>;
  let normalizationService: jest.Mocked<INormalizationService>;

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
      getBaseUrl: jest.fn(),
    } as jest.Mocked<IDiscourseClient>;
    normalizationService = {
      normalizeDiscourseProposal: jest.fn(),
      stripHtml: jest.fn(),
    } as jest.Mocked<INormalizationService>;

    fetcher = new DiscourseFetcher(logger, discourseClient, normalizationService, 20);
  });

  describe("getLatestProposals", () => {
    it("fetches proposals, normalizes, and returns CreateProposalInput[]", async () => {
      // Arrange
      const topicList = createMockProposalList([{ id: 100, slug: "proposal-a" }]);
      const proposalDetails = createMockProposal(100, "proposal-a");
      const normalizedInput = {
        source: ProposalSource.DISCOURSE,
        sourceId: "100",
        url: "https://research.lido.fi/t/proposal-a/100",
        title: "Proposal 100",
        author: "author",
        sourceCreatedAt: new Date("2024-01-15"),
        rawProposalText: "Content",
      };

      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValue(proposalDetails);
      normalizationService.normalizeDiscourseProposal.mockReturnValue(normalizedInput);

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toEqual([normalizedInput]);
      expect(discourseClient.fetchLatestProposals).toHaveBeenCalled();
      expect(discourseClient.fetchProposalDetails).toHaveBeenCalledWith(100);
      expect(normalizationService.normalizeDiscourseProposal).toHaveBeenCalledWith(proposalDetails);
    });

    it("returns empty array when fetchLatestProposals returns undefined", async () => {
      // Arrange
      discourseClient.fetchLatestProposals.mockResolvedValue(undefined);

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toEqual([]);
      expect(logger.warn).toHaveBeenCalledWith("Failed to fetch latest proposals from Discourse");
    });

    it("limits topics to maxTopicsPerPoll", async () => {
      // Arrange
      const topics = Array.from({ length: 5 }, (_, i) => ({ id: i + 1, slug: `proposal-${i + 1}` }));
      const topicList = createMockProposalList(topics);

      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValue(undefined);

      // Create fetcher with maxTopicsPerPoll=2
      const limitedFetcher = new DiscourseFetcher(logger, discourseClient, normalizationService, 2);

      // Act
      await limitedFetcher.getLatestProposals();

      // Assert
      expect(discourseClient.fetchProposalDetails).toHaveBeenCalledTimes(2);
      expect(discourseClient.fetchProposalDetails).toHaveBeenCalledWith(1);
      expect(discourseClient.fetchProposalDetails).toHaveBeenCalledWith(2);
    });

    it("skips topics where details fetch fails and continues with others", async () => {
      // Arrange
      const topicList = createMockProposalList([
        { id: 100, slug: "fails" },
        { id: 101, slug: "succeeds" },
      ]);
      const proposalDetails = createMockProposal(101, "succeeds");
      const normalizedInput = {
        source: ProposalSource.DISCOURSE,
        sourceId: "101",
        url: "https://research.lido.fi/t/succeeds/101",
        title: "Proposal 101",
        author: "author",
        sourceCreatedAt: new Date("2024-01-15"),
        rawProposalText: "Content",
      };

      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValueOnce(undefined).mockResolvedValueOnce(proposalDetails);
      normalizationService.normalizeDiscourseProposal.mockReturnValue(normalizedInput);

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toEqual([normalizedInput]);
      expect(logger.warn).toHaveBeenCalledWith("Failed to fetch proposal details", { topicId: 100 });
    });

    it("skips topics where normalization throws and continues with others", async () => {
      // Arrange
      const topicList = createMockProposalList([
        { id: 100, slug: "bad-normalize" },
        { id: 101, slug: "good-normalize" },
      ]);
      const badDetails = createMockProposal(100, "bad-normalize");
      const goodDetails = createMockProposal(101, "good-normalize");
      const normalizedInput = {
        source: ProposalSource.DISCOURSE,
        sourceId: "101",
        url: "https://research.lido.fi/t/good-normalize/101",
        title: "Proposal 101",
        author: "author",
        sourceCreatedAt: new Date("2024-01-15"),
        rawProposalText: "Content",
      };

      discourseClient.fetchLatestProposals.mockResolvedValue(topicList);
      discourseClient.fetchProposalDetails.mockResolvedValueOnce(badDetails).mockResolvedValueOnce(goodDetails);
      normalizationService.normalizeDiscourseProposal
        .mockImplementationOnce(() => {
          throw new Error("Normalization error");
        })
        .mockReturnValueOnce(normalizedInput);

      // Act
      const result = await fetcher.getLatestProposals();

      // Assert
      expect(result).toEqual([normalizedInput]);
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to normalize proposal",
        expect.objectContaining({ topicId: 100 }),
      );
    });
  });
});
