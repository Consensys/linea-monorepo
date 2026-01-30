import { ILogger } from "@consensys/linea-shared-utils";

import { ISlackClient } from "../core/clients/ISlackClient.js";
import { Assessment } from "../core/entities/Assessment.js";
import { Proposal } from "../core/entities/Proposal.js";
import { ProposalState } from "../core/entities/ProposalState.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { INotificationService } from "../core/services/INotificationService.js";

export class NotificationService implements INotificationService {
  private intervalId: NodeJS.Timeout | null = null;

  constructor(
    private readonly logger: ILogger,
    private readonly slackClient: ISlackClient,
    private readonly proposalRepository: IProposalRepository,
    private readonly processingIntervalMs: number,
  ) {}

  start(): void {
    this.logger.info("NotificationService started", { intervalMs: this.processingIntervalMs });

    // Initial process
    void this.processOnce();

    // Schedule subsequent processing
    this.intervalId = setInterval(() => {
      void this.processOnce();
    }, this.processingIntervalMs);
  }

  stop(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
    this.logger.info("NotificationService stopped");
  }

  async processOnce(): Promise<void> {
    const pendingProposals = await this.proposalRepository.findByState(ProposalState.PENDING_NOTIFY);
    const failedProposals = await this.proposalRepository.findByState(ProposalState.NOTIFY_FAILED);
    const proposals = [...pendingProposals, ...failedProposals];

    if (proposals.length === 0) {
      this.logger.debug("No proposals to notify");
      return;
    }

    this.logger.debug("Processing proposals for notification", { count: proposals.length });

    for (const proposal of proposals) {
      await this.notifyProposalInternal(proposal);
    }
  }

  async notifyProposal(proposal: Proposal, assessment: Assessment): Promise<boolean> {
    const result = await this.slackClient.sendProposalAlert(proposal, assessment);
    return result.success;
  }

  private async notifyProposalInternal(proposal: Proposal): Promise<void> {
    try {
      // Validate assessment exists
      if (!proposal.assessmentJson) {
        this.logger.error("Proposal missing assessment data", { proposalId: proposal.id });
        return;
      }

      // Increment attempt count first
      const updated = await this.proposalRepository.incrementNotifyAttempt(proposal.id);

      // Send Slack notification
      const result = await this.slackClient.sendProposalAlert(proposal, proposal.assessmentJson as Assessment);

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
