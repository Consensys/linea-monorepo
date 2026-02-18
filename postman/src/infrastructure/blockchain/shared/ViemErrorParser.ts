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

import { DatabaseAccessError } from "../../../domain/errors/DatabaseAccessError";
import { ErrorCode } from "../../../domain/ports/IErrorParser";

import type { IErrorParser, ParsedErrorResult, Severity } from "../../../domain/ports/IErrorParser";

const TRANSIENT_ERROR_CODES: ReadonlySet<ErrorCode> = new Set([
  ErrorCode.NETWORK_ERROR,
  ErrorCode.NONCE_EXPIRED,
  ErrorCode.GAS_FEE_ERROR,
  ErrorCode.DATABASE_ERROR,
  ErrorCode.INSUFFICIENT_FUNDS,
]);

function severityFor(code: ErrorCode): Severity {
  return TRANSIENT_ERROR_CODES.has(code) ? "warn" : "error";
}

function result(errorCode: ErrorCode, errorMessage: string, shouldRetry: boolean, data?: string): ParsedErrorResult {
  return {
    errorCode,
    errorMessage,
    severity: severityFor(errorCode),
    mitigation: { shouldRetry },
    ...(data !== undefined ? { data } : {}),
  };
}

export class ViemErrorParser implements IErrorParser {
  public parse(error: unknown): ParsedErrorResult {
    if (!error) {
      return result(ErrorCode.UNKNOWN_ERROR, "No error provided", false);
    }

    if (error instanceof DatabaseAccessError) {
      return result(ErrorCode.DATABASE_ERROR, error.message, true);
    }

    if (error instanceof ViemBaseError) {
      return this.parseViemError(error);
    }

    return result(ErrorCode.UNKNOWN_ERROR, error instanceof Error ? error.message : String(error), false);
  }

  private parseViemError(error: ViemBaseError): ParsedErrorResult {
    const insufficientFunds = error.walk((e) => e instanceof InsufficientFundsError);
    if (insufficientFunds) {
      return result(ErrorCode.INSUFFICIENT_FUNDS, (insufficientFunds as ViemBaseError).shortMessage, true);
    }

    const nonceTooLow = error.walk((e) => e instanceof NonceTooLowError);
    if (nonceTooLow) {
      return result(ErrorCode.NONCE_EXPIRED, (nonceTooLow as ViemBaseError).shortMessage, true);
    }

    const nonceTooHigh = error.walk((e) => e instanceof NonceTooHighError);
    if (nonceTooHigh) {
      return result(ErrorCode.NONCE_EXPIRED, (nonceTooHigh as ViemBaseError).shortMessage, true);
    }

    const feeCapTooLow = error.walk((e) => e instanceof FeeCapTooLowError);
    if (feeCapTooLow) {
      return result(ErrorCode.GAS_FEE_ERROR, (feeCapTooLow as ViemBaseError).shortMessage, true);
    }

    const tipAboveFeeCap = error.walk((e) => e instanceof TipAboveFeeCapError);
    if (tipAboveFeeCap) {
      return result(ErrorCode.GAS_FEE_ERROR, (tipAboveFeeCap as ViemBaseError).shortMessage, true);
    }

    const revert = error.walk((e) => e instanceof ContractFunctionRevertedError);
    if (revert) {
      const revertError = revert as ContractFunctionRevertedError;
      return result(
        ErrorCode.EXECUTION_REVERTED,
        revertError.reason ?? revertError.shortMessage,
        false,
        revertError.data?.errorName,
      );
    }

    const executionReverted = error.walk((e) => e instanceof ExecutionRevertedError);
    if (executionReverted) {
      return result(ErrorCode.EXECUTION_REVERTED, (executionReverted as ViemBaseError).shortMessage, false);
    }

    if (error instanceof HttpRequestError || error instanceof TimeoutError) {
      return result(ErrorCode.NETWORK_ERROR, error.shortMessage, true);
    }

    const chainDisconnected = error.walk((e) => e instanceof ChainDisconnectedError);
    if (chainDisconnected) {
      return result(ErrorCode.NETWORK_ERROR, (chainDisconnected as ViemBaseError).shortMessage, true);
    }

    const limitExceeded = error.walk((e) => e instanceof LimitExceededRpcError);
    if (limitExceeded) {
      return result(ErrorCode.NETWORK_ERROR, (limitExceeded as ViemBaseError).shortMessage, true);
    }

    const rpcError = error.walk((e) => e instanceof RpcRequestError);
    if (rpcError) {
      return this.parseRpcError(rpcError as RpcRequestError);
    }

    const userRejected = error.walk((e) => e instanceof UserRejectedRequestError);
    if (userRejected) {
      return result(ErrorCode.ACTION_REJECTED, (userRejected as ViemBaseError).shortMessage, false);
    }

    return result(ErrorCode.UNKNOWN_ERROR, error.shortMessage ?? error.message, false);
  }

  private parseRpcError(error: RpcRequestError): ParsedErrorResult {
    const code = error.code;

    if (code === -32603) {
      return result(ErrorCode.NETWORK_ERROR, error.shortMessage, true);
    }

    if (code === -32000) {
      return result(ErrorCode.NETWORK_ERROR, error.shortMessage, true);
    }

    if (code === -32005) {
      return result(ErrorCode.NETWORK_ERROR, error.shortMessage, true);
    }

    return result(ErrorCode.UNKNOWN_ERROR, error.shortMessage, false);
  }
}
