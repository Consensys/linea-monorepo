import { describe, it, expect } from "@jest/globals";
import { ConfigSchema, loadConfigFromEnv } from "../config/index.js";

describe("ConfigSchema", () => {
  describe("validation", () => {
    it("validates a complete valid config", () => {
      // Arrange
      const validConfig = {
        database: { url: "postgresql://localhost:5432/test" },
        discourse: {
          baseUrl: "https://research.lido.fi",
          proposalsCategoryId: 9,
          pollingIntervalMs: 3600000,
        },
        anthropic: {
          apiKey: "sk-ant-xxx",
          model: "claude-sonnet-4-20250514",
        },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: {
          threshold: 60,
          promptVersion: "v1.0",
          domainContext: "Context",
          maxAnalysisAttempts: 3,
        },
        processing: {
          intervalMs: 60000,
          maxNotifyAttempts: 3,
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
          baseUrl: "https://research.lido.fi",
          proposalsCategoryId: 9,
          pollingIntervalMs: 3600000,
        },
        anthropic: { apiKey: "sk-ant-xxx", model: "claude-sonnet-4" },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: {
          threshold: 60,
          promptVersion: "v1.0",
          domainContext: "Context",
          maxAnalysisAttempts: 3,
        },
        processing: { intervalMs: 60000, maxNotifyAttempts: 3 },
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
          baseUrl: "https://research.lido.fi",
          proposalsCategoryId: 9,
          pollingIntervalMs: 3600000,
        },
        anthropic: { apiKey: "sk-ant-xxx", model: "claude-sonnet-4" },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: {
          threshold: 150, // Invalid: > 100
          promptVersion: "v1.0",
          domainContext: "Context",
          maxAnalysisAttempts: 3,
        },
        processing: { intervalMs: 60000, maxNotifyAttempts: 3 },
      };

      // Act
      const result = ConfigSchema.safeParse(invalidConfig);

      // Assert
      expect(result.success).toBe(false);
    });

    it("rejects negative polling interval", () => {
      // Arrange
      const invalidConfig = {
        database: { url: "postgresql://localhost:5432/test" },
        discourse: {
          baseUrl: "https://research.lido.fi",
          proposalsCategoryId: 9,
          pollingIntervalMs: -1000, // Invalid: negative
        },
        anthropic: { apiKey: "sk-ant-xxx", model: "claude-sonnet-4" },
        slack: { webhookUrl: "https://hooks.slack.com/services/xxx" },
        riskAssessment: {
          threshold: 60,
          promptVersion: "v1.0",
          domainContext: "Context",
          maxAnalysisAttempts: 3,
        },
        processing: { intervalMs: 60000, maxNotifyAttempts: 3 },
      };

      // Act
      const result = ConfigSchema.safeParse(invalidConfig);

      // Assert
      expect(result.success).toBe(false);
    });
  });
});

describe("loadConfigFromEnv", () => {
  it("loads config from environment variables", () => {
    // Arrange
    const env = {
      DATABASE_URL: "postgresql://localhost:5432/test",
      DISCOURSE_BASE_URL: "https://research.lido.fi",
      DISCOURSE_PROPOSALS_CATEGORY_ID: "9",
      DISCOURSE_POLLING_INTERVAL_MS: "3600000",
      ANTHROPIC_API_KEY: "sk-ant-xxx",
      CLAUDE_MODEL: "claude-sonnet-4-20250514",
      SLACK_WEBHOOK_URL: "https://hooks.slack.com/services/xxx",
      RISK_THRESHOLD: "60",
      PROMPT_VERSION: "v1.0",
      DOMAIN_CONTEXT: "Domain context here",
      MAX_ANALYSIS_ATTEMPTS: "3",
      PROCESSING_INTERVAL_MS: "60000",
      MAX_NOTIFY_ATTEMPTS: "3",
    };

    // Act
    const config = loadConfigFromEnv(env);

    // Assert
    expect(config.database.url).toBe("postgresql://localhost:5432/test");
    expect(config.discourse.baseUrl).toBe("https://research.lido.fi");
    expect(config.discourse.proposalsCategoryId).toBe(9);
    expect(config.anthropic.apiKey).toBe("sk-ant-xxx");
    expect(config.riskAssessment.threshold).toBe(60);
  });

  it("uses default values when optional env vars are missing", () => {
    // Arrange
    const env = {
      DATABASE_URL: "postgresql://localhost:5432/test",
      DISCOURSE_BASE_URL: "https://research.lido.fi",
      ANTHROPIC_API_KEY: "sk-ant-xxx",
      SLACK_WEBHOOK_URL: "https://hooks.slack.com/services/xxx",
      DOMAIN_CONTEXT: "Context",
    };

    // Act
    const config = loadConfigFromEnv(env);

    // Assert
    expect(config.discourse.proposalsCategoryId).toBe(9); // Default
    expect(config.riskAssessment.threshold).toBe(60); // Default
    expect(config.processing.intervalMs).toBe(60000); // Default
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
});
