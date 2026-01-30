import { jest, describe, it, expect, beforeEach } from "@jest/globals";
import { ProposalProcessor } from "../../services/ProposalProcessor.js";
import { NotificationService } from "../../services/NotificationService.js";
import { ProposalState } from "../../core/entities/ProposalState.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { Proposal } from "../../core/entities/Proposal.js";
import { Assessment } from "../../core/entities/Assessment.js";
import { ILogger } from "@consensys/linea-shared-utils";
import { IAIClient } from "../../core/clients/IAIClient.js";
import { ISlackClient } from "../../core/clients/ISlackClient.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

describe("Proposal Lifecycle Integration", () => {
  let logger: jest.Mocked<ILogger>;
  let aiClient: jest.Mocked<IAIClient>;
  let slackClient: jest.Mocked<ISlackClient>;
  let proposalRepository: jest.Mocked<IProposalRepository>;
  let processor: ProposalProcessor;
  let notificationService: NotificationService;

  const createMockAssessment = (riskScore: number): Assessment => ({
    riskScore,
    riskLevel: riskScore >= 61 ? "high" : riskScore >= 31 ? "medium" : "low",
    confidence: 85,
    proposalType: "discourse",
    impactTypes: ["technical"],
    affectedComponents: ["StakingVault"],
    whatChanged: "Contract upgrade",
    nativeYieldInvariantsAtRisk: riskScore >= 61 ? ["A_valid_yield_reporting"] : [],
    whyItMattersForLineaNativeYield: "May affect withdrawals",
    recommendedAction: riskScore >= 71 ? "escalate" : riskScore >= 51 ? "comment" : "monitor",
    urgency: riskScore >= 71 ? "pre_execution" : "none",
    supportingQuotes: ["quote"],
    keyUnknowns: [],
  });

  const createMockProposal = (state: ProposalState, assessment?: Assessment): Proposal => ({
    id: "uuid-1",
    source: ProposalSource.DISCOURSE,
    sourceId: "12345",
    url: "https://research.lido.fi/t/test/12345",
    title: "Test Proposal",
    author: "testuser",
    sourceCreatedAt: new Date("2024-01-15"),
    text: "Proposal content",
    state,
    createdAt: new Date(),
    updatedAt: new Date(),
    stateUpdatedAt: new Date(),
    analysisAttemptCount: 0,
    llmModel: assessment ? "claude-sonnet-4" : null,
    riskThreshold: assessment ? 60 : null,
    assessmentPromptVersion: assessment ? "v1.0" : null,
    analyzedAt: assessment ? new Date() : null,
    assessmentJson: assessment ?? null,
    riskScore: assessment?.riskScore ?? null,
    notifyAttemptCount: 0,
    notifiedAt: null,
    slackMessageTs: null,
  });

  beforeEach(() => {
    logger = createLoggerMock();
    aiClient = {
      analyzeProposal: jest.fn(),
      getModelName: jest.fn().mockReturnValue("claude-sonnet-4"),
    } as jest.Mocked<IAIClient>;
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

    processor = new ProposalProcessor(
      logger,
      aiClient,
      proposalRepository,
      60, // riskThreshold
      "v1.0",
      "Domain context",
      3,
      60000
    );

    notificationService = new NotificationService(logger, slackClient, proposalRepository, 3, 60000);
  });

  describe("Full lifecycle: NEW → ANALYZED → PENDING_NOTIFY → NOTIFIED", () => {
    it("processes high-risk proposal through complete lifecycle", async () => {
      // Arrange
      const highRiskAssessment = createMockAssessment(75);
      const newProposal = createMockProposal(ProposalState.NEW);
      const analyzedProposal = createMockProposal(ProposalState.ANALYZED, highRiskAssessment);
      const pendingNotifyProposal = createMockProposal(ProposalState.PENDING_NOTIFY, highRiskAssessment);
      const notifiedProposal = createMockProposal(ProposalState.NOTIFIED, highRiskAssessment);

      // Phase 1: Process NEW proposal
      proposalRepository.findByState.mockResolvedValueOnce([newProposal]);
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue({ ...newProposal, analysisAttemptCount: 1 });
      aiClient.analyzeProposal.mockResolvedValue(highRiskAssessment);
      proposalRepository.saveAnalysis.mockResolvedValue(analyzedProposal);
      proposalRepository.updateState.mockResolvedValueOnce(pendingNotifyProposal);

      // Act - Phase 1
      await processor.processOnce();

      // Assert - Phase 1
      expect(aiClient.analyzeProposal).toHaveBeenCalled();
      expect(proposalRepository.saveAnalysis).toHaveBeenCalledWith(
        newProposal.id,
        highRiskAssessment,
        75,
        "claude-sonnet-4",
        60,
        "v1.0"
      );
      expect(proposalRepository.updateState).toHaveBeenCalledWith(newProposal.id, ProposalState.PENDING_NOTIFY);

      // Phase 2: Notify PENDING_NOTIFY proposal
      proposalRepository.findByState.mockResolvedValueOnce([pendingNotifyProposal]);
      proposalRepository.incrementNotifyAttempt.mockResolvedValue({ ...pendingNotifyProposal, notifyAttemptCount: 1 });
      slackClient.sendProposalAlert.mockResolvedValue({ success: true, messageTs: "ts-123" });
      proposalRepository.markNotified.mockResolvedValue(notifiedProposal);

      // Act - Phase 2
      await notificationService.processOnce();

      // Assert - Phase 2
      expect(slackClient.sendProposalAlert).toHaveBeenCalledWith(pendingNotifyProposal, highRiskAssessment);
      expect(proposalRepository.markNotified).toHaveBeenCalledWith(pendingNotifyProposal.id, "ts-123");
    });
  });

  describe("Low-risk proposal lifecycle: NEW → ANALYZED → NOT_NOTIFIED", () => {
    it("processes low-risk proposal and skips notification", async () => {
      // Arrange
      const lowRiskAssessment = createMockAssessment(30);
      const newProposal = createMockProposal(ProposalState.NEW);
      const analyzedProposal = createMockProposal(ProposalState.ANALYZED, lowRiskAssessment);
      const notNotifiedProposal = { ...analyzedProposal, state: ProposalState.NOT_NOTIFIED };

      proposalRepository.findByState.mockResolvedValueOnce([newProposal]);
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue({ ...newProposal, analysisAttemptCount: 1 });
      aiClient.analyzeProposal.mockResolvedValue(lowRiskAssessment);
      proposalRepository.saveAnalysis.mockResolvedValue(analyzedProposal);
      proposalRepository.updateState.mockResolvedValue(notNotifiedProposal);

      // Act
      await processor.processOnce();

      // Assert
      expect(proposalRepository.updateState).toHaveBeenCalledWith(newProposal.id, ProposalState.NOT_NOTIFIED);
      expect(logger.info).toHaveBeenCalledWith("Proposal below notification threshold", expect.any(Object));
    });
  });

  describe("Analysis failure lifecycle: NEW → ANALYSIS_FAILED", () => {
    it("transitions to ANALYSIS_FAILED after max attempts", async () => {
      // Arrange
      const newProposal = createMockProposal(ProposalState.NEW);

      proposalRepository.findByState.mockResolvedValue([newProposal]);
      aiClient.analyzeProposal.mockResolvedValue(undefined); // Analysis fails

      // Simulate 3 failed attempts
      for (let attempt = 1; attempt <= 3; attempt++) {
        proposalRepository.incrementAnalysisAttempt.mockResolvedValueOnce({
          ...newProposal,
          analysisAttemptCount: attempt,
        });
      }
      proposalRepository.updateState.mockResolvedValue({ ...newProposal, state: ProposalState.ANALYSIS_FAILED });

      // Act - 3 processing cycles
      await processor.processOnce();
      await processor.processOnce();
      await processor.processOnce();

      // Assert
      expect(proposalRepository.updateState).toHaveBeenCalledWith(newProposal.id, ProposalState.ANALYSIS_FAILED);
      expect(logger.error).toHaveBeenCalledWith("Analysis failed after max attempts", expect.any(Object));
    });
  });

  describe("Notification failure lifecycle: PENDING_NOTIFY → NOTIFY_FAILED", () => {
    it("transitions to NOTIFY_FAILED after max attempts", async () => {
      // Arrange
      const assessment = createMockAssessment(75);
      const pendingProposal = createMockProposal(ProposalState.PENDING_NOTIFY, assessment);

      proposalRepository.findByState.mockResolvedValue([pendingProposal]);
      slackClient.sendProposalAlert.mockResolvedValue({ success: false, error: "Webhook failed" });

      // Simulate 3 failed attempts
      for (let attempt = 1; attempt <= 3; attempt++) {
        proposalRepository.incrementNotifyAttempt.mockResolvedValueOnce({
          ...pendingProposal,
          notifyAttemptCount: attempt,
        });
      }
      proposalRepository.updateState.mockResolvedValue({ ...pendingProposal, state: ProposalState.NOTIFY_FAILED });

      // Act - 3 notification cycles
      await notificationService.processOnce();
      await notificationService.processOnce();
      await notificationService.processOnce();

      // Assert
      expect(proposalRepository.updateState).toHaveBeenCalledWith(pendingProposal.id, ProposalState.NOTIFY_FAILED);
      expect(logger.error).toHaveBeenCalledWith("Notification failed after max attempts", expect.any(Object));
    });
  });
});
