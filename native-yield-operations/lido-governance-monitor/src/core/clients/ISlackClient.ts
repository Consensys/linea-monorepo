import { Assessment } from "../entities/Assessment.js";
import { Proposal } from "../entities/Proposal.js";

export interface SlackNotificationResult {
  success: boolean;
  messageTs?: string;
  error?: string;
}

export interface ISlackClient {
  sendProposalAlert(proposal: Proposal, assessment: Assessment): Promise<SlackNotificationResult>;
}
