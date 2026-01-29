import { Proposal } from "../entities/Proposal.js";
import { Assessment } from "../entities/Assessment.js";

export interface INotificationService {
  notifyProposal(proposal: Proposal, assessment: Assessment): Promise<boolean>;
}
