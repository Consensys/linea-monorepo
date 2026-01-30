import { jest, describe, it, expect, beforeEach } from "@jest/globals";
import { SlackClient } from "../SlackClient.js";
import { ILogger } from "@consensys/linea-shared-utils";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { ProposalState } from "../../core/entities/ProposalState.js";
import { Proposal } from "../../core/entities/Proposal.js";
import { Assessment } from "../../core/entities/Assessment.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

describe("SlackClient", () => {
  let client: SlackClient;
  let logger: jest.Mocked<ILogger>;
  let fetchMock: jest.Mock;

  const createMockProposal = (overrides: Partial<Proposal> = {}): Proposal => ({
    id: "uuid-1",
    source: ProposalSource.DISCOURSE,
    sourceId: "12345",
    url: "https://research.lido.fi/t/test/12345",
    title: "Test Proposal",
    author: "testuser",
    sourceCreatedAt: new Date("2024-01-15"),
    text: "Proposal content",
    state: ProposalState.PENDING_NOTIFY,
    createdAt: new Date(),
    updatedAt: new Date(),
    stateUpdatedAt: new Date(),
    analysisAttemptCount: 1,
    llmModel: "claude-sonnet-4",
    riskThreshold: 60,
    assessmentPromptVersion: "v1.0",
    analyzedAt: new Date(),
    assessmentJson: null,
    riskScore: 75,
    notifyAttemptCount: 0,
    notifiedAt: null,
    slackMessageTs: null,
    ...overrides,
  });

  const createMockAssessment = (overrides: Partial<Assessment> = {}): Assessment => ({
    riskScore: 75,
    riskLevel: "high",
    confidence: 85,
    proposalType: "discourse",
    impactTypes: ["technical"],
    affectedComponents: ["StakingVault"],
    whatChanged: "Contract upgrade to StakingVault v2",
    nativeYieldInvariantsAtRisk: ["A_valid_yield_reporting"],
    whyItMattersForLineaNativeYield: "May affect withdrawal mechanics and yield reporting",
    recommendedAction: "escalate",
    urgency: "pre_execution",
    supportingQuotes: ["The upgrade will modify the withdrawal queue..."],
    keyUnknowns: [],
    ...overrides,
  });

  beforeEach(() => {
    logger = createLoggerMock();
    fetchMock = jest.fn();
    global.fetch = fetchMock as unknown as typeof fetch;
    client = new SlackClient(logger, "https://hooks.slack.com/services/xxx");
  });

  describe("sendProposalAlert", () => {
    it("sends formatted message to Slack webhook and returns success", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const mockAssessment = createMockAssessment();
      fetchMock.mockResolvedValue({ ok: true, text: () => Promise.resolve("ok") });

      // Act
      const result = await client.sendProposalAlert(mockProposal, mockAssessment);

      // Assert
      expect(result.success).toBe(true);
      expect(fetchMock).toHaveBeenCalledWith(
        "https://hooks.slack.com/services/xxx",
        expect.objectContaining({
          method: "POST",
          headers: { "Content-Type": "application/json" },
        })
      );
    });

    it("includes risk level emoji based on severity", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const criticalAssessment = createMockAssessment({ riskLevel: "critical", riskScore: 90 });
      fetchMock.mockResolvedValue({ ok: true, text: () => Promise.resolve("ok") });

      // Act
      await client.sendProposalAlert(mockProposal, criticalAssessment);

      // Assert
      const callBody = JSON.parse(fetchMock.mock.calls[0][1].body);
      expect(callBody.blocks[0].text.text).toContain(":rotating_light:");
    });

    it("includes affected components and invariants in message", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const mockAssessment = createMockAssessment({
        affectedComponents: ["StakingVault", "VaultHub"],
        nativeYieldInvariantsAtRisk: ["A_valid_yield_reporting", "B_user_principal_protection"],
      });
      fetchMock.mockResolvedValue({ ok: true, text: () => Promise.resolve("ok") });

      // Act
      await client.sendProposalAlert(mockProposal, mockAssessment);

      // Assert
      const callBody = JSON.parse(fetchMock.mock.calls[0][1].body);
      const bodyString = JSON.stringify(callBody);
      expect(bodyString).toContain("StakingVault");
      expect(bodyString).toContain("VaultHub");
    });

    it("returns failure on webhook error response", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const mockAssessment = createMockAssessment();
      fetchMock.mockResolvedValue({ ok: false, status: 500, text: () => Promise.resolve("internal_error") });

      // Act
      const result = await client.sendProposalAlert(mockProposal, mockAssessment);

      // Assert
      expect(result.success).toBe(false);
      expect(result.error).toBe("internal_error");
      expect(logger.error).toHaveBeenCalled();
    });

    it("returns failure on network error", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const mockAssessment = createMockAssessment();
      fetchMock.mockRejectedValue(new Error("Network error"));

      // Act
      const result = await client.sendProposalAlert(mockProposal, mockAssessment);

      // Assert
      expect(result.success).toBe(false);
      expect(result.error).toBe("Network error");
      expect(logger.error).toHaveBeenCalled();
    });

    it("uses warning emoji for high risk level", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const highAssessment = createMockAssessment({ riskLevel: "high" });
      fetchMock.mockResolvedValue({ ok: true, text: () => Promise.resolve("ok") });

      // Act
      await client.sendProposalAlert(mockProposal, highAssessment);

      // Assert
      const callBody = JSON.parse(fetchMock.mock.calls[0][1].body);
      expect(callBody.blocks[0].text.text).toContain(":warning:");
    });

    it("uses info emoji for medium risk level", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const mediumAssessment = createMockAssessment({ riskLevel: "medium", riskScore: 50 });
      fetchMock.mockResolvedValue({ ok: true, text: () => Promise.resolve("ok") });

      // Act
      await client.sendProposalAlert(mockProposal, mediumAssessment);

      // Assert
      const callBody = JSON.parse(fetchMock.mock.calls[0][1].body);
      expect(callBody.blocks[0].text.text).toContain(":information_source:");
    });

    it("uses info emoji for low risk level", async () => {
      // Arrange
      const mockProposal = createMockProposal();
      const lowAssessment = createMockAssessment({ riskLevel: "low", riskScore: 25 });
      fetchMock.mockResolvedValue({ ok: true, text: () => Promise.resolve("ok") });

      // Act
      await client.sendProposalAlert(mockProposal, lowAssessment);

      // Assert
      const callBody = JSON.parse(fetchMock.mock.calls[0][1].body);
      expect(callBody.blocks[0].text.text).toContain(":information_source:");
    });
  });
});
