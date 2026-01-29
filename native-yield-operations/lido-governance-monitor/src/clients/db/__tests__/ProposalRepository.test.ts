import { jest, describe, it, expect, beforeEach } from "@jest/globals";
import { ProposalRepository } from "../ProposalRepository.js";
import { ProposalSource } from "../../../core/entities/ProposalSource.js";
import { ProposalState } from "../../../core/entities/ProposalState.js";
import { PrismaClient } from "@prisma/client";

const mockPrisma = {
  proposal: {
    findUnique: jest.fn(),
    findMany: jest.fn(),
    create: jest.fn(),
    update: jest.fn(),
  },
};

describe("ProposalRepository", () => {
  let repository: ProposalRepository;

  beforeEach(() => {
    jest.clearAllMocks();
    repository = new ProposalRepository(mockPrisma as unknown as PrismaClient);
  });

  describe("findBySourceAndSourceId", () => {
    it("returns proposal when found", async () => {
      const mockProposal = { id: "uuid-1", source: "DISCOURSE", sourceId: "12345", state: "NEW" };
      mockPrisma.proposal.findUnique.mockResolvedValue(mockProposal);

      const result = await repository.findBySourceAndSourceId(ProposalSource.DISCOURSE, "12345");

      expect(result).toEqual(mockProposal);
      expect(mockPrisma.proposal.findUnique).toHaveBeenCalledWith({
        where: { source_sourceId: { source: "DISCOURSE", sourceId: "12345" } },
      });
    });

    it("returns null when not found", async () => {
      mockPrisma.proposal.findUnique.mockResolvedValue(null);
      const result = await repository.findBySourceAndSourceId(ProposalSource.DISCOURSE, "99999");
      expect(result).toBeNull();
    });
  });

  describe("findByState", () => {
    it("returns proposals with matching state ordered by stateUpdatedAt", async () => {
      const mockProposals = [{ id: "1", state: "NEW" }, { id: "2", state: "NEW" }];
      mockPrisma.proposal.findMany.mockResolvedValue(mockProposals);

      const result = await repository.findByState(ProposalState.NEW);

      expect(result).toEqual(mockProposals);
      expect(mockPrisma.proposal.findMany).toHaveBeenCalledWith({
        where: { state: "NEW" },
        orderBy: { stateUpdatedAt: "asc" },
      });
    });
  });

  describe("create", () => {
    it("creates a new proposal with NEW state", async () => {
      const input = {
        source: ProposalSource.DISCOURSE,
        sourceId: "12345",
        url: "https://research.lido.fi/t/12345",
        title: "Test Proposal",
        author: "testuser",
        sourceCreatedAt: new Date("2024-01-15"),
        text: "Proposal content",
      };
      mockPrisma.proposal.create.mockResolvedValue({ id: "new-uuid", ...input, state: "NEW" });

      const result = await repository.create(input);

      expect(result.state).toBe("NEW");
      expect(mockPrisma.proposal.create).toHaveBeenCalled();
    });
  });

  describe("updateState", () => {
    it("updates proposal state and stateUpdatedAt", async () => {
      mockPrisma.proposal.update.mockResolvedValue({ id: "uuid-1", state: "ANALYZED" });

      const result = await repository.updateState("uuid-1", ProposalState.ANALYZED);

      expect(result.state).toBe("ANALYZED");
      expect(mockPrisma.proposal.update).toHaveBeenCalledWith({
        where: { id: "uuid-1" },
        data: { state: "ANALYZED", stateUpdatedAt: expect.any(Date) },
      });
    });
  });

  describe("saveAnalysis", () => {
    it("saves assessment and transitions to ANALYZED state", async () => {
      const assessment = {
        riskScore: 75,
        impactType: "technical" as const,
        riskLevel: "high" as const,
        whatChanged: "Contract upgrade",
        whyItMattersForLineaNativeYield: "May affect withdrawals",
        recommendedAction: "escalate" as const,
        supportingQuotes: ["quote1"],
      };
      mockPrisma.proposal.update.mockResolvedValue({ id: "uuid-1", state: "ANALYZED" });

      await repository.saveAnalysis("uuid-1", assessment, 75, "claude-sonnet-4", 60, "v1.0");

      expect(mockPrisma.proposal.update).toHaveBeenCalledWith({
        where: { id: "uuid-1" },
        data: expect.objectContaining({
          state: "ANALYZED",
          assessmentJson: assessment,
          riskScore: 75,
          llmModel: "claude-sonnet-4",
        }),
      });
    });
  });

  describe("incrementAnalysisAttempt", () => {
    it("increments the analysis attempt count", async () => {
      mockPrisma.proposal.update.mockResolvedValue({ id: "uuid-1", analysisAttemptCount: 2 });

      await repository.incrementAnalysisAttempt("uuid-1");

      expect(mockPrisma.proposal.update).toHaveBeenCalledWith({
        where: { id: "uuid-1" },
        data: { analysisAttemptCount: { increment: 1 } },
      });
    });
  });

  describe("incrementNotifyAttempt", () => {
    it("increments the notify attempt count", async () => {
      mockPrisma.proposal.update.mockResolvedValue({ id: "uuid-1", notifyAttemptCount: 2 });

      await repository.incrementNotifyAttempt("uuid-1");

      expect(mockPrisma.proposal.update).toHaveBeenCalledWith({
        where: { id: "uuid-1" },
        data: { notifyAttemptCount: { increment: 1 } },
      });
    });
  });

  describe("markNotified", () => {
    it("marks proposal as NOTIFIED with slack message timestamp", async () => {
      mockPrisma.proposal.update.mockResolvedValue({ id: "uuid-1", state: "NOTIFIED" });

      await repository.markNotified("uuid-1", "slack-ts-123");

      expect(mockPrisma.proposal.update).toHaveBeenCalledWith({
        where: { id: "uuid-1" },
        data: expect.objectContaining({
          state: "NOTIFIED",
          slackMessageTs: "slack-ts-123",
          notifiedAt: expect.any(Date),
        }),
      });
    });
  });
});
