import { jest, describe, it, expect, beforeEach, afterEach } from "@jest/globals";
import { NotificationService } from "../NotificationService.js";
import { ILogger } from "@consensys/linea-shared-utils";
import { ISlackClient, SlackNotificationResult } from "../../core/clients/ISlackClient.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";
import { ProposalState } from "../../core/entities/ProposalState.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { Proposal } from "../../core/entities/Proposal.js";
import { Assessment } from "../../core/entities/Assessment.js";

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
    urgency: "pre_execution",
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
    state: ProposalState.PENDING_NOTIFY,
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
    jest.useFakeTimers();
    logger = createLoggerMock();
    slackClient = {
      sendProposalAlert: jest.fn(),
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

    service = new NotificationService(
      logger,
      slackClient,
      proposalRepository,
      60000 // processingIntervalMs
    );
  });

  afterEach(() => {
    service.stop();
    jest.useRealTimers();
  });

  describe("processOnce", () => {
    it("fetches PENDING_NOTIFY and NOTIFY_FAILED proposals from repository", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      await service.processOnce();

      // Assert
      expect(proposalRepository.findByState).toHaveBeenCalledWith(ProposalState.PENDING_NOTIFY);
      expect(proposalRepository.findByState).toHaveBeenCalledWith(ProposalState.NOTIFY_FAILED);
    });

    it("sends Slack notification for each pending proposal", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // PENDING_NOTIFY
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(proposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue(proposal);

      // Act
      await service.processOnce();

      // Assert
      expect(slackClient.sendProposalAlert).toHaveBeenCalledWith(proposal, proposal.assessmentJson);
    });

    it("marks proposal as NOTIFIED on successful Slack notification", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // PENDING_NOTIFY
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(proposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue({ ...proposal, state: ProposalState.NOTIFIED });

      // Act
      await service.processOnce();

      // Assert
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(proposal.id, "ts-123");
      expect(logger.info).toHaveBeenCalledWith("Proposal notification sent", expect.any(Object));
    });

    it("increments attempt count and logs warning when Slack notification fails", async () => {
      // Arrange
      const proposal = createMockProposal({ notifyAttemptCount: 0 });
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // PENDING_NOTIFY
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue({ ...proposal, notifyAttemptCount: 1 });
      slackClient.sendProposalAlert.mockResolvedValue({ success: false, error: "Webhook failed" });

      // Act
      await service.processOnce();

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
        .mockResolvedValueOnce([]) // PENDING_NOTIFY
        .mockResolvedValueOnce([failedProposal]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue({ ...failedProposal, notifyAttemptCount: 4 });
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-456" });
      proposalRepository.markNotified.mockResolvedValue({ ...failedProposal, state: ProposalState.NOTIFIED });

      // Act
      await service.processOnce();

      // Assert
      expect(slackClient.sendProposalAlert).toHaveBeenCalled();
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(failedProposal.id, "ts-456");
    });

    it("does nothing when no proposals need notification", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      await service.processOnce();

      // Assert
      expect(slackClient.sendProposalAlert).not.toHaveBeenCalled();
      expect(logger.debug).toHaveBeenCalledWith("No proposals to notify");
    });

    it("skips proposals without assessment data", async () => {
      // Arrange
      const proposalWithoutAssessment = createMockProposal({ assessmentJson: null });
      proposalRepository.findByState
        .mockResolvedValueOnce([proposalWithoutAssessment]) // PENDING_NOTIFY
        .mockResolvedValueOnce([]); // NOTIFY_FAILED

      // Act
      await service.processOnce();

      // Assert
      expect(slackClient.sendProposalAlert).not.toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith("Proposal missing assessment data", expect.any(Object));
    });

    it("handles errors during notification gracefully", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // PENDING_NOTIFY
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockRejectedValue(new Error("Database error"));

      // Act
      await service.processOnce();

      // Assert
      expect(logger.error).toHaveBeenCalledWith("Error notifying proposal", expect.any(Object));
    });

    it("uses empty messageTs when not returned from Slack", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // PENDING_NOTIFY
        .mockResolvedValueOnce([]); // NOTIFY_FAILED
      proposalRepository.incrementNotifyAttempt.mockResolvedValue(proposal);
      slackClient.sendProposalAlert.mockResolvedValue({ success: true }); // No messageTs
      proposalRepository.markNotified.mockResolvedValue(proposal);

      // Act
      await service.processOnce();

      // Assert
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(proposal.id, "");
    });
  });

  describe("notifyProposal", () => {
    it("returns true when Slack notification succeeds", async () => {
      // Arrange
      const proposal = createMockProposal();
      const assessment = createMockAssessment();
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });

      // Act
      const result = await service.notifyProposal(proposal, assessment);

      // Assert
      expect(result).toBe(true);
      expect(slackClient.sendProposalAlert).toHaveBeenCalledWith(proposal, assessment);
    });

    it("returns false when Slack notification fails", async () => {
      // Arrange
      const proposal = createMockProposal();
      const assessment = createMockAssessment();
      slackClient.sendProposalAlert.mockResolvedValue({ success: false, error: "Failed" });

      // Act
      const result = await service.notifyProposal(proposal, assessment);

      // Assert
      expect(result).toBe(false);
    });
  });

  describe("start and stop", () => {
    it("starts processing at configured interval", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      service.start();

      // Allow initial async processOnce to complete
      await jest.advanceTimersByTimeAsync(0);

      // Assert - initial process (calls findByState twice: PENDING_NOTIFY and NOTIFY_FAILED)
      expect(proposalRepository.findByState).toHaveBeenCalledTimes(2);

      // Advance timer and check subsequent process
      await jest.advanceTimersByTimeAsync(60000);
      expect(proposalRepository.findByState).toHaveBeenCalledTimes(4);
    });

    it("stops processing when stop is called", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      service.start();

      // Allow initial async processOnce to complete
      await jest.advanceTimersByTimeAsync(0);

      service.stop();
      await jest.advanceTimersByTimeAsync(60000);

      // Assert - only the initial process should have happened (2 calls: PENDING_NOTIFY and NOTIFY_FAILED)
      expect(proposalRepository.findByState).toHaveBeenCalledTimes(2);
    });

    it("logs when starting and stopping", () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      service.start();
      service.stop();

      // Assert
      expect(logger.info).toHaveBeenCalledWith("NotificationService started", expect.any(Object));
      expect(logger.info).toHaveBeenCalledWith("NotificationService stopped");
    });
  });
});
