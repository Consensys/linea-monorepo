import { ILogger } from "@consensys/linea-shared-utils";
import { jest, describe, it, expect, beforeEach, afterEach } from "@jest/globals";

import { IAIClient } from "../../core/clients/IAIClient.js";
import { Assessment } from "../../core/entities/Assessment.js";
import { Proposal } from "../../core/entities/Proposal.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { ProposalState } from "../../core/entities/ProposalState.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";
import { ProposalProcessor } from "../ProposalProcessor.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

describe("ProposalProcessor", () => {
  let processor: ProposalProcessor;
  let logger: jest.Mocked<ILogger>;
  let aiClient: jest.Mocked<IAIClient>;
  let proposalRepository: jest.Mocked<IProposalRepository>;

  const createMockProposal = (overrides: Partial<Proposal> = {}): Proposal => ({
    id: "uuid-1",
    source: ProposalSource.DISCOURSE,
    sourceId: "12345",
    url: "https://research.lido.fi/t/test/12345",
    title: "Test Proposal",
    author: "testuser",
    sourceCreatedAt: new Date("2024-01-15"),
    text: "Proposal content for analysis",
    state: ProposalState.NEW,
    createdAt: new Date(),
    updatedAt: new Date(),
    stateUpdatedAt: new Date(),
    analysisAttemptCount: 0,
    llmModel: null,
    riskThreshold: null,
    assessmentPromptVersion: null,
    analyzedAt: null,
    assessmentJson: null,
    riskScore: null,
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
    whatChanged: "Contract upgrade",
    nativeYieldInvariantsAtRisk: ["A_valid_yield_reporting"],
    whyItMattersForLineaNativeYield: "May affect withdrawals",
    recommendedAction: "escalate",
    urgency: "pre_execution",
    supportingQuotes: ["quote"],
    keyUnknowns: [],
    ...overrides,
  });

  beforeEach(() => {
    jest.useFakeTimers();
    logger = createLoggerMock();
    aiClient = {
      analyzeProposal: jest.fn(),
      getModelName: jest.fn().mockReturnValue("claude-sonnet-4"),
    } as jest.Mocked<IAIClient>;
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
      "v1.0", // promptVersion
      "Domain context", // domainContext
      60000, // processingIntervalMs
    );
  });

  afterEach(() => {
    processor.stop();
    jest.useRealTimers();
  });

  describe("processOnce", () => {
    it("fetches NEW and ANALYSIS_FAILED proposals from repository", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      await processor.processOnce();

      // Assert
      expect(proposalRepository.findByState).toHaveBeenCalledWith(ProposalState.NEW);
      expect(proposalRepository.findByState).toHaveBeenCalledWith(ProposalState.ANALYSIS_FAILED);
    });

    it("analyzes each NEW proposal with AI client", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // NEW
        .mockResolvedValueOnce([]); // ANALYSIS_FAILED
      aiClient.analyzeProposal.mockResolvedValue(createMockAssessment());
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue(proposal);
      proposalRepository.saveAnalysis.mockResolvedValue(proposal);

      // Act
      await processor.processOnce();

      // Assert
      expect(aiClient.analyzeProposal).toHaveBeenCalledWith({
        proposalTitle: proposal.title,
        proposalText: proposal.text,
        proposalUrl: proposal.url,
        proposalType: "discourse",
        domainContext: "Domain context",
      });
    });

    it("saves analysis and transitions to PENDING_NOTIFY when riskScore >= threshold", async () => {
      // Arrange
      const proposal = createMockProposal();
      const assessment = createMockAssessment({ riskScore: 75 }); // Above 60 threshold
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // NEW
        .mockResolvedValueOnce([]); // ANALYSIS_FAILED
      aiClient.analyzeProposal.mockResolvedValue(assessment);
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue(proposal);
      proposalRepository.saveAnalysis.mockResolvedValue({ ...proposal, state: ProposalState.ANALYZED });
      proposalRepository.updateState.mockResolvedValue({ ...proposal, state: ProposalState.PENDING_NOTIFY });

      // Act
      await processor.processOnce();

      // Assert
      expect(proposalRepository.saveAnalysis).toHaveBeenCalledWith(
        proposal.id,
        assessment,
        75,
        "claude-sonnet-4",
        60,
        "v1.0",
      );
      expect(proposalRepository.updateState).toHaveBeenCalledWith(proposal.id, ProposalState.PENDING_NOTIFY);
      expect(logger.info).toHaveBeenCalledWith("Proposal requires notification", expect.any(Object));
    });

    it("saves analysis and transitions to NOT_NOTIFIED when riskScore < threshold", async () => {
      // Arrange
      const proposal = createMockProposal();
      const assessment = createMockAssessment({ riskScore: 30 }); // Below 60 threshold
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // NEW
        .mockResolvedValueOnce([]); // ANALYSIS_FAILED
      aiClient.analyzeProposal.mockResolvedValue(assessment);
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue(proposal);
      proposalRepository.saveAnalysis.mockResolvedValue({ ...proposal, state: ProposalState.ANALYZED });
      proposalRepository.updateState.mockResolvedValue({ ...proposal, state: ProposalState.NOT_NOTIFIED });

      // Act
      await processor.processOnce();

      // Assert
      expect(proposalRepository.updateState).toHaveBeenCalledWith(proposal.id, ProposalState.NOT_NOTIFIED);
      expect(logger.info).toHaveBeenCalledWith("Proposal below notification threshold", expect.any(Object));
    });

    it("increments attempt count and logs warning when AI analysis fails", async () => {
      // Arrange
      const proposal = createMockProposal({ analysisAttemptCount: 0 });
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // NEW
        .mockResolvedValueOnce([]); // ANALYSIS_FAILED
      aiClient.analyzeProposal.mockResolvedValue(undefined);
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue({ ...proposal, analysisAttemptCount: 1 });

      // Act
      await processor.processOnce();

      // Assert
      expect(proposalRepository.incrementAnalysisAttempt).toHaveBeenCalledWith(proposal.id);
      expect(logger.warn).toHaveBeenCalledWith("AI analysis failed, will retry", expect.any(Object));
    });

    it("retries ANALYSIS_FAILED proposals", async () => {
      // Arrange
      const failedProposal = createMockProposal({
        state: ProposalState.ANALYSIS_FAILED,
        analysisAttemptCount: 3,
      });
      const assessment = createMockAssessment({ riskScore: 75 });
      proposalRepository.findByState
        .mockResolvedValueOnce([]) // NEW
        .mockResolvedValueOnce([failedProposal]); // ANALYSIS_FAILED
      aiClient.analyzeProposal.mockResolvedValue(assessment);
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue({ ...failedProposal, analysisAttemptCount: 4 });
      proposalRepository.saveAnalysis.mockResolvedValue({ ...failedProposal, state: ProposalState.ANALYZED });
      proposalRepository.updateState.mockResolvedValue({ ...failedProposal, state: ProposalState.PENDING_NOTIFY });

      // Act
      await processor.processOnce();

      // Assert
      expect(aiClient.analyzeProposal).toHaveBeenCalled();
      expect(proposalRepository.saveAnalysis).toHaveBeenCalled();
      expect(proposalRepository.updateState).toHaveBeenCalledWith(failedProposal.id, ProposalState.PENDING_NOTIFY);
    });

    it("does nothing when no proposals need processing", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      await processor.processOnce();

      // Assert
      expect(aiClient.analyzeProposal).not.toHaveBeenCalled();
      expect(logger.debug).toHaveBeenCalledWith("No proposals to process");
    });

    it("handles errors during processing gracefully", async () => {
      // Arrange
      const proposal = createMockProposal();
      proposalRepository.findByState
        .mockResolvedValueOnce([proposal]) // NEW
        .mockResolvedValueOnce([]); // ANALYSIS_FAILED
      proposalRepository.incrementAnalysisAttempt.mockRejectedValue(new Error("Database error"));

      // Act
      await processor.processOnce();

      // Assert
      expect(logger.error).toHaveBeenCalledWith("Error processing proposal", expect.any(Object));
    });

    it("maps proposal source to proposalType correctly for snapshot", async () => {
      // Arrange
      const snapshotProposal = createMockProposal({ source: ProposalSource.SNAPSHOT });
      proposalRepository.findByState
        .mockResolvedValueOnce([snapshotProposal]) // NEW
        .mockResolvedValueOnce([]); // ANALYSIS_FAILED
      aiClient.analyzeProposal.mockResolvedValue(createMockAssessment());
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue(snapshotProposal);
      proposalRepository.saveAnalysis.mockResolvedValue(snapshotProposal);

      // Act
      await processor.processOnce();

      // Assert
      expect(aiClient.analyzeProposal).toHaveBeenCalledWith(expect.objectContaining({ proposalType: "snapshot" }));
    });

    it("maps onchain voting contract sources to onchain_vote type", async () => {
      // Arrange
      const ldoProposal = createMockProposal({ source: ProposalSource.LDO_VOTING_CONTRACT });
      proposalRepository.findByState
        .mockResolvedValueOnce([ldoProposal]) // NEW
        .mockResolvedValueOnce([]); // ANALYSIS_FAILED
      aiClient.analyzeProposal.mockResolvedValue(createMockAssessment());
      proposalRepository.incrementAnalysisAttempt.mockResolvedValue(ldoProposal);
      proposalRepository.saveAnalysis.mockResolvedValue(ldoProposal);

      // Act
      await processor.processOnce();

      // Assert
      expect(aiClient.analyzeProposal).toHaveBeenCalledWith(expect.objectContaining({ proposalType: "onchain_vote" }));
    });
  });

  describe("start and stop", () => {
    it("starts processing at configured interval", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      processor.start();

      // Allow initial async processOnce to complete
      await jest.advanceTimersByTimeAsync(0);

      // Assert - initial process (calls findByState twice: NEW and ANALYSIS_FAILED)
      expect(proposalRepository.findByState).toHaveBeenCalledTimes(2);

      // Advance timer and check subsequent process
      await jest.advanceTimersByTimeAsync(60000);
      expect(proposalRepository.findByState).toHaveBeenCalledTimes(4);
    });

    it("stops processing when stop is called", async () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      processor.start();

      // Allow initial async processOnce to complete
      await jest.advanceTimersByTimeAsync(0);

      processor.stop();
      await jest.advanceTimersByTimeAsync(60000);

      // Assert - only the initial process should have happened (2 calls: NEW and ANALYSIS_FAILED)
      expect(proposalRepository.findByState).toHaveBeenCalledTimes(2);
    });

    it("logs when starting and stopping", () => {
      // Arrange
      proposalRepository.findByState.mockResolvedValue([]);

      // Act
      processor.start();
      processor.stop();

      // Assert
      expect(logger.info).toHaveBeenCalledWith("ProposalProcessor started", expect.any(Object));
      expect(logger.info).toHaveBeenCalledWith("ProposalProcessor stopped");
    });
  });
});
