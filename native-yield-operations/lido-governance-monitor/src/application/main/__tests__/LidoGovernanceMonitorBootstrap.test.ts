import { jest, describe, it, expect } from "@jest/globals";

// Mock PrismaPg adapter
jest.mock("@prisma/adapter-pg", () => ({
  PrismaPg: jest.fn().mockImplementation(() => ({})),
}));

// Mock PrismaClient
jest.mock("../../../../prisma/client/client.js", () => ({
  PrismaClient: jest.fn().mockImplementation(() => ({
    $connect: jest.fn().mockResolvedValue(undefined),
    $disconnect: jest.fn().mockResolvedValue(undefined),
  })),
}));

// Mock linea-shared-utils with WinstonLogger and ExponentialBackoffRetryService
jest.mock("@consensys/linea-shared-utils", () => ({
  WinstonLogger: jest.fn().mockImplementation(() => ({
    name: "test-logger",
    debug: jest.fn(),
    error: jest.fn(),
    info: jest.fn(),
    warn: jest.fn(),
  })),
  ExponentialBackoffRetryService: jest.fn().mockImplementation(() => ({
    retry: jest.fn().mockImplementation(<T>(fn: () => Promise<T>) => fn()),
  })),
}));

// Anthropic is mocked via moduleNameMapper in jest.config.js

import { Config } from "../config/index.js";
import { LidoGovernanceMonitorBootstrap } from "../LidoGovernanceMonitorBootstrap.js";

describe("LidoGovernanceMonitorBootstrap", () => {
  const createMockConfig = (): Config => ({
    database: { url: "postgresql://localhost:5432/test" },
    discourse: {
      proposalsUrl: "https://research.lido.fi/c/proposals/9/l/latest.json",
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
  });

  describe("create", () => {
    it("creates bootstrap instance with all dependencies wired", () => {
      // Arrange
      const config = createMockConfig();
      const systemPrompt = "You are a security analyst...";

      // Act
      const bootstrap = LidoGovernanceMonitorBootstrap.create(config, systemPrompt);

      // Assert
      expect(bootstrap).toBeDefined();
      expect(bootstrap.getProposalFetcher()).toBeDefined();
      expect(bootstrap.getProposalProcessor()).toBeDefined();
      expect(bootstrap.getNotificationService()).toBeDefined();
    });
  });

  describe("getters", () => {
    it("returns ProposalFetcher instance", () => {
      // Arrange
      const config = createMockConfig();
      const systemPrompt = "You are a security analyst...";
      const bootstrap = LidoGovernanceMonitorBootstrap.create(config, systemPrompt);

      // Act
      const poller = bootstrap.getProposalFetcher();

      // Assert
      expect(poller).toBeDefined();
      expect(typeof poller.pollOnce).toBe("function");
    });

    it("returns ProposalProcessor instance", () => {
      // Arrange
      const config = createMockConfig();
      const systemPrompt = "You are a security analyst...";
      const bootstrap = LidoGovernanceMonitorBootstrap.create(config, systemPrompt);

      // Act
      const processor = bootstrap.getProposalProcessor();

      // Assert
      expect(processor).toBeDefined();
      expect(typeof processor.processOnce).toBe("function");
    });

    it("returns NotificationService instance", () => {
      // Arrange
      const config = createMockConfig();
      const systemPrompt = "You are a security analyst...";
      const bootstrap = LidoGovernanceMonitorBootstrap.create(config, systemPrompt);

      // Act
      const notificationService = bootstrap.getNotificationService();

      // Assert
      expect(notificationService).toBeDefined();
      expect(typeof notificationService.notifyOnce).toBe("function");
    });
  });
});
