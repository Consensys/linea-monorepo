import { Assessment } from "../entities/Assessment.js";

export interface AIAnalysisRequest {
  proposalTitle: string;
  proposalText: string;
  proposalUrl: string;
  domainContext: string;
}

export interface IAIClient {
  analyzeProposal(request: AIAnalysisRequest): Promise<Assessment | undefined>;
  getModelName(): string;
}
