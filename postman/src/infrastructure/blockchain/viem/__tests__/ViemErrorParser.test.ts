import { describe, it, expect } from "@jest/globals";
import {
  HttpRequestError,
  TimeoutError,
  ContractFunctionRevertedError,
  RpcRequestError,
  UserRejectedRequestError,
  TransactionRejectedRpcError,
  InternalRpcError,
  LimitExceededRpcError,
  ResourceUnavailableRpcError,
  ContractFunctionExecutionError,
  BaseError as ViemBaseError,
} from "viem";

import { DatabaseErrorType, DatabaseRepoName } from "../../../../core/enums";
import { DatabaseAccessError } from "../../../../core/errors/DatabaseErrors";
import { ViemErrorParser } from "../ViemErrorParser";

describe("ViemErrorParser", () => {
  const parser = new ViemErrorParser();

  describe("DatabaseAccessError", () => {
    it("should return retryable for DatabaseAccessError", () => {
      const error = new DatabaseAccessError(
        DatabaseRepoName.MessageRepository,
        DatabaseErrorType.Read,
        new Error("connection timeout"),
      );
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
      expect(result.message).toContain("connection timeout");
    });
  });

  describe("Non-viem errors", () => {
    it("should return retryable for generic Error", () => {
      const error = new Error("generic error");
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
      expect(result.message).toBe("generic error");
    });

    it("should return retryable for null", () => {
      const result = parser.parse(null);
      expect(result.retryable).toBe(true);
      expect(result.message).toBe("");
    });

    it("should return retryable for undefined", () => {
      const result = parser.parse(undefined);
      expect(result.retryable).toBe(true);
      expect(result.message).toBe("");
    });

    it("should return retryable for string error", () => {
      const result = parser.parse("something went wrong");
      expect(result.retryable).toBe(true);
      expect(result.message).toBe("something went wrong");
    });
  });

  describe("Retryable viem errors", () => {
    it("should return retryable for HttpRequestError", () => {
      const error = new HttpRequestError({ url: "http://localhost" });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("should return retryable for TimeoutError", () => {
      const error = new TimeoutError({ body: {}, url: "http://localhost" });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("should return retryable for InternalRpcError", () => {
      const error = new InternalRpcError(new Error("internal"));
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("should return retryable for LimitExceededRpcError", () => {
      const error = new LimitExceededRpcError(new Error("rate limited"));
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("should return retryable for ResourceUnavailableRpcError", () => {
      const error = new ResourceUnavailableRpcError(new Error("unavailable"));
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });
  });

  describe("Non-retryable viem errors", () => {
    it("should return not retryable for UserRejectedRequestError", () => {
      const error = new UserRejectedRequestError(new Error("rejected"));
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });

    it("should return not retryable for TransactionRejectedRpcError", () => {
      const error = new TransactionRejectedRpcError(new Error("rejected"));
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });
  });

  describe("RpcRequestError with details", () => {
    it("should return not retryable for execution reverted", () => {
      const error = new RpcRequestError({
        body: {},
        url: "http://localhost",
        error: { code: -32000, message: "execution reverted" },
      });
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });

    it("should return retryable for nonce too low", () => {
      const error = new RpcRequestError({
        body: {},
        url: "http://localhost",
        error: { code: -32000, message: "nonce too low" },
      });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("should return retryable for already known", () => {
      const error = new RpcRequestError({
        body: {},
        url: "http://localhost",
        error: { code: -32000, message: "already known" },
      });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });
  });

  describe("ContractFunctionRevertedError (nested)", () => {
    it("should return not retryable for contract revert via ContractFunctionExecutionError", () => {
      const revertError = new ContractFunctionRevertedError({
        abi: [],
        functionName: "test",
      });
      const executionError = new ContractFunctionExecutionError(revertError, {
        abi: [],
        functionName: "test",
      });
      const result = parser.parse(executionError);
      expect(result.retryable).toBe(false);
    });
  });

  describe("Default viem error", () => {
    it("should return retryable for unknown ViemBaseError", () => {
      const error = new ViemBaseError("unknown viem error");
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
      expect(result.message).toBe("unknown viem error");
    });
  });

  describe("Message extraction", () => {
    it("should use shortMessage when available", () => {
      const error = new HttpRequestError({ url: "http://localhost" });
      const result = parser.parse(error);
      expect(result.message).toBe(error.shortMessage || error.message);
    });
  });
});
