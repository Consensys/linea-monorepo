import { ILogger } from "@consensys/linea-shared-utils";

import { ISlackClient, SlackNotificationResult } from "../core/clients/ISlackClient.js";
import { Assessment, RiskLevel } from "../core/entities/Assessment.js";
import { Proposal } from "../core/entities/Proposal.js";

export class SlackClient implements ISlackClient {
  constructor(
    private readonly logger: ILogger,
    private readonly webhookUrl: string,
  ) {}

  async sendProposalAlert(proposal: Proposal, assessment: Assessment): Promise<SlackNotificationResult> {
    const payload = this.buildSlackPayload(proposal, assessment);

    try {
      const response = await fetch(this.webhookUrl, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorText = await response.text();
        this.logger.error("Slack webhook failed", { status: response.status, error: errorText });
        return { success: false, error: errorText };
      }

      this.logger.info("Slack notification sent", { proposalId: proposal.id, title: proposal.title });
      return { success: true, messageTs: Date.now().toString() };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error";
      this.logger.error("Slack notification error", { error: errorMessage });
      return { success: false, error: errorMessage };
    }
  }

  private getRiskEmoji(riskLevel: RiskLevel): string {
    switch (riskLevel) {
      case "critical":
        return ":rotating_light:";
      case "high":
        return ":warning:";
      case "medium":
      case "low":
      default:
        return ":information_source:";
    }
  }

  private buildSlackPayload(proposal: Proposal, assessment: Assessment): object {
    const riskEmoji = this.getRiskEmoji(assessment.riskLevel);

    return {
      blocks: [
        {
          type: "header",
          text: {
            type: "plain_text",
            text: `${riskEmoji} Lido Governance Alert: ${proposal.title}`,
            emoji: true,
          },
        },
        {
          type: "section",
          fields: [
            { type: "mrkdwn", text: `*Risk Score:* ${assessment.riskScore}/100` },
            { type: "mrkdwn", text: `*Risk Level:* ${assessment.riskLevel.toUpperCase()}` },
            { type: "mrkdwn", text: `*Confidence:* ${assessment.confidence}%` },
            { type: "mrkdwn", text: `*Urgency:* ${assessment.urgency.replace("_", " ")}` },
          ],
        },
        {
          type: "section",
          fields: [
            { type: "mrkdwn", text: `*Impact Types:* ${assessment.impactTypes.join(", ")}` },
            { type: "mrkdwn", text: `*Action:* ${assessment.recommendedAction}` },
          ],
        },
        {
          type: "section",
          text: {
            type: "mrkdwn",
            text: `*Affected Components:*\n${assessment.affectedComponents.join(", ")}`,
          },
        },
        {
          type: "section",
          text: {
            type: "mrkdwn",
            text: `*Invariants at Risk:*\n${assessment.nativeYieldInvariantsAtRisk.map((i) => `• ${i}`).join("\n")}`,
          },
        },
        {
          type: "section",
          text: { type: "mrkdwn", text: `*What Changed:*\n${assessment.whatChanged}` },
        },
        {
          type: "section",
          text: {
            type: "mrkdwn",
            text: `*Why It Matters for Native Yield:*\n${assessment.whyItMattersForLineaNativeYield}`,
          },
        },
        {
          type: "actions",
          elements: [
            {
              type: "button",
              text: { type: "plain_text", text: "View Proposal", emoji: true },
              url: proposal.url,
              action_id: "view_proposal",
            },
          ],
        },
      ],
    };
  }
}
