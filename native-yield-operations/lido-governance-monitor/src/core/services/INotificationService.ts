import { Assessment } from "../entities/Assessment.js";
import { Proposal } from "../entities/Proposal.js";

export interface INotificationService {
  notifyProposal(proposal: Proposal, assessment: Assessment): Promise<boolean>;
}
