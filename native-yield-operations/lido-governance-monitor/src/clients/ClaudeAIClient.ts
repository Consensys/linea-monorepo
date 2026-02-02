import Anthropic from "@anthropic-ai/sdk";
import { ILogger } from "@consensys/linea-shared-utils";
import { z } from "zod";

import { IAIClient, AIAnalysisRequest } from "../core/clients/IAIClient.js";
import { Assessment } from "../core/entities/Assessment.js";

const AssessmentSchema = z.object({
  riskScore: z.number().int().min(0).max(100),
  riskLevel: z.enum(["low", "medium", "high", "critical"]),
  confidence: z.number().int().min(0).max(100),
  proposalType: z.enum(["discourse", "snapshot", "onchain_vote"]),
  impactTypes: z.array(z.enum(["economic", "technical", "operational", "governance-process"])),
  affectedComponents: z.array(
    z.enum(["StakingVault", "VaultHub", "LazyOracle", "OperatorGrid", "PredepositGuarantee", "Dashboard", "Other"]),
  ),
  whatChanged: z.string().min(1),
  nativeYieldInvariantsAtRisk: z.array(
    z.enum([
      "A_valid_yield_reporting",
      "B_user_principal_protection",
      "C_pause_deposits_on_deficit_or_liability_or_ossification",
      "Other",
    ]),
  ),
  whyItMattersForLineaNativeYield: z.string().min(1),
  recommendedAction: z.enum(["no-action", "monitor", "comment", "escalate"]),
  urgency: z.enum(["none", "this_week", "pre_execution", "immediate"]),
  supportingQuotes: z.array(z.string()),
  keyUnknowns: z.array(z.string()),
});

export class ClaudeAIClient implements IAIClient {
  constructor(
    private readonly logger: ILogger,
    private readonly anthropicClient: Anthropic,
    private readonly modelName: string,
    private readonly systemPromptTemplate: string,
  ) {}

  async analyzeProposal(request: AIAnalysisRequest): Promise<Assessment | undefined> {
    const userPrompt = this.buildUserPrompt(request);

    try {
      const response = await this.anthropicClient.messages.create({
        model: this.modelName,
        max_tokens: 2048,
        system: this.systemPromptTemplate,
        messages: [{ role: "user", content: userPrompt }],
      });

      const textContent = response.content.find((c) => c.type === "text");
      if (!textContent || textContent.type !== "text") {
        this.logger.error("AI response missing text content");
        return undefined;
      }

      const parsed = this.parseAndValidate(textContent.text);
      if (!parsed) return undefined;

      this.logger.debug("AI analysis completed", {
        proposalTitle: request.proposalTitle,
        riskScore: parsed.riskScore,
      });
      return parsed;
    } catch (error) {
      this.logger.error("AI analysis failed", { error });
      return undefined;
    }
  }

  getModelName(): string {
    return this.modelName;
  }

  private parseAndValidate(responseText: string): Assessment | undefined {
    try {
      const jsonMatch = responseText.match(/\{[\s\S]*\}/);
      if (!jsonMatch) {
        this.logger.error("No JSON found in AI response");
        return undefined;
      }
      const parsed = JSON.parse(jsonMatch[0]);
      const result = AssessmentSchema.safeParse(parsed);
      if (!result.success) {
        this.logger.error("AI response failed schema validation", { errors: result.error.errors });
        return undefined;
      }
      return result.data as Assessment;
    } catch (error) {
      this.logger.error("Failed to parse AI response as JSON", { error });
      return undefined;
    }
  }

  private buildUserPrompt(request: AIAnalysisRequest): string {
    const payloadSection = request.proposalPayload ? `\nPayload:\n${request.proposalPayload.substring(0, 5000)}` : "";

    return `Analyze this Lido governance proposal:

Title: ${request.proposalTitle}
URL: ${request.proposalUrl}
Type: ${request.proposalType}

Content:
${request.proposalText.substring(0, 10000)}${payloadSection}`;
  }
}
