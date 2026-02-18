import { describe, it, expect } from "@jest/globals";
import {
  BaseError as ViemBaseError,
  ChainDisconnectedError,
  ContractFunctionRevertedError,
  ExecutionRevertedError,
  HttpRequestError,
  InsufficientFundsError,
  FeeCapTooLowError,
  LimitExceededRpcError,
  NonceTooHighError,
  NonceTooLowError,
  RpcRequestError,
  TimeoutError,
  TipAboveFeeCapError,
  UserRejectedRequestError,
} from "viem";

import { DatabaseAccessError } from "../../../../domain/errors/DatabaseAccessError";
import { ErrorCode } from "../../../../domain/ports/IErrorParser";
import { DatabaseErrorType, DatabaseRepoName } from "../../../../domain/types/enums";
import { ViemErrorParser } from "../ViemErrorParser";

describe("ViemErrorParser", () => {
  const parser = new ViemErrorParser();

  describe("non-viem errors", () => {
    it("should return UNKNOWN_ERROR with shouldRetry=false for null/undefined", () => {
      const result = parser.parse(null);
      expect(result.errorCode).toBe(ErrorCode.UNKNOWN_ERROR);
      expect(result.mitigation.shouldRetry).toBe(false);
      expect(result.severity).toBe("error");
    });

    it("should return UNKNOWN_ERROR with shouldRetry=false for plain Error", () => {
      const result = parser.parse(new Error("something went wrong"));
      expect(result.errorCode).toBe(ErrorCode.UNKNOWN_ERROR);
      expect(result.errorMessage).toBe("something went wrong");
      expect(result.mitigation.shouldRetry).toBe(false);
      expect(result.severity).toBe("error");
    });

    it("should return UNKNOWN_ERROR with shouldRetry=false for string errors", () => {
      const result = parser.parse("raw string error");
      expect(result.errorCode).toBe(ErrorCode.UNKNOWN_ERROR);
      expect(result.errorMessage).toBe("raw string error");
      expect(result.mitigation.shouldRetry).toBe(false);
    });

    it("should return DATABASE_ERROR with shouldRetry=true for DatabaseAccessError", () => {
      const dbError = new DatabaseAccessError(
        DatabaseRepoName.MessageRepository,
        DatabaseErrorType.Read,
        new Error("connection lost"),
      );
      const result = parser.parse(dbError);
      expect(result.errorCode).toBe(ErrorCode.DATABASE_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
      expect(result.severity).toBe("warn");
    });
  });

  describe("viem transient errors", () => {
    it("should classify InsufficientFundsError as INSUFFICIENT_FUNDS with shouldRetry=true", () => {
      const inner = new InsufficientFundsError({ cause: new ViemBaseError("not enough") });
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.INSUFFICIENT_FUNDS);
      expect(result.mitigation.shouldRetry).toBe(true);
      expect(result.severity).toBe("warn");
    });

    it("should classify NonceTooLowError as NONCE_EXPIRED with shouldRetry=true", () => {
      const inner = new NonceTooLowError({ cause: new ViemBaseError("nonce too low"), nonce: 5 });
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NONCE_EXPIRED);
      expect(result.mitigation.shouldRetry).toBe(true);
      expect(result.severity).toBe("warn");
    });

    it("should classify NonceTooHighError as NONCE_EXPIRED with shouldRetry=true", () => {
      const inner = new NonceTooHighError({ cause: new ViemBaseError("nonce too high"), nonce: 100 });
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NONCE_EXPIRED);
      expect(result.mitigation.shouldRetry).toBe(true);
    });

    it("should classify FeeCapTooLowError as GAS_FEE_ERROR with shouldRetry=true", () => {
      const inner = new FeeCapTooLowError({ cause: new ViemBaseError("fee cap too low") });
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.GAS_FEE_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
      expect(result.severity).toBe("warn");
    });

    it("should classify TipAboveFeeCapError as GAS_FEE_ERROR with shouldRetry=true", () => {
      const inner = new TipAboveFeeCapError({ cause: new ViemBaseError("tip above fee cap") });
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.GAS_FEE_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
    });

    it("should classify HttpRequestError as NETWORK_ERROR with shouldRetry=true", () => {
      const error = new HttpRequestError({ url: "http://rpc.test", body: {} });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NETWORK_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
      expect(result.severity).toBe("warn");
    });

    it("should classify TimeoutError as NETWORK_ERROR with shouldRetry=true", () => {
      const error = new TimeoutError({ body: {}, url: "http://rpc.test" });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NETWORK_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
    });

    it("should classify ChainDisconnectedError as NETWORK_ERROR with shouldRetry=true", () => {
      const inner = new ChainDisconnectedError(new ViemBaseError("chain disconnected"));
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NETWORK_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
    });

    it("should classify LimitExceededRpcError as NETWORK_ERROR with shouldRetry=true", () => {
      const inner = new LimitExceededRpcError(new ViemBaseError("rate limited"));
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NETWORK_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
    });
  });

  describe("viem permanent errors", () => {
    it("should classify ContractFunctionRevertedError as EXECUTION_REVERTED with shouldRetry=false", () => {
      const inner = new ContractFunctionRevertedError({
        abi: [],
        functionName: "claim",
      });
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.EXECUTION_REVERTED);
      expect(result.mitigation.shouldRetry).toBe(false);
      expect(result.severity).toBe("error");
    });

    it("should classify ExecutionRevertedError as EXECUTION_REVERTED with shouldRetry=false", () => {
      const inner = new ExecutionRevertedError({ cause: new ViemBaseError("reverted") });
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.EXECUTION_REVERTED);
      expect(result.mitigation.shouldRetry).toBe(false);
      expect(result.severity).toBe("error");
    });

    it("should classify UserRejectedRequestError as ACTION_REJECTED with shouldRetry=false", () => {
      const inner = new UserRejectedRequestError(new ViemBaseError("user rejected"));
      const error = new ViemBaseError("wrapper", { cause: inner });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.ACTION_REJECTED);
      expect(result.mitigation.shouldRetry).toBe(false);
      expect(result.severity).toBe("error");
    });
  });

  describe("RPC errors", () => {
    it("should classify RpcRequestError with code -32603 as NETWORK_ERROR", () => {
      const rpcError = new RpcRequestError({
        body: {},
        url: "http://rpc.test",
        error: { code: -32603, message: "internal error" },
      });
      const error = new ViemBaseError("wrapper", { cause: rpcError });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NETWORK_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
    });

    it("should classify RpcRequestError with code -32000 as NETWORK_ERROR", () => {
      const rpcError = new RpcRequestError({
        body: {},
        url: "http://rpc.test",
        error: { code: -32000, message: "server error" },
      });
      const error = new ViemBaseError("wrapper", { cause: rpcError });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NETWORK_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
    });

    it("should classify RpcRequestError with code -32005 as NETWORK_ERROR", () => {
      const rpcError = new RpcRequestError({
        body: {},
        url: "http://rpc.test",
        error: { code: -32005, message: "limit exceeded" },
      });
      const error = new ViemBaseError("wrapper", { cause: rpcError });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.NETWORK_ERROR);
      expect(result.mitigation.shouldRetry).toBe(true);
    });

    it("should classify unknown RPC error codes as UNKNOWN_ERROR with shouldRetry=false", () => {
      const rpcError = new RpcRequestError({
        body: {},
        url: "http://rpc.test",
        error: { code: -32601, message: "method not found" },
      });
      const error = new ViemBaseError("wrapper", { cause: rpcError });
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.UNKNOWN_ERROR);
      expect(result.mitigation.shouldRetry).toBe(false);
    });
  });

  describe("unknown viem errors", () => {
    it("should default to UNKNOWN_ERROR with shouldRetry=false for unrecognized viem errors", () => {
      const error = new ViemBaseError("something unexpected happened");
      const result = parser.parse(error);
      expect(result.errorCode).toBe(ErrorCode.UNKNOWN_ERROR);
      expect(result.mitigation.shouldRetry).toBe(false);
      expect(result.severity).toBe("error");
    });
  });

  describe("severity mapping", () => {
    it("should return warn for transient errors", () => {
      const dbError = new DatabaseAccessError(
        DatabaseRepoName.MessageRepository,
        DatabaseErrorType.Read,
        new Error("timeout"),
      );
      expect(parser.parse(dbError).severity).toBe("warn");
    });

    it("should return error for permanent errors", () => {
      expect(parser.parse(new Error("unexpected")).severity).toBe("error");
    });
  });
});
