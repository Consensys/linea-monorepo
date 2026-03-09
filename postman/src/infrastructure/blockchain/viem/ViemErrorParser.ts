import {
  BaseError,
  ContractFunctionRevertedError,
  HttpRequestError,
  TimeoutError,
  TransactionRejectedRpcError,
} from "viem";

import { DatabaseAccessError } from "../../../core/errors/DatabaseErrors";
import { IErrorParser, ParsedError } from "../../../core/errors/IErrorParser";

export class ViemErrorParser implements IErrorParser {
  public parse(error: unknown): ParsedError {
    if (error instanceof DatabaseAccessError) {
      return { retryable: true, message: error.message };
    }

    if (!(error instanceof BaseError)) {
      return {
        retryable: false,
        message: error instanceof Error ? error.message : String(error),
      };
    }

    return {
      retryable: this.isRetryable(error),
      message: error.shortMessage ?? error.message,
    };
  }

  private isRetryable(error: BaseError): boolean {
    // User rejection — not retryable
    if (error.walk((e) => e instanceof TransactionRejectedRpcError)) {
      return false;
    }

    // Contract revert — not retryable
    if (error.walk((e) => e instanceof ContractFunctionRevertedError)) {
      return false;
    }

    // HTTP / network errors — retryable
    if (error.walk((e) => e instanceof HttpRequestError)) {
      return true;
    }

    // Timeout — retryable
    if (error.walk((e) => e instanceof TimeoutError)) {
      return true;
    }

    // Walk for RPC error codes embedded in the chain
    const rpcCause = error.walk((e) => {
      const code = (e as { code?: number }).code;
      return code !== undefined && code !== null;
    });

    if (rpcCause) {
      const code = (rpcCause as { code?: number }).code;
      const message: string = (rpcCause as { message?: string }).message ?? "";

      if (code === -32603) {
        // Internal JSON-RPC error — retryable
        return true;
      }

      if (code === 4001) {
        // User rejection (EIP-1193) — not retryable
        return false;
      }

      if (code === -32000) {
        if (message.includes("execution reverted")) return false;
        if (
          message.includes("gas required exceeds allowance") ||
          message.includes("max priority fee per gas higher than max fee per gas") ||
          message.includes("max fee per gas less than block base fee") ||
          /below sender account nonce/.test(message)
        ) {
          return true;
        }
      }
    }

    // Default for unknown viem errors — not retryable
    return false;
  }
}
