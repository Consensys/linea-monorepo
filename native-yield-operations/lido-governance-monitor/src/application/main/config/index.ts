import "dotenv/config";

import { z } from "zod";

// Validation helpers that reject empty/whitespace-only strings
const NonEmptyString = (message: string) => z.string().trim().min(1, message);
const NonEmptyUrl = (message: string) => z.string().trim().url(message);

export const ConfigSchema = z.object({
  database: z.object({
    url: NonEmptyString("Database URL is required"),
  }),
  discourse: z.object({
    proposalsUrl: NonEmptyUrl("Invalid Discourse proposals URL"),
    maxTopicsPerPoll: z.number().int().positive("Max topics per poll must be positive"),
  }),
  anthropic: z.object({
    apiKey: NonEmptyString("Anthropic API key is required"),
    model: NonEmptyString("Model name is required"),
    maxOutputTokens: z.number().int().positive("Max output tokens must be positive"),
    maxProposalChars: z.number().int().positive("Max proposal chars must be positive"),
  }),
  slack: z.object({
    webhookUrl: NonEmptyUrl("Invalid Slack webhook URL"),
    auditWebhookUrl: NonEmptyUrl("Invalid Slack audit webhook URL").optional(),
  }),
  riskAssessment: z.object({
    threshold: z.number().int().min(0).max(100, "Threshold must be 0-100"),
    promptVersion: NonEmptyString("Prompt version is required"),
  }),
  ethereum: z.object({
    rpcUrl: NonEmptyUrl("Ethereum RPC URL is required"),
    ldoVotingContractAddress: NonEmptyString("LDO voting contract address is required"),
    maxVotesPerPoll: z.number().int().positive("Max votes per poll must be positive"),
  }),
  http: z.object({
    timeoutMs: z.number().int().positive("HTTP timeout must be positive"),
  }),
});

export type Config = z.infer<typeof ConfigSchema>;

export function loadConfigFromEnv(env: Record<string, string | undefined>): Config {
  const rawConfig = {
    database: {
      url: env.DATABASE_URL ?? "",
    },
    discourse: {
      proposalsUrl: env.DISCOURSE_PROPOSALS_URL ?? "",
      maxTopicsPerPoll: parseInt(env.MAX_TOPICS_PER_POLL ?? "20", 10),
    },
    anthropic: {
      apiKey: env.ANTHROPIC_API_KEY ?? "",
      model: env.CLAUDE_MODEL ?? "claude-sonnet-4-20250514",
      maxOutputTokens: parseInt(env.ANTHROPIC_MAX_OUTPUT_TOKENS ?? "2048", 10),
      maxProposalChars: parseInt(env.ANTHROPIC_MAX_PROPOSAL_CHARS ?? "700000", 10),
    },
    slack: {
      webhookUrl: env.SLACK_WEBHOOK_URL ?? "",
      auditWebhookUrl: env.SLACK_AUDIT_WEBHOOK_URL,
    },
    riskAssessment: {
      threshold: parseInt(env.RISK_THRESHOLD ?? "60", 10),
      promptVersion: env.PROMPT_VERSION ?? "v1.0",
    },
    ethereum: {
      rpcUrl: env.ETHEREUM_RPC_URL ?? "",
      ldoVotingContractAddress: env.LDO_VOTING_CONTRACT_ADDRESS ?? "",
      maxVotesPerPoll: parseInt(env.MAX_VOTES_PER_POLL ?? "20", 10),
    },
    http: {
      timeoutMs: parseInt(env.HTTP_TIMEOUT_MS ?? "15000", 10),
    },
  };

  const result = ConfigSchema.safeParse(rawConfig);
  if (!result.success) {
    const errors = result.error.errors.map((e) => `${e.path.join(".")}: ${e.message}`).join(", ");
    throw new Error(`Configuration validation failed: ${errors}`);
  }

  return result.data;
}
