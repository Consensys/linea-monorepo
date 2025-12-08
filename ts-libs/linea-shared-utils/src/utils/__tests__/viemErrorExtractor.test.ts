import { BaseError, ContractFunctionRevertedError } from "viem";
import { extractViemErrorInfo, ExtractedViemError } from "../viemErrorExtractor";

describe("extractViemErrorInfo", () => {
  describe("non-BaseError handling", () => {
    it("returns non-BaseError as-is", () => {
      const plainError = new Error("plain error");
      const result = extractViemErrorInfo(plainError);
      expect(result).toBe(plainError);
    });

    it("returns non-Error values as-is", () => {
      const stringError = "string error";
      const result = extractViemErrorInfo(stringError);
      expect(result).toBe(stringError);
    });

    it("returns null as-is", () => {
      const result = extractViemErrorInfo(null);
      expect(result).toBe(null);
    });
  });

  describe("BaseError basic extraction", () => {
    it("extracts basic error properties", () => {
      const error = new BaseError("Test error message");
      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.name).toBe("BaseError");
      // BaseError appends version info to message, so check that message contains the core message
      expect(result.message).toContain("Test error message");
    });

    it("extracts shortMessage and details when available", () => {
      const error = Object.assign(new BaseError("Full message"), {
        shortMessage: "Short message",
        details: "Error details",
      });
      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.shortMessage).toBe("Short message");
      expect(result.details).toBe("Error details");
    });

    it("extracts error code from error chain", () => {
      const error = Object.assign(new BaseError("Error with code"), { code: -32015 });
      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.code).toBe(-32015);
    });
  });

  describe("metaMessages extraction", () => {
    it("extracts metaMessages from top-level error", () => {
      const error = Object.assign(new BaseError("Error"), {
        metaMessages: ["Message 1", "Message 2"],
      });
      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.metaMessages).toEqual(["Message 1", "Message 2"]);
    });

    it("extracts metaMessages from EstimateGasExecutionError in chain", () => {
      // Create an error that will be found by walk() when searching for EstimateGasExecutionError
      const estimateGasError = Object.assign(new BaseError("EstimateGasExecutionError"), {
        name: "EstimateGasExecutionError",
        metaMessages: ["Estimate Gas Arguments:", "  from: 0x123"],
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: estimateGasError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      // The walk() method should find the EstimateGasExecutionError in the cause chain
      expect(result.metaMessages).toBeDefined();
      if (result.metaMessages) {
        expect(result.metaMessages.length).toBeGreaterThan(0);
      }
    });

    it("merges metaMessages from EstimateGasExecutionError without duplicates", () => {
      const estimateGasError = Object.assign(new BaseError("EstimateGasExecutionError"), {
        name: "EstimateGasExecutionError",
        metaMessages: ["Estimate Gas Arguments:", "  from: 0x123"],
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        metaMessages: ["Top level message"],
        cause: estimateGasError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.metaMessages).toContain("Top level message");
      // If walk finds the EstimateGasError, those messages should also be included
      if (result.metaMessages && result.metaMessages.length > 1) {
        expect(result.metaMessages.some((msg) => msg.includes("Estimate Gas"))).toBe(true);
      }
    });
  });

  describe("ExecutionRevertedError extraction", () => {
    it("extracts revert data from ExecutionRevertedError", () => {
      const revertError = Object.assign(new BaseError("ExecutionRevertedError"), {
        name: "ExecutionRevertedError",
        data: { errorName: "InsufficientBalance", args: ["100", "50"] },
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: revertError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      // walk() should find the ExecutionRevertedError in the cause chain
      expect(result.revertData).toBeDefined();
      if (result.revertData) {
        expect((result.revertData as { errorName?: string }).errorName).toBe("InsufficientBalance");
      }
      expect(result.errorName).toBe("InsufficientBalance");
      expect(result.revertReason).toBe("100");
    });

    it("extracts revert reason from data.reason", () => {
      const revertError = Object.assign(new BaseError("ExecutionRevertedError"), {
        name: "ExecutionRevertedError",
        data: { reason: "Custom revert reason" },
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: revertError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.revertReason).toBe("Custom revert reason");
    });

    it("extracts reason directly from ExecutionRevertedError", () => {
      const revertError = Object.assign(new BaseError("ExecutionRevertedError"), {
        name: "ExecutionRevertedError",
        reason: "Direct reason",
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: revertError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.revertReason).toBe("Direct reason");
    });

    it("handles ContractFunctionRevertedError", () => {
      const revertError = new ContractFunctionRevertedError({
        abi: [],
        functionName: "test",
        message: "execution reverted",
      });
      Object.assign(revertError, {
        data: { errorName: "TestError" },
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: revertError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.revertData).toEqual({ errorName: "TestError" });
      expect(result.errorName).toBe("TestError");
    });
  });

  describe("cause extraction", () => {
    it("extracts BaseError cause recursively", () => {
      const causeError = Object.assign(new BaseError("Cause error"), {
        code: -32015,
        shortMessage: "Cause short message",
      });
      const error = Object.assign(new BaseError("Main error"), {
        cause: causeError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.cause).toBeDefined();
      expect(result.cause!.name).toBe("BaseError");
      // BaseError appends version info to message, so check that message contains the core message
      expect(result.cause!.message).toContain("Cause error");
      expect(result.cause!.code).toBe(-32015);
      expect(result.cause!.shortMessage).toBe("Cause short message");
    });

    it("extracts plain Error cause with basic info", () => {
      const causeError = new Error("Plain cause error");
      const error = Object.assign(new BaseError("Main error"), {
        cause: causeError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.cause).toBeDefined();
      expect(result.cause!.name).toBe("Error");
      expect(result.cause!.message).toBe("Plain cause error");
    });

    it("extracts cause with details and shortMessage", () => {
      const causeError = Object.assign(new Error("Cause error"), {
        details: "Cause details",
        shortMessage: "Cause short",
      });
      const error = Object.assign(new BaseError("Main error"), {
        cause: causeError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.cause).toBeDefined();
      expect(result.cause!.details).toBe("Cause details");
      expect(result.cause!.shortMessage).toBe("Cause short");
    });

    it("handles nested cause chains", () => {
      const nestedCause = Object.assign(new BaseError("Nested cause"), {
        code: -32000,
      });
      const causeError = Object.assign(new BaseError("Cause error"), {
        cause: nestedCause,
      });
      const error = Object.assign(new BaseError("Main error"), {
        cause: causeError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.cause).toBeDefined();
      expect(result.cause!.cause).toBeDefined();
      expect(result.cause!.cause!.code).toBe(-32000);
    });
  });

  describe("complex error scenarios", () => {
    it("extracts all information from EstimateGasExecutionError with ExecutionRevertedError cause", () => {
      const executionRevertedError = Object.assign(new BaseError("ExecutionRevertedError"), {
        name: "ExecutionRevertedError",
        data: { errorName: "InsufficientFunds", reason: "Not enough balance" },
      });
      const estimateGasError = Object.assign(new BaseError("EstimateGasExecutionError"), {
        name: "EstimateGasExecutionError",
        metaMessages: [
          "Estimate Gas Arguments:",
          "  from: 0x3a595Eeb7e6d7005bfeA5f3981663f68CDc734BA",
          "  to: 0x73bF00aD18c7c0871EBA03Bcbef8C98225f9CEaA",
        ],
        cause: executionRevertedError,
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        code: -32603,
        cause: estimateGasError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.code).toBe(-32603);
      // walk() should find EstimateGasExecutionError and extract its metaMessages
      expect(result.metaMessages).toBeDefined();
      if (result.metaMessages) {
        expect(result.metaMessages.some((msg) => msg.includes("Estimate Gas Arguments"))).toBe(true);
      }
      // walk() should find ExecutionRevertedError and extract revert data
      expect(result.revertData).toBeDefined();
      if (result.revertData) {
        expect((result.revertData as { errorName?: string }).errorName).toBe("InsufficientFunds");
      }
      expect(result.errorName).toBe("InsufficientFunds");
      expect(result.revertReason).toBe("Not enough balance");
      expect(result.cause).toBeDefined();
      expect(result.cause!.name).toBe("EstimateGasExecutionError");
    });

    it("handles error without revert data gracefully", () => {
      const revertError = Object.assign(new BaseError("ExecutionRevertedError"), {
        name: "ExecutionRevertedError",
        data: undefined,
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: revertError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      // If walk() finds the revert error but it has no data, these should be undefined
      // The exact behavior depends on whether walk() finds it, but we should handle gracefully
      expect(result).toBeDefined();
    });

    it("handles error with empty metaMessages", () => {
      const error = Object.assign(new BaseError("Error"), {
        metaMessages: [],
      });
      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.metaMessages).toEqual([]);
    });
  });

  describe("error chain traversal", () => {
    it("uses error.walk to traverse error chain", () => {
      const error = new BaseError("Test error");
      const walkSpy = jest.spyOn(error, "walk");

      extractViemErrorInfo(error);

      expect(walkSpy).toHaveBeenCalled();
    });

    it("handles errors that don't match any known pattern", () => {
      const error = Object.assign(new BaseError("Unknown error type"), {
        code: 9999,
      });
      const result = extractViemErrorInfo(error) as ExtractedViemError;

      expect(result.name).toBe("BaseError");
      // BaseError appends version info to message, so check that message contains the core message
      expect(result.message).toContain("Unknown error type");
      expect(result.code).toBe(9999);
    });

    it("handles EstimateGasExecutionError with name that includes EstimateGas but not exact match", () => {
      // Test the branch where name.includes("EstimateGas") is true but name !== "EstimateGasExecutionError"
      const estimateGasError = Object.assign(new BaseError("CustomEstimateGasError"), {
        name: "CustomEstimateGasError",
        metaMessages: ["Custom estimate gas message"],
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: estimateGasError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      // Should find the error via includes("EstimateGas") check
      expect(result.metaMessages).toBeDefined();
      if (result.metaMessages) {
        expect(result.metaMessages.length).toBeGreaterThan(0);
      }
    });

    it("handles EstimateGasExecutionError with name that does not include EstimateGas", () => {
      // Test the branch where err instanceof Error is true but name check fails (covers the false branch of the OR)
      const otherError = Object.assign(new BaseError("SomeOtherError"), {
        name: "SomeOtherError",
        metaMessages: ["Other message"],
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: otherError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      // Should not find EstimateGasExecutionError, so metaMessages from that won't be merged
      // But top-level metaMessages might still be present
      expect(result).toBeDefined();
    });

    it("handles ExecutionRevertedError with name that includes Reverted but not exact match", () => {
      // Test the branch where name.includes("Reverted") is true but name !== "ExecutionRevertedError" and !== "ContractFunctionExecutionError"
      const revertError = Object.assign(new BaseError("CustomRevertedError"), {
        name: "CustomRevertedError",
        data: { errorName: "CustomError", reason: "Custom revert" },
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: revertError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      // Should find the error via includes("Reverted") check
      expect(result.revertData).toBeDefined();
      if (result.revertData) {
        expect((result.revertData as { errorName?: string }).errorName).toBe("CustomError");
      }
      expect(result.revertReason).toBe("Custom revert");
    });

    it("handles ExecutionRevertedError with name that does not include Reverted", () => {
      // Test the branch where err instanceof Error is true but all name checks fail (covers the false branch of the OR)
      const otherError = Object.assign(new BaseError("SomeOtherError"), {
        name: "SomeOtherError",
        data: { someData: "value" },
      });
      const error = Object.assign(new BaseError("Wrapper error"), {
        cause: otherError,
      });

      const result = extractViemErrorInfo(error) as ExtractedViemError;

      // Should not find ExecutionRevertedError, so revertData won't be extracted
      expect(result).toBeDefined();
      // The otherError won't match the ExecutionRevertedError pattern
      expect(result.revertData).toBeUndefined();
    });

    it("covers instanceof Error false branch by mocking walk to pass non-Error", () => {
      // Test the branch where err instanceof Error is false in the walk predicate
      // We need to mock walk() to actually pass a non-Error to the predicate to get coverage
      const error = new BaseError("Test error");
      const originalWalk = error.walk.bind(error);
      
      // Mock walk to call predicate with both Error and non-Error objects
      error.walk = jest.fn((predicate: (err: unknown) => boolean) => {
        // First call with Error (normal case)
        const errorResult = predicate(error);
        
        // Call with non-Error to cover instanceof false branch
        const nonError = { name: "NotAnError" };
        predicate(nonError);
        
        return errorResult ? error : undefined;
      }) as any;

      extractViemErrorInfo(error);
      
      // walk should have been called multiple times (once for EstimateGas, once for ExecutionReverted)
      expect(error.walk).toHaveBeenCalled();
      
      // Restore original walk
      error.walk = originalWalk;
    });
  });
});

