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
      "v1.0"
    );
  });

  describe("analyzeProposal", () => {
    it("returns valid assessment when AI response is well-formed", async () => {
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

      const result = await client.analyzeProposal({
        proposalTitle: "Test Proposal",
        proposalText: "Proposal content",
        proposalUrl: "https://example.com",
        domainContext: "Domain context",
      });

      expect(result).toEqual(validAssessment);
      expect(logger.debug).toHaveBeenCalledWith("AI analysis completed", expect.any(Object));
    });

    it("extracts JSON from response with surrounding text", async () => {
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

      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      expect(result).toEqual(validAssessment);
    });

    it("returns undefined when AI response fails schema validation", async () => {
      const invalidAssessment = { riskScore: 150, impactType: "invalid-type" };
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: JSON.stringify(invalidAssessment) }],
      });

      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("AI response failed schema validation", expect.any(Object));
    });

    it("returns undefined when response has no JSON", async () => {
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: "I cannot analyze this proposal." }],
      });

      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("No JSON found in AI response");
    });

    it("returns undefined on API failure", async () => {
      mockAnthropicClient.messages.create.mockRejectedValue(new Error("API rate limit exceeded"));

      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("AI analysis failed", expect.any(Object));
    });

    it("returns undefined when response has no text content", async () => {
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [],
      });

      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("AI response missing text content");
    });

    it("returns undefined when JSON is malformed", async () => {
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: "{ invalid json }" }],
      });

      const result = await client.analyzeProposal({
        proposalTitle: "Test",
        proposalText: "Content",
        proposalUrl: "https://example.com",
        domainContext: "Context",
      });

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to parse AI response as JSON", expect.any(Object));
    });
  });

  describe("getModelName", () => {
    it("returns the configured model name", () => {
      expect(client.getModelName()).toBe("claude-sonnet-4-20250514");
    });
  });
});
