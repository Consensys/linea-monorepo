import { ILogger } from "@consensys/linea-shared-utils";

import { ISlackClient } from "../core/clients/ISlackClient.js";
import { Assessment } from "../core/entities/Assessment.js";
import { Proposal } from "../core/entities/Proposal.js";
import { ProposalState } from "../core/entities/ProposalState.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { INotificationService } from "../core/services/INotificationService.js";

export class NotificationService implements INotificationService {
  constructor(
    private readonly logger: ILogger,
    private readonly slackClient: ISlackClient,
    private readonly proposalRepository: IProposalRepository,
    private readonly riskThreshold: number,
  ) {}

  async notifyOnce(): Promise<void> {
    try {
      this.logger.info("Starting notification processing");

      const analyzedProposals = await this.proposalRepository.findByState(ProposalState.ANALYZED);
      const failedProposals = await this.proposalRepository.findByState(ProposalState.NOTIFY_FAILED);
      const proposals = [...analyzedProposals, ...failedProposals];

      if (proposals.length === 0) {
        this.logger.debug("No proposals to notify");
        return;
      }

      this.logger.debug("Processing proposals for notification", { count: proposals.length });

      for (const proposal of proposals) {
        await this.notifyProposalInternal(proposal);
      }

      this.logger.info("Notification processing completed");
    } catch (error) {
      this.logger.error("Notification processing failed", error);
    }
  }

  private async notifyProposalInternal(proposal: Proposal): Promise<void> {
    try {
      // Validate assessment exists
      if (!proposal.assessmentJson) {
        this.logger.error("Proposal missing assessment data", { proposalId: proposal.id });
        return;
      }

      const assessment = proposal.assessmentJson as Assessment;

      // Send to audit channel unconditionally
      const auditResult = await this.slackClient.sendAuditLog(proposal, assessment);
      if (!auditResult.success) {
        // Log but don't fail - audit is best-effort
        this.logger.warn("Audit log failed, continuing", {
          proposalId: proposal.id,
          error: auditResult.error,
        });
      }

      // Check risk threshold BEFORE attempting notification
      if (proposal.riskScore === null || proposal.riskScore < this.riskThreshold) {
        // Below threshold - mark as NOT_NOTIFIED without sending notification
        await this.proposalRepository.updateState(proposal.id, ProposalState.NOT_NOTIFIED);
        this.logger.info("Proposal below notification threshold, skipped", {
          proposalId: proposal.id,
          riskScore: proposal.riskScore,
          threshold: this.riskThreshold,
        });
        return;
      }

      // Increment attempt count first
      const updated = await this.proposalRepository.incrementNotifyAttempt(proposal.id);

      // Send Slack notification
      const result = await this.slackClient.sendProposalAlert(proposal, assessment);

      if (result.success) {
        await this.proposalRepository.markNotified(proposal.id, result.messageTs ?? "");
        this.logger.info("Proposal notification sent", {
          proposalId: proposal.id,
          messageTs: result.messageTs,
        });
      } else {
        // Notification failed - will retry on next cycle
        this.logger.warn("Slack notification failed, will retry", {
          proposalId: proposal.id,
          attempt: updated.notifyAttemptCount,
          error: result.error,
        });
      }
    } catch (error) {
      this.logger.error("Error notifying proposal", { proposalId: proposal.id, error });
    }
  }
}
