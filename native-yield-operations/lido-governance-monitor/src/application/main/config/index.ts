import { z } from "zod";

export const ConfigSchema = z.object({
  database: z.object({
    url: z.string().min(1, "Database URL is required"),
  }),
  discourse: z.object({
    proposalsUrl: z.string().url("Invalid Discourse proposals URL"),
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
});

export type Config = z.infer<typeof ConfigSchema>;

export function loadConfigFromEnv(env: Record<string, string | undefined>): Config {
  const rawConfig = {
    database: {
      url: env.DATABASE_URL ?? "",
    },
    discourse: {
      proposalsUrl: env.DISCOURSE_PROPOSALS_URL ?? "",
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
  };

  const result = ConfigSchema.safeParse(rawConfig);
  if (!result.success) {
    const errors = result.error.errors.map((e) => `${e.path.join(".")}: ${e.message}`).join(", ");
    throw new Error(`Configuration validation failed: ${errors}`);
  }

  return result.data;
}
