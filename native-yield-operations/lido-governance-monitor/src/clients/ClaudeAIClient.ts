import Anthropic from "@anthropic-ai/sdk";
import { z } from "zod";
import { ILogger } from "@consensys/linea-shared-utils";
import { IAIClient, AIAnalysisRequest } from "../core/clients/IAIClient.js";
import { Assessment } from "../core/entities/Assessment.js";

const AssessmentSchema = z.object({
  riskScore: z.number().int().min(0).max(100),
  impactType: z.enum(["economic", "technical", "operational", "governance-process"]),
  riskLevel: z.enum(["low", "medium", "high"]),
  whatChanged: z.string().min(1),
  whyItMattersForLineaNativeYield: z.string().min(1),
  recommendedAction: z.enum(["monitor", "comment", "escalate", "no-action"]),
  supportingQuotes: z.array(z.string()),
});

export class ClaudeAIClient implements IAIClient {
  constructor(
    private readonly logger: ILogger,
    private readonly anthropicClient: Anthropic,
    private readonly modelName: string,
    private readonly promptVersion: string
  ) {}

  async analyzeProposal(request: AIAnalysisRequest): Promise<Assessment | undefined> {
    const systemPrompt = this.buildSystemPrompt(request.domainContext);
    const userPrompt = this.buildUserPrompt(request);

    try {
      const response = await this.anthropicClient.messages.create({
        model: this.modelName,
        max_tokens: 2048,
        system: systemPrompt,
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

  private buildSystemPrompt(domainContext: string): string {
    return `You are a security analyst evaluating Lido governance proposals for their potential impact on Linea's Native Yield system.

${domainContext}

Respond with a valid JSON object matching this schema:
{
  "riskScore": <number 0-100>,
  "impactType": "economic" | "technical" | "operational" | "governance-process",
  "riskLevel": "low" | "medium" | "high",
  "whatChanged": "<brief description>",
  "whyItMattersForLineaNativeYield": "<explanation>",
  "recommendedAction": "monitor" | "comment" | "escalate" | "no-action",
  "supportingQuotes": ["<relevant quotes>"]
}

Be conservative - when in doubt, assign a higher risk score.`;
  }

  private buildUserPrompt(request: AIAnalysisRequest): string {
    return `Analyze this Lido governance proposal:

Title: ${request.proposalTitle}
URL: ${request.proposalUrl}

Content:
${request.proposalText.substring(0, 10000)}

Provide your risk assessment as JSON.`;
  }
}
