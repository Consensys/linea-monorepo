import { jest, describe, it, expect } from "@jest/globals";

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

describe("LdoVotingContractFetcher", () => {
  it("returns empty array from getLatestProposals", async () => {
    // Arrange
    const logger = createLoggerMock();
    const fetcher = new LdoVotingContractFetcher(logger);

    // Act
    const result = await fetcher.getLatestProposals();

    // Assert
    expect(result).toEqual([]);
  });
});
