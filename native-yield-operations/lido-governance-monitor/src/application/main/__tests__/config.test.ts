import { describe, it, expect } from "@jest/globals";

import { ConfigSchema, loadConfigFromEnv } from "../config/index.js";

describe("ConfigSchema", () => {
  describe("validation", () => {
    it("validates a complete valid config", () => {
      // Arrange
      const validConfig = {
        database: { url: "postgresql://localhost:5432/test" },
        discourse: {
          proposalsUrl: "https://research.lido.fi/c/proposals/9/l/latest.json",
          maxTopicsPerPoll: 20,
        },
        anthropic: {
          apiKey: "sk-ant-xxx",
          model: "claude-sonnet-4-20250514",
          maxOutputTokens: 2048,
          maxProposalChars: 50000,
        },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: {
          threshold: 60,
          promptVersion: "v1.0",
        },
        http: {
          timeoutMs: 15000,
        },
      };

      // Act
      const result = ConfigSchema.safeParse(validConfig);

      // Assert
      expect(result.success).toBe(true);
    });

    it("rejects config with missing required fields", () => {
      // Arrange
      const invalidConfig = {
        database: { url: "postgresql://localhost:5432/test" },
        // Missing other required fields
      };

      // Act
      const result = ConfigSchema.safeParse(invalidConfig);

      // Assert
      expect(result.success).toBe(false);
    });

    it("rejects invalid database URL", () => {
      // Arrange
      const invalidConfig = {
        database: { url: "" },
        discourse: {
          proposalsUrl: "https://research.lido.fi/c/proposals/9/l/latest.json",
        },
        anthropic: { apiKey: "sk-ant-xxx", model: "claude-sonnet-4" },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: {
          threshold: 60,
          promptVersion: "v1.0",
        },
      };

      // Act
      const result = ConfigSchema.safeParse(invalidConfig);

      // Assert
      expect(result.success).toBe(false);
    });

    it("rejects risk threshold outside valid range", () => {
      // Arrange
      const invalidConfig = {
        database: { url: "postgresql://localhost:5432/test" },
        discourse: {
          proposalsUrl: "https://research.lido.fi/c/proposals/9/l/latest.json",
        },
        anthropic: { apiKey: "sk-ant-xxx", model: "claude-sonnet-4" },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: {
          threshold: 150, // Invalid: > 100
          promptVersion: "v1.0",
        },
      };

      // Act
      const result = ConfigSchema.safeParse(invalidConfig);

      // Assert
      expect(result.success).toBe(false);
    });

    describe("whitespace validation", () => {
      const validBase = {
        database: { url: "postgresql://localhost:5432/test" },
        discourse: {
          proposalsUrl: "https://research.lido.fi/c/proposals/9/l/latest.json",
          maxTopicsPerPoll: 20,
        },
        anthropic: { apiKey: "sk-ant-xxx", model: "claude-sonnet-4", maxOutputTokens: 2048, maxProposalChars: 50000 },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: { threshold: 60, promptVersion: "v1.0" },
        http: { timeoutMs: 15000 },
      };

      it("rejects database URL with only spaces", () => {
        // Arrange
        const config = { ...validBase, database: { url: "   " } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
        if (!result.success) {
          expect(result.error.errors[0].message).toContain("Database URL is required");
        }
      });

      it("rejects database URL with only tabs", () => {
        // Arrange
        const config = { ...validBase, database: { url: "\t\t" } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
      });

      it("rejects database URL with only newlines", () => {
        // Arrange
        const config = { ...validBase, database: { url: "\n\n" } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
      });

      it("rejects database URL with mixed whitespace", () => {
        // Arrange
        const config = { ...validBase, database: { url: " \t\n " } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
      });

      it("accepts and trims database URL with leading/trailing whitespace", () => {
        // Arrange
        const config = { ...validBase, database: { url: "  postgresql://localhost:5432/test  " } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(true);
        if (result.success) {
          expect(result.data.database.url).toBe("postgresql://localhost:5432/test");
        }
      });

      it("rejects anthropic API key with only spaces", () => {
        // Arrange
        const config = { ...validBase, anthropic: { ...validBase.anthropic, apiKey: "   " } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
        if (!result.success) {
          expect(result.error.errors[0].message).toContain("Anthropic API key is required");
        }
      });

      it("rejects discourse URL with only whitespace", () => {
        // Arrange
        const config = { ...validBase, discourse: { ...validBase.discourse, proposalsUrl: "   " } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
      });

      it("rejects slack webhook URL with only whitespace", () => {
        // Arrange
        const config = { ...validBase, slack: { webhookUrl: "   " } };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
      });

      it("rejects prompt version with only whitespace", () => {
        // Arrange
        const config = {
          ...validBase,
          riskAssessment: { ...validBase.riskAssessment, promptVersion: "   " },
        };

        // Act
        const result = ConfigSchema.safeParse(config);

        // Assert
        expect(result.success).toBe(false);
      });
    });
  });
});

describe("loadConfigFromEnv", () => {
  it("loads config from environment variables", () => {
    // Arrange
    const env = {
      DATABASE_URL: "postgresql://localhost:5432/test",
      DISCOURSE_PROPOSALS_URL: "https://research.lido.fi/c/proposals/9/l/latest.json",
      ANTHROPIC_API_KEY: "sk-ant-xxx",
      CLAUDE_MODEL: "claude-sonnet-4-20250514",
      SLACK_WEBHOOK_URL: "https://hooks.slack.com/services/xxx",
      RISK_THRESHOLD: "60",
      PROMPT_VERSION: "v1.0",
    };

    // Act
    const config = loadConfigFromEnv(env);

    // Assert
    expect(config.database.url).toBe("postgresql://localhost:5432/test");
    expect(config.discourse.proposalsUrl).toBe("https://research.lido.fi/c/proposals/9/l/latest.json");
    expect(config.anthropic.apiKey).toBe("sk-ant-xxx");
    expect(config.riskAssessment.threshold).toBe(60);
  });

  it("uses default values when optional env vars are missing", () => {
    // Arrange
    const env = {
      DATABASE_URL: "postgresql://localhost:5432/test",
      DISCOURSE_PROPOSALS_URL: "https://research.lido.fi/c/proposals/9/l/latest.json",
      ANTHROPIC_API_KEY: "sk-ant-xxx",
      SLACK_WEBHOOK_URL: "https://hooks.slack.com/services/xxx",
    };

    // Act
    const config = loadConfigFromEnv(env);

    // Assert
    expect(config.riskAssessment.threshold).toBe(60); // Default
    expect(config.discourse.maxTopicsPerPoll).toBe(20); // Default
  });

  it("throws on missing required env vars", () => {
    // Arrange
    const env = {
      DATABASE_URL: "postgresql://localhost:5432/test",
      // Missing ANTHROPIC_API_KEY and other required vars
    };

    // Act & Assert
    expect(() => loadConfigFromEnv(env)).toThrow();
  });

  it("uses empty string fallback when DATABASE_URL is undefined", () => {
    // Arrange - DATABASE_URL is undefined
    const env = {
      // DATABASE_URL is missing
      DISCOURSE_PROPOSALS_URL: "https://research.lido.fi/c/proposals/9/l/latest.json",
      ANTHROPIC_API_KEY: "sk-ant-xxx",
      SLACK_WEBHOOK_URL: "https://hooks.slack.com/services/xxx",
    };

    // Act & Assert - Should throw because empty string fails validation,
    // but this covers the ?? "" fallback branch on line 40
    expect(() => loadConfigFromEnv(env)).toThrow();
  });
});
