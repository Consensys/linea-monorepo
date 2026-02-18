import {
  BaseError as ViemBaseError,
  ContractFunctionRevertedError,
  ExecutionRevertedError,
  HttpRequestError,
  InsufficientFundsError,
  FeeCapTooLowError,
  NonceTooHighError,
  NonceTooLowError,
  TimeoutError,
  TipAboveFeeCapError,
  UserRejectedRequestError,
} from "viem";

import { DatabaseAccessError } from "../../../domain/errors/DatabaseAccessError";
import { ErrorCode } from "../../../domain/ports/IErrorParser";

import type { IErrorParser, ParsedErrorResult } from "../../../domain/ports/IErrorParser";

export class ViemErrorParser implements IErrorParser {
  public parseErrorWithMitigation(error: unknown): ParsedErrorResult | null {
    if (!error) {
      return null;
    }

    if (error instanceof DatabaseAccessError) {
      return {
        errorCode: ErrorCode.DATABASE_ERROR,
        errorMessage: error.message,
        mitigation: { shouldRetry: true },
      };
    }

    if (error instanceof ViemBaseError) {
      return this.parseViemError(error);
    }

    return {
      errorCode: ErrorCode.UNKNOWN_ERROR,
      errorMessage: error instanceof Error ? error.message : String(error),
      mitigation: { shouldRetry: false },
    };
  }

  private parseViemError(error: ViemBaseError): ParsedErrorResult {
    const insufficientFunds = error.walk((e) => e instanceof InsufficientFundsError);
    if (insufficientFunds) {
      return {
        errorCode: ErrorCode.INSUFFICIENT_FUNDS,
        errorMessage: (insufficientFunds as ViemBaseError).shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    const nonceTooLow = error.walk((e) => e instanceof NonceTooLowError);
    if (nonceTooLow) {
      return {
        errorCode: ErrorCode.NONCE_EXPIRED,
        errorMessage: (nonceTooLow as ViemBaseError).shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    const nonceTooHigh = error.walk((e) => e instanceof NonceTooHighError);
    if (nonceTooHigh) {
      return {
        errorCode: ErrorCode.NONCE_EXPIRED,
        errorMessage: (nonceTooHigh as ViemBaseError).shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    const feeCapTooLow = error.walk((e) => e instanceof FeeCapTooLowError);
    if (feeCapTooLow) {
      return {
        errorCode: ErrorCode.GAS_FEE_ERROR,
        errorMessage: (feeCapTooLow as ViemBaseError).shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    const tipAboveFeeCap = error.walk((e) => e instanceof TipAboveFeeCapError);
    if (tipAboveFeeCap) {
      return {
        errorCode: ErrorCode.GAS_FEE_ERROR,
        errorMessage: (tipAboveFeeCap as ViemBaseError).shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    const revert = error.walk((e) => e instanceof ContractFunctionRevertedError);
    if (revert) {
      const revertError = revert as ContractFunctionRevertedError;
      return {
        errorCode: ErrorCode.EXECUTION_REVERTED,
        errorMessage: revertError.reason ?? revertError.shortMessage,
        data: revertError.data?.errorName,
        mitigation: { shouldRetry: false },
      };
    }

    const executionReverted = error.walk((e) => e instanceof ExecutionRevertedError);
    if (executionReverted) {
      return {
        errorCode: ErrorCode.EXECUTION_REVERTED,
        errorMessage: (executionReverted as ViemBaseError).shortMessage,
        mitigation: { shouldRetry: false },
      };
    }

    if (error instanceof HttpRequestError || error instanceof TimeoutError) {
      return {
        errorCode: ErrorCode.NETWORK_ERROR,
        errorMessage: error.shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    const userRejected = error.walk((e) => e instanceof UserRejectedRequestError);
    if (userRejected) {
      return {
        errorCode: ErrorCode.ACTION_REJECTED,
        errorMessage: (userRejected as ViemBaseError).shortMessage,
        mitigation: { shouldRetry: false },
      };
    }

    if (error.shortMessage?.includes("nonce") || error.message?.includes("replacement underpriced")) {
      return {
        errorCode: ErrorCode.NONCE_EXPIRED,
        errorMessage: error.shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    return {
      errorCode: ErrorCode.UNKNOWN_ERROR,
      errorMessage: error.shortMessage ?? error.message,
      mitigation: { shouldRetry: true },
    };
  }
}
