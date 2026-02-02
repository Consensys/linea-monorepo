import Anthropic from "@anthropic-ai/sdk";
import { ILogger } from "@consensys/linea-shared-utils";
import { jest, describe, it, expect, beforeEach } from "@jest/globals";

import { ClaudeAIClient } from "../ClaudeAIClient.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
});

const TEST_SYSTEM_PROMPT = `You are a security analyst.

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
      TEST_SYSTEM_PROMPT,
    );
  });

  describe("analyzeProposal", () => {
    const createValidAssessment = (overrides = {}) => ({
      riskScore: 75,
      riskLevel: "high" as const,
      confidence: 85,
      proposalType: "discourse" as const,
      impactTypes: ["technical"] as const,
      affectedComponents: ["StakingVault"] as const,
      whatChanged: "Contract upgrade to v2",
      nativeYieldInvariantsAtRisk: ["A_valid_yield_reporting"] as const,
      whyItMattersForLineaNativeYield: "May affect withdrawal mechanics",
      recommendedAction: "escalate" as const,
      urgency: "pre_execution" as const,
      supportingQuotes: ["The upgrade will modify..."],
      keyUnknowns: [],
      ...overrides,
    });

    const createAnalysisRequest = (overrides = {}) => ({
      proposalTitle: "Test Proposal",
      proposalText: "Proposal content",
      proposalUrl: "https://example.com",
      proposalType: "discourse" as const,
      ...overrides,
    });

    it("returns valid assessment when AI response is well-formed", async () => {
      // Arrange
      const validAssessment = createValidAssessment();
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: JSON.stringify(validAssessment) }],
      });

      // Act
      const result = await client.analyzeProposal(createAnalysisRequest());

      // Assert
      expect(result).toEqual(validAssessment);
      expect(logger.debug).toHaveBeenCalledWith("AI analysis completed", expect.any(Object));
    });

    it("extracts JSON from response with surrounding text", async () => {
      // Arrange
      const validAssessment = createValidAssessment({
        riskScore: 50,
        riskLevel: "medium",
        impactTypes: ["economic"],
        whatChanged: "Fee structure change",
        whyItMattersForLineaNativeYield: "May impact yields",
        recommendedAction: "monitor",
        urgency: "this_week",
      });
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: `Here's my analysis:\n${JSON.stringify(validAssessment)}\nEnd of analysis.` }],
      });

      // Act
      const result = await client.analyzeProposal(createAnalysisRequest());

      // Assert
      expect(result).toEqual(validAssessment);
    });

    it("returns undefined when AI response fails schema validation", async () => {
      // Arrange
      const invalidAssessment = { riskScore: 150, impactTypes: ["invalid-type"] };
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: JSON.stringify(invalidAssessment) }],
      });

      // Act
      const result = await client.analyzeProposal(createAnalysisRequest());

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
      const result = await client.analyzeProposal(createAnalysisRequest());

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("No JSON found in AI response");
    });

    it("returns undefined on API failure", async () => {
      // Arrange
      mockAnthropicClient.messages.create.mockRejectedValue(new Error("API rate limit exceeded"));

      // Act
      const result = await client.analyzeProposal(createAnalysisRequest());

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
      const result = await client.analyzeProposal(createAnalysisRequest());

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
      const result = await client.analyzeProposal(createAnalysisRequest());

      // Assert
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to parse AI response as JSON", expect.any(Object));
    });

    it("includes proposal payload in user prompt when provided", async () => {
      // Arrange
      const validAssessment = createValidAssessment();
      mockAnthropicClient.messages.create.mockResolvedValue({
        content: [{ type: "text", text: JSON.stringify(validAssessment) }],
      });

      // Act
      await client.analyzeProposal(
        createAnalysisRequest({
          proposalType: "onchain_vote",
          proposalPayload: "0x1234567890abcdef",
        }),
      );

      // Assert
      expect(mockAnthropicClient.messages.create).toHaveBeenCalledWith(
        expect.objectContaining({
          messages: [
            expect.objectContaining({
              content: expect.stringContaining("Payload:\n0x1234567890abcdef"),
            }),
          ],
        }),
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
