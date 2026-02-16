import {
  BaseError as ViemBaseError,
  ContractFunctionRevertedError,
  EstimateGasExecutionError,
  TransactionNotFoundError,
  HttpRequestError,
  TimeoutError,
  RpcRequestError,
} from "viem";

import { DatabaseAccessError } from "../../../domain/errors/DatabaseAccessError";

import type { IErrorParser, ParsedErrorResult } from "../../../domain/ports/IErrorParser";

export class ViemErrorParser implements IErrorParser {
  public parseErrorWithMitigation(error: unknown): ParsedErrorResult | null {
    if (!error) {
      return null;
    }

    if (error instanceof DatabaseAccessError) {
      return {
        errorCode: "UNKNOWN_ERROR",
        errorMessage: error.message,
        mitigation: { shouldRetry: true },
      };
    }

    if (error instanceof ViemBaseError) {
      return this.parseViemError(error);
    }

    return {
      errorCode: "UNKNOWN_ERROR",
      errorMessage: error instanceof Error ? error.message : String(error),
      mitigation: { shouldRetry: false },
    };
  }

  private parseViemError(error: ViemBaseError): ParsedErrorResult {
    if (error instanceof HttpRequestError || error instanceof TimeoutError) {
      return {
        errorCode: "NETWORK_ERROR",
        errorMessage: error.message,
        mitigation: { shouldRetry: true },
      };
    }

    if (error instanceof RpcRequestError) {
      const rpcCode = error.code;

      if (rpcCode === -32000) {
        const msg = error.message.toLowerCase();
        if (
          msg.includes("gas required exceeds allowance") ||
          msg.includes("max priority fee per gas higher than max fee per gas") ||
          msg.includes("max fee per gas less than block base fee") ||
          msg.includes("below sender account nonce") ||
          msg.includes("nonce too low") ||
          msg.includes("insufficient funds")
        ) {
          return {
            errorCode: "CALL_EXCEPTION",
            errorMessage: error.message,
            mitigation: { shouldRetry: true },
          };
        }

        if (msg.includes("execution reverted")) {
          return {
            errorCode: "CALL_EXCEPTION",
            errorMessage: error.message,
            mitigation: { shouldRetry: false },
          };
        }
      }

      if (rpcCode === -32603) {
        return {
          errorCode: "CALL_EXCEPTION",
          errorMessage: error.message,
          mitigation: { shouldRetry: false },
        };
      }

      if (rpcCode === 4001) {
        return {
          errorCode: "ACTION_REJECTED",
          errorMessage: error.message,
          mitigation: { shouldRetry: false },
        };
      }

      return {
        errorCode: "RPC_ERROR",
        errorMessage: error.message,
        mitigation: { shouldRetry: true },
      };
    }

    if (error instanceof ContractFunctionRevertedError) {
      return {
        errorCode: "CALL_EXCEPTION",
        errorMessage: error.reason ?? error.shortMessage,
        data: error.data?.errorName,
        mitigation: { shouldRetry: false },
      };
    }

    if (error instanceof EstimateGasExecutionError) {
      return {
        errorCode: "CALL_EXCEPTION",
        errorMessage: error.shortMessage,
        mitigation: { shouldRetry: true },
      };
    }

    if (error instanceof TransactionNotFoundError) {
      return {
        errorCode: "UNKNOWN_ERROR",
        errorMessage: error.message,
        mitigation: { shouldRetry: true },
      };
    }

    if (error.shortMessage?.includes("nonce") || error.message?.includes("replacement underpriced")) {
      return {
        errorCode: "NONCE_EXPIRED",
        errorMessage: error.message,
        mitigation: { shouldRetry: true },
      };
    }

    return {
      errorCode: "UNKNOWN_ERROR",
      errorMessage: error.shortMessage ?? error.message,
      mitigation: { shouldRetry: true },
    };
  }
}
