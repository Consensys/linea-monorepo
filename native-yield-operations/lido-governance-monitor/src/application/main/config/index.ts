import { z } from "zod";

export const ConfigSchema = z.object({
  database: z.object({
    url: z.string().min(1, "Database URL is required"),
  }),
  discourse: z.object({
    baseUrl: z.string().url("Invalid Discourse base URL"),
    proposalsCategoryId: z.number().int().positive("Category ID must be positive"),
    pollingIntervalMs: z.number().int().positive("Polling interval must be positive"),
  }),
  anthropic: z.object({
    apiKey: z.string().min(1, "Anthropic API key is required"),
    model: z.string().min(1, "Model name is required"),
  }),
  slack: z.object({
    webhookUrl: z.string().url("Invalid Slack webhook URL"),
  }),
  riskAssessment: z.object({
    threshold: z.number().int().min(0).max(100, "Threshold must be 0-100"),
    promptVersion: z.string().min(1, "Prompt version is required"),
    domainContext: z.string().min(1, "Domain context is required"),
  }),
  processing: z.object({
    intervalMs: z.number().int().positive("Processing interval must be positive"),
  }),
});

export type Config = z.infer<typeof ConfigSchema>;

export function loadConfigFromEnv(env: Record<string, string | undefined>): Config {
  const rawConfig = {
    database: {
      url: env.DATABASE_URL ?? "",
    },
    discourse: {
      baseUrl: env.DISCOURSE_BASE_URL ?? "",
      proposalsCategoryId: parseInt(env.DISCOURSE_PROPOSALS_CATEGORY_ID ?? "9", 10),
      pollingIntervalMs: parseInt(env.DISCOURSE_POLLING_INTERVAL_MS ?? "3600000", 10),
    },
    anthropic: {
      apiKey: env.ANTHROPIC_API_KEY ?? "",
      model: env.CLAUDE_MODEL ?? "claude-sonnet-4-20250514",
    },
    slack: {
      webhookUrl: env.SLACK_WEBHOOK_URL ?? "",
    },
    riskAssessment: {
      threshold: parseInt(env.RISK_THRESHOLD ?? "60", 10),
      promptVersion: env.PROMPT_VERSION ?? "v1.0",
      domainContext: env.DOMAIN_CONTEXT ?? "",
    },
    processing: {
      intervalMs: parseInt(env.PROCESSING_INTERVAL_MS ?? "60000", 10),
    },
  };

  const result = ConfigSchema.safeParse(rawConfig);
  if (!result.success) {
    const errors = result.error.errors.map((e) => `${e.path.join(".")}: ${e.message}`).join(", ");
    throw new Error(`Configuration validation failed: ${errors}`);
  }

  return result.data;
}
