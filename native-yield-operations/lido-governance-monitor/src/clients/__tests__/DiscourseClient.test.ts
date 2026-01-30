import { ILogger, IRetryService } from "@consensys/linea-shared-utils";
import { jest, describe, it, expect, beforeEach } from "@jest/globals";

import { DiscourseClient } from "../DiscourseClient.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

const createRetryServiceMock = (): jest.Mocked<IRetryService> => ({
  retry: jest.fn().mockImplementation(<T>(fn: () => Promise<T>) => fn()),
});

describe("DiscourseClient", () => {
  let client: DiscourseClient;
  let logger: jest.Mocked<ILogger>;
  let retryService: jest.Mocked<IRetryService>;
  let fetchMock: jest.Mock;

  beforeEach(() => {
    logger = createLoggerMock();
    retryService = createRetryServiceMock();
    fetchMock = jest.fn();
    global.fetch = fetchMock as unknown as typeof fetch;
    client = new DiscourseClient(logger, retryService, "https://research.lido.fi/c/proposals/9/l/latest.json");
  });

  describe("fetchLatestProposals", () => {
    it("fetches and returns latest proposals from Discourse API", async () => {
      // Arrange
      const mockResponse = { topic_list: { topics: [{ id: 11107, slug: "test-proposal" }] } };
      fetchMock.mockResolvedValue({ ok: true, json: () => Promise.resolve(mockResponse) });

      // Act
      const result = await client.fetchLatestProposals();

      // Assert
      expect(result).toEqual(mockResponse);
      expect(fetchMock).toHaveBeenCalledWith("https://research.lido.fi/c/proposals/9/l/latest.json");
    });

    it("returns undefined on fetch failure", async () => {
      // Arrange
      fetchMock.mockResolvedValue({ ok: false, status: 500, statusText: "Internal Server Error" });

      // Act
      const result = await client.fetchLatestProposals();

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalled();
    });

    it("returns undefined on network error", async () => {
      // Arrange
      fetchMock.mockRejectedValue(new Error("Network error"));

      // Act
      const result = await client.fetchLatestProposals();

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalled();
    });

    it("uses retry service for fetch call", async () => {
      // Arrange
      const mockResponse = { topic_list: { topics: [] } };
      fetchMock.mockResolvedValue({ ok: true, json: () => Promise.resolve(mockResponse) });

      // Act
      await client.fetchLatestProposals();

      // Assert
      expect(retryService.retry).toHaveBeenCalledTimes(1);
      expect(retryService.retry).toHaveBeenCalledWith(expect.any(Function));
    });
  });

  describe("fetchProposalDetails", () => {
    it("fetches and returns proposal details", async () => {
      // Arrange
      const mockProposal = {
        id: 11107,
        slug: "test-proposal",
        title: "Test Proposal",
        created_at: "2024-01-01T00:00:00.000Z",
        post_stream: {
          posts: [
            {
              id: 24002,
              username: "testuser",
              cooked: "<p>Content</p>",
              post_url: "/t/test/1",
              created_at: "2024-01-01",
            },
          ],
        },
      };
      fetchMock.mockResolvedValue({ ok: true, json: () => Promise.resolve(mockProposal) });

      // Act
      const result = await client.fetchProposalDetails(11107);

      // Assert
      expect(result).toEqual(mockProposal);
      expect(fetchMock).toHaveBeenCalledWith("https://research.lido.fi/t/11107.json");
    });

    it("returns undefined on fetch failure", async () => {
      // Arrange
      fetchMock.mockResolvedValue({ ok: false, status: 404, statusText: "Not Found" });

      // Act
      const result = await client.fetchProposalDetails(99999);

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalled();
    });

    it("returns undefined on network error", async () => {
      // Arrange
      fetchMock.mockRejectedValue(new Error("Network error"));

      // Act
      const result = await client.fetchProposalDetails(11107);

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalled();
    });

    it("uses retry service for fetch call", async () => {
      // Arrange
      const mockProposal = { id: 11107, title: "Test" };
      fetchMock.mockResolvedValue({ ok: true, json: () => Promise.resolve(mockProposal) });

      // Act
      await client.fetchProposalDetails(11107);

      // Assert
      expect(retryService.retry).toHaveBeenCalledTimes(1);
      expect(retryService.retry).toHaveBeenCalledWith(expect.any(Function));
    });
  });
});
