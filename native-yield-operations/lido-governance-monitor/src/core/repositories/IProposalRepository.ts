import { Assessment } from "../entities/Assessment.js";
import { Proposal, CreateProposalInput } from "../entities/Proposal.js";
import { ProposalSource } from "../entities/ProposalSource.js";
import { ProposalState } from "../entities/ProposalState.js";

export interface IProposalRepository {
  findBySourceAndSourceId(source: ProposalSource, sourceId: string): Promise<Proposal | null>;
  findByState(state: ProposalState): Promise<Proposal[]>;
  create(input: CreateProposalInput): Promise<Proposal>;
  updateState(id: string, state: ProposalState): Promise<Proposal>;
  saveAnalysis(
    id: string,
    assessment: Assessment,
    riskScore: number,
    llmModel: string,
    riskThreshold: number,
    promptVersion: string,
  ): Promise<Proposal>;
  incrementAnalysisAttempt(id: string): Promise<Proposal>;
  incrementNotifyAttempt(id: string): Promise<Proposal>;
  markNotified(id: string, slackMessageTs: string): Promise<Proposal>;
  findLatestSourceIdBySource(source: ProposalSource): Promise<string | null>;
}
