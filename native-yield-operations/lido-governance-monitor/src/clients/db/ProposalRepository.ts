import { PrismaClient } from "@prisma/client";

import { Assessment } from "../../core/entities/Assessment.js";
import { Proposal, CreateProposalInput } from "../../core/entities/Proposal.js";
import { ProposalSource } from "../../core/entities/ProposalSource.js";
import { ProposalState } from "../../core/entities/ProposalState.js";
import { IProposalRepository } from "../../core/repositories/IProposalRepository.js";

export class ProposalRepository implements IProposalRepository {
  constructor(private readonly prisma: PrismaClient) {}

  async findBySourceAndSourceId(source: ProposalSource, sourceId: string): Promise<Proposal | null> {
    return this.prisma.proposal.findUnique({
      where: { source_sourceId: { source, sourceId } },
    }) as Promise<Proposal | null>;
  }

  async findByState(state: ProposalState): Promise<Proposal[]> {
    return this.prisma.proposal.findMany({
      where: { state },
      orderBy: { stateUpdatedAt: "asc" },
    }) as Promise<Proposal[]>;
  }

  async create(input: CreateProposalInput): Promise<Proposal> {
    return this.prisma.proposal.create({
      data: {
        source: input.source,
        sourceId: input.sourceId,
        url: input.url,
        title: input.title,
        author: input.author,
        sourceCreatedAt: input.sourceCreatedAt,
        text: input.text,
        state: ProposalState.NEW,
        stateUpdatedAt: new Date(),
      },
    }) as Promise<Proposal>;
  }

  async updateState(id: string, state: ProposalState): Promise<Proposal> {
    return this.prisma.proposal.update({
      where: { id },
      data: { state, stateUpdatedAt: new Date() },
    }) as Promise<Proposal>;
  }

  async saveAnalysis(
    id: string,
    assessment: Assessment,
    riskScore: number,
    llmModel: string,
    riskThreshold: number,
    promptVersion: string,
  ): Promise<Proposal> {
    return this.prisma.proposal.update({
      where: { id },
      data: {
        state: ProposalState.ANALYZED,
        stateUpdatedAt: new Date(),
        assessmentJson: assessment as object,
        riskScore,
        llmModel,
        riskThreshold,
        assessmentPromptVersion: promptVersion,
        analyzedAt: new Date(),
      },
    }) as Promise<Proposal>;
  }

  async incrementAnalysisAttempt(id: string): Promise<Proposal> {
    return this.prisma.proposal.update({
      where: { id },
      data: { analysisAttemptCount: { increment: 1 } },
    }) as Promise<Proposal>;
  }

  async incrementNotifyAttempt(id: string): Promise<Proposal> {
    return this.prisma.proposal.update({
      where: { id },
      data: { notifyAttemptCount: { increment: 1 } },
    }) as Promise<Proposal>;
  }

  async markNotified(id: string, slackMessageTs: string): Promise<Proposal> {
    return this.prisma.proposal.update({
      where: { id },
      data: {
        state: ProposalState.NOTIFIED,
        stateUpdatedAt: new Date(),
        notifiedAt: new Date(),
        slackMessageTs,
      },
    }) as Promise<Proposal>;
  }
}
