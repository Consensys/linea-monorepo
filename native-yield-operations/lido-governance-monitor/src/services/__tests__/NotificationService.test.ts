import { ILogger } from "@consensys/linea-shared-utils";
import { jest, describe, it, expect, beforeEach, afterEach } from "@jest/globals";

import { ISlackClient } from "../../core/clients/ISlackClient.js";
import { Assessment } from "../../core/entities/Assessment.js";
import { Proposal } from "../../core/entities/Proposal.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { ProposalState } from "../../core/entities/ProposalState.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";
import { NotificationService } from "../NotificationService.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

describe("NotificationService", () => {
  let service: NotificationService;
  let logger: jest.Mocked<ILogger>;
  let slackClient: jest.Mocked<ISlackClient>;
  let proposalRepository: jest.Mocked<IProposalRepository>;

  const createMockAssessment = (): Assessment => ({
    riskScore: 75,
    riskLevel: "high",
    confidence: 85,
    proposalType: "discourse",
    impactTypes: ["technical"],
    affectedComponents: ["StakingVault"],
    whatChanged: "Contract upgrade",
    nativeYieldInvariantsAtRisk: ["A_valid_yield_reporting"],
    whyItMattersForLineaNativeYield: "May affect withdrawals",
    recommendedAction: "escalate",
    urgency: "urgent",
    supportingQuotes: ["quote"],
    keyUnknowns: [],
  });

  const createMockProposal = (overrides: Partial<Proposal> = {}): Proposal => ({
    id: "uuid-1",
    source: ProposalSource.DISCOURSE,
    sourceId: "12345",
    url: "https://research.lido.fi/t/test/12345",
    title: "Test Proposal",
    author: "testuser",
    sourceCreatedAt: new Date("2024-01-15"),
    text: "Proposal content",
    state: ProposalState.ANALYZED,
    createdAt: new Date(),
    updatedAt: new Date(),
    stateUpdatedAt: new Date(),
    analysisAttemptCount: 1,
    llmModel: "claude-sonnet-4",
    riskThreshold: 60,
    assessmentPromptVersion: "v1.0",
    analyzedAt: new Date(),
    assessmentJson: createMockAssessment(),
    riskScore: 75,
    notifyAttemptCount: 0,
    notifiedAt: null,
    slackMessageTs: null,
    ...overrides,
  });

  beforeEach(() => {
    logger = createLoggerMock();
    slackClient = {
      sendProposalAlert: jest.fn(),
      sendAuditLog: jest.fn(),
    } as jest.Mocked<ISlackClient>;
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

    // Default audit log mock to return success
    slackClient.sendAuditLog.mockResolvedValue({ success: true });

    service = new NotificationService(
      logger,
      slackClient,
      proposalRepository,
      60, // riskThreshold
    );
  });

  afterEach(() => {});

  describe("notifyOnce", () => {
    it("fetches ANALYZED and NOTIFY_FAILED proposals from repository", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      await service.notifyOnce();

      // Assert
      expect(proposalRepository.findByState).toHaveBeenCalledWith(ProposalState.ANALYZED);
      expect(proposalRepository.findByState).toHaveBeenCalledWith(ProposalState.NOTIFY_FAILED);
    });

    it("sends Slack notification for each pending proposal", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(proposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue(proposal);

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendProposalAlert).toHaveBeenCalledWith(proposal, proposal.assessmentJson);
    });

    it("marks proposal as NOTIFIED on successful Slack notification", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(proposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue({ ...proposal, state: ProposalState.NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(proposal.id, "ts-123");
      expect(logger.info).toHaveBeenCalledWith("Proposal notification sent", expect.any(Object));
    });

    it("increments attempt count and logs warning when Slack notification fails", async () => {
      // Arrange
      const proposal = createMockProposal({ notifyAttemptCount: 0 });
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue({ ...proposal, notifyAttemptCount: 1 });
      slackClient.sendProposalAlert.mockResolvedValue({ success: false, error: "Webhook failed" });

      // Act
      await service.notifyOnce();

      // Assert
      expect(proposalRepository.incrementNotifyAttempt).toHaveBeenCalledWith(proposal.id);
      expect(logger.warn).toHaveBeenCalledWith("Slack notification failed, will retry", expect.any(Object));
    });

    it("retries NOTIFY_FAILED proposals", async () => {
      // Arrange
      const failedProposal = createMockProposal({
        state: ProposalState.NOTIFY_FAILED,
        notifyAttemptCount: 3,
      });
      proposalRepository.findByState
        .mockResolvedValueOnce([]) // ANALYZED
        .mockResolvedValueOnce([failedProposal]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue({ ...failedProposal, notifyAttemptCount: 4 });
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-456" });
      proposalRepository.markNotified.mockResolvedValue({ ...failedProposal, state: ProposalState.NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendProposalAlert).toHaveBeenCalled();
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(failedProposal.id, "ts-456");
    });

    it("does nothing when no proposals need notification", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendProposalAlert).not.toHaveBeenCalled();
      expect(logger.debug).toHaveBeenCalledWith("No proposals to notify");
    });

    it("skips proposals without assessment data", async () => {
      // Arrange
      const proposalWithoutAssessment = createMockProposal({ assessmentJson: null });
      proposalRepository.findByState
        .mockResolvedValueOnce([proposalWithoutAssessment]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendProposalAlert).not.toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith("Proposal missing assessment data", expect.any(Object));
    });

    it("handles errors during notification gracefully", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockRejectedValue(new Error("Database error"));

      // Act
      await service.notifyOnce();

      // Assert
      expect(logger.error).toHaveBeenCalledWith("Error notifying proposal", expect.any(Object));
    });

    it("uses empty messageTs when not returned from Slack", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(proposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true }); // No messageTs
      proposalRepository.markNotified.mockResolvedValue(proposal);

      // Act
      await service.notifyOnce();

      // Assert
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(proposal.id, "");
    });

    it("catches and logs errors without throwing", async () => {
      // Arrange
      proposalRepository.findByState.mockRejectedValue(new Error("Database connection error"));

      // Act
      await service.notifyOnce();

      // Assert
      expect(logger.error).toHaveBeenCalledWith("Notification processing failed", expect.any(Error));
    });

    it("skips notification for proposals below risk threshold", async () => {
      // Arrange
      const lowRiskProposal = createMockProposal({ riskScore: 30 }); // Below threshold of 60
      proposalRepository.findByState
        .mockResolvedValueOnce([lowRiskProposal]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.updateState.mockResolvedValue({ ...lowRiskProposal, state: ProposalState.NOT_NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendProposalAlert).not.toHaveBeenCalled();
      expect(proposalRepository.updateState).toHaveBeenCalledWith(lowRiskProposal.id, ProposalState.NOT_NOTIFIED);
      expect(logger.info).toHaveBeenCalledWith("Proposal below notification threshold, skipped", expect.any(Object));
    });

    it("sends notification for proposals at or above risk threshold", async () => {
      // Arrange
      const highRiskProposal = createMockProposal({ riskScore: 60 }); // At threshold
      proposalRepository.findByState
        .mockResolvedValueOnce([highRiskProposal]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(highRiskProposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue({ ...highRiskProposal, state: ProposalState.NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendProposalAlert).toHaveBeenCalledWith(highRiskProposal, highRiskProposal.assessmentJson);
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(highRiskProposal.id, "ts-123");
    });

    it("skips notification when riskScore is null", async () => {
      // Arrange
      const proposalWithoutRiskScore = createMockProposal({ riskScore: null });
      proposalRepository.findByState
        .mockResolvedValueOnce([proposalWithoutRiskScore]) // ANALYZED
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.updateState.mockResolvedValue({
        ...proposalWithoutRiskScore,
        state: ProposalState.NOT_NOTIFIED,
      });

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendProposalAlert).not.toHaveBeenCalled();
      expect(proposalRepository.updateState).toHaveBeenCalledWith(
        proposalWithoutRiskScore.id,
        ProposalState.NOT_NOTIFIED,
      );
    });

    it("sends audit log for high-risk proposals", async () => {
      // Arrange
      const highRiskProposal = createMockProposal({ riskScore: 75 });
      proposalRepository.findByState.mockResolvedValueOnce([highRiskProposal]).mockResolvedValueOnce([]);
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(highRiskProposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue({ ...highRiskProposal, state: ProposalState.NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendAuditLog).toHaveBeenCalledWith(highRiskProposal, highRiskProposal.assessmentJson);
    });

    it("sends audit log for low-risk proposals", async () => {
      // Arrange
      const lowRiskProposal = createMockProposal({ riskScore: 30 });
      proposalRepository.findByState.mockResolvedValueOnce([lowRiskProposal]).mockResolvedValueOnce([]);
      proposalRepository.updateState.mockResolvedValue({ ...lowRiskProposal, state: ProposalState.NOT_NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendAuditLog).toHaveBeenCalledWith(lowRiskProposal, lowRiskProposal.assessmentJson);
    });

    it("continues with alert when audit log fails", async () => {
      // Arrange
      const highRiskProposal = createMockProposal({ riskScore: 75 });
      proposalRepository.findByState.mockResolvedValueOnce([highRiskProposal]).mockResolvedValueOnce([]);
      slackClient.sendAuditLog.mockResolvedValue({ success: false, error: "Audit webhook failed" });
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(highRiskProposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue({ ...highRiskProposal, state: ProposalState.NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(logger.warn).toHaveBeenCalledWith("Audit log failed, continuing", expect.any(Object));
      expect(slackClient.sendProposalAlert).toHaveBeenCalledWith(highRiskProposal, highRiskProposal.assessmentJson);
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(highRiskProposal.id, "ts-123");
    });

    it("marks low-risk proposal as NOT_NOTIFIED even if audit log fails", async () => {
      // Arrange
      const lowRiskProposal = createMockProposal({ riskScore: 30 });
      proposalRepository.findByState.mockResolvedValueOnce([lowRiskProposal]).mockResolvedValueOnce([]);
      slackClient.sendAuditLog.mockResolvedValue({ success: false, error: "Audit webhook failed" });
      proposalRepository.updateState.mockResolvedValue({ ...lowRiskProposal, state: ProposalState.NOT_NOTIFIED });

      // Act
      await service.notifyOnce();

      // Assert
      expect(logger.warn).toHaveBeenCalledWith("Audit log failed, continuing", expect.any(Object));
      expect(proposalRepository.updateState).toHaveBeenCalledWith(lowRiskProposal.id, ProposalState.NOT_NOTIFIED);
    });

    it("does not send audit log when proposal missing assessment", async () => {
      // Arrange
      const proposalWithoutAssessment = createMockProposal({ assessmentJson: null });
      proposalRepository.findByState.mockResolvedValueOnce([proposalWithoutAssessment]).mockResolvedValueOnce([]);

      // Act
      await service.notifyOnce();

      // Assert
      expect(slackClient.sendAuditLog).not.toHaveBeenCalled();
    });
  });
});
