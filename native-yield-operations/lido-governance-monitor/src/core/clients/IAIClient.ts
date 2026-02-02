import { Assessment, ProposalType } from "../entities/Assessment.js";

export interface AIAnalysisRequest {
  proposalTitle: string;
  proposalText: string;
  proposalUrl: string;
  proposalType: ProposalType;
  proposalPayload?: string;
}

export interface IAIClient {
  analyzeProposal(request: AIAnalysisRequest): Promise<Assessment | undefined>;
  getModelName(): string;
}
