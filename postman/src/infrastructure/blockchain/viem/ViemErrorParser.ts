import {
  BaseError as ViemBaseError,
  HttpRequestError,
  TimeoutError,
  ContractFunctionRevertedError,
  RpcRequestError,
  UserRejectedRequestError,
  TransactionRejectedRpcError,
  InternalRpcError,
  LimitExceededRpcError,
  ResourceUnavailableRpcError,
} from "viem";

import { DatabaseAccessError } from "../../../core/errors/DatabaseErrors";
import { IErrorParser, ParsedError } from "../../../core/errors/IErrorParser";

export class ViemErrorParser implements IErrorParser {
  parse(error: unknown): ParsedError {
    if (error instanceof DatabaseAccessError) {
      return { retryable: true, message: error.message };
    }

    if (!(error instanceof ViemBaseError)) {
      if (error instanceof Error) {
        return { retryable: true, message: error.message };
      }
      return { retryable: true, message: String(error ?? "") };
    }

    const message = error.shortMessage || error.message;

    if (error.walk((e) => e instanceof ContractFunctionRevertedError)) {
      return { retryable: false, message };
    }

    if (error.walk((e) => e instanceof UserRejectedRequestError)) {
      return { retryable: false, message };
    }

    if (error.walk((e) => e instanceof TransactionRejectedRpcError)) {
      return { retryable: false, message };
    }

    if (error.walk((e) => e instanceof HttpRequestError)) {
      return { retryable: true, message };
    }

    if (error.walk((e) => e instanceof TimeoutError)) {
      return { retryable: true, message };
    }

    if (error.walk((e) => e instanceof InternalRpcError)) {
      return { retryable: true, message };
    }

    if (error.walk((e) => e instanceof LimitExceededRpcError)) {
      return { retryable: true, message };
    }

    if (error.walk((e) => e instanceof ResourceUnavailableRpcError)) {
      return { retryable: true, message };
    }

    const rpcError = error.walk((e) => e instanceof RpcRequestError);
    if (rpcError && rpcError instanceof RpcRequestError) {
      const details = rpcError.details?.toLowerCase();
      if (details?.includes("execution reverted")) {
        return { retryable: false, message };
      }
      if (details?.includes("nonce too low") || details?.includes("already known")) {
        return { retryable: true, message };
      }
    }

    return { retryable: true, message };
  }
}
