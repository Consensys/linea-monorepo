import { describe, it, expect } from "@jest/globals";
import {
  BaseError,
  ContractFunctionRevertedError,
  HttpRequestError,
  TimeoutError,
  TransactionRejectedRpcError,
} from "viem";

import { DatabaseErrorType, DatabaseRepoName } from "../../../../core/enums";
import { DatabaseAccessError } from "../../../../core/errors/DatabaseErrors";
import { ViemErrorParser } from "../ViemErrorParser";

describe("ViemErrorParser", () => {
  const parser = new ViemErrorParser();

  describe("non-Error values", () => {
    it("marks plain string as not retryable", () => {
      const result = parser.parse("something went wrong");
      expect(result.retryable).toBe(false);
      expect(result.message).toBe("something went wrong");
    });

    it("marks plain Error as not retryable", () => {
      const result = parser.parse(new Error("generic error"));
      expect(result.retryable).toBe(false);
      expect(result.message).toBe("generic error");
    });
  });

  describe("DatabaseAccessError", () => {
    it("marks DatabaseAccessError as retryable", () => {
      const dbError = new DatabaseAccessError(
        DatabaseRepoName.MessageRepository,
        DatabaseErrorType.Read,
        new Error("db read failed"),
      );
      const result = parser.parse(dbError);
      expect(result.retryable).toBe(true);
    });
  });

  describe("viem BaseError subtypes", () => {
    it("marks HttpRequestError as retryable", () => {
      const error = new HttpRequestError({ url: "http://localhost:8545", status: 503, body: "" });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("marks TimeoutError as retryable", () => {
      const error = new TimeoutError({ body: {}, url: "http://localhost:8545" });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("marks ContractFunctionRevertedError as not retryable", () => {
      const error = new ContractFunctionRevertedError({
        abi: [],
        functionName: "claimMessage",
        data: undefined,
      });
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });

    it("marks TransactionRejectedRpcError as not retryable", () => {
      const error = new TransactionRejectedRpcError(new Error("user denied"));
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });

    it("marks BaseError with RPC code -32603 as retryable", () => {
      const inner = Object.assign(new Error("Internal error"), { code: -32603 });
      const error = new BaseError("wrapped", { cause: inner });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("marks BaseError with RPC code 4001 as not retryable", () => {
      const inner = Object.assign(new Error("User rejected"), { code: 4001 });
      const error = new BaseError("wrapped", { cause: inner });
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });

    it("marks -32000 'execution reverted' as not retryable", () => {
      const inner = Object.assign(new Error("execution reverted"), { code: -32000 });
      const error = new BaseError("wrapped", { cause: inner });
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });

    it("marks -32000 'gas required exceeds allowance' as retryable", () => {
      const inner = Object.assign(new Error("gas required exceeds allowance (0)"), { code: -32000 });
      const error = new BaseError("wrapped", { cause: inner });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("marks -32000 'max fee per gas less than block base fee' as retryable", () => {
      const inner = Object.assign(new Error("max fee per gas less than block base fee"), { code: -32000 });
      const error = new BaseError("wrapped", { cause: inner });
      const result = parser.parse(error);
      expect(result.retryable).toBe(true);
    });

    it("marks unknown BaseError as not retryable", () => {
      const error = new BaseError("unknown viem error");
      const result = parser.parse(error);
      expect(result.retryable).toBe(false);
    });
  });
});
