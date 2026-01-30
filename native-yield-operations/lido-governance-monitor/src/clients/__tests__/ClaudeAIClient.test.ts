import { jest, describe, it, expect, beforeEach } from "@jest/globals";
import { ClaudeAIClient } from "../ClaudeAIClient.js";
import { ILogger } from "@consensys/linea-shared-utils";
import Anthropic from "@anthropic-ai/sdk";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

const TEST_SYSTEM_PROMPT = `You are a security analyst.

{{DOMAIN_CONTEXT}}

Respond with valid JSON.`;

describe("ClaudeAIClient", () => {
  let client: ClaudeAIClient;
  let logger: jest.Mocked<ILogger>;
  let mockAnthropicClient: { messages: { create: jest.Mock } };

  beforeEach(() => {
    logger = createLoggerMock();
    mockAnthropicClient = { messages: { create: jest.fn() } };
    client = new ClaudeAIClient(
      logger,
      mockAnthropicClient as unknown as Anthropic,
      "claude-sonnet-4-20250514",
      TEST_SYSTEM_PROMPT
    );
  });

  describe("analyzeProposal", () => {
    it("returns valid assessment when AI response is well-formed", async () => {
      // Arrange
      const validAssessment = {
        riskScore: 75,
        impactType: "technical",
        riskLevel: "high",
        whatChanged: "Contract upgrade to v2",
        whyItMattersForLineaNativeYield: "May affect withdrawal mechanics",
        recommendedAction: "escalate",
        supportingQuotes: ["The upgrade will modify..."],
      };
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: JSON.stringify(validAssessment) }],
      });

      // Act
      const result = await client.analyzeProposal({
        proposalTitle: "Test Proposal",
        proposalText: "Proposal content",
        proposalUrl: "https://example.com",
        domainContext: "Domain context",
      });

      // Assert
      expect(result).toEqual(validAssessment);
      expect(logger.debug).toHaveBeenCalledWith("AI analysis completed", expect.any(Object));
    });

    it("extracts JSON from response with surrounding text", async () => {
      // Arrange
      const validAssessment = {
        riskScore: 50,
        impactType: "economic",
        riskLevel: "medium",
        whatChanged: "Fee structure change",
        whyItMattersForLineaNativeYield: "May impact yields",
        recommendedAction: "monitor",
        supportingQuotes: ["quote"],
      };
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: `Here's my analysis:\n${JSON.stringify(validAssessment)}\nEnd of analysis.` }],
      });

      // Act
      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      // Assert
      expect(result).toEqual(validAssessment);
    });

    it("returns undefined when AI response fails schema validation", async () => {
      // Arrange
      const invalidAssessment = { riskScore: 150, impactType: "invalid-type" };
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: JSON.stringify(invalidAssessment) }],
      });

      // Act
      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("AI response failed schema validation", expect.any(Object));
    });

    it("returns undefined when response has no JSON", async () => {
      // Arrange
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: "I cannot analyze this proposal." }],
      });

      // Act
      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("No JSON found in AI response");
    });

    it("returns undefined on API failure", async () => {
      // Arrange
      mockAnthropicClient.messages.create.mockRejectedValue(new Error("API rate limit exceeded"));

      // Act
      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("AI analysis failed", expect.any(Object));
    });

    it("returns undefined when response has no text content", async () => {
      // Arrange
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [],
      });

      // Act
      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("AI response missing text content");
    });

    it("returns undefined when JSON is malformed", async () => {
      // Arrange
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: "{ invalid json }" }],
      });

      // Act
      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to parse AI response as JSON", expect.any(Object));
    });

    it("replaces domain context placeholder in system prompt", async () => {
      // Arrange
      const validAssessment = {
        riskScore: 30,
        impactType: "operational",
        riskLevel: "low",
        whatChanged: "Minor update",
        whyItMattersForLineaNativeYield: "No direct impact",
        recommendedAction: "no-action",
        supportingQuotes: [],
      };
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: JSON.stringify(validAssessment) }],
      });

      // Act
      await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Custom domain context here",
      });

      // Assert
      expect(mockAnthropicClient.messages.create).toHaveBeenCalledWith(
        expect.objectContaining({
          system: expect.stringContaining("Custom domain context here"),
        })
      );
    });
  });

  describe("getModelName", () => {
    it("returns the configured model name", () => {
      // Act & Assert
      expect(client.getModelName()).toBe("claude-sonnet-4-20250514");
    });
  });
});
