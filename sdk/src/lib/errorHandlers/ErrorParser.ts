import {
  getParsedEthersError,
  EthersError,
  RETURN_VALUE_ERROR_CODES,
  ReturnValue,
} from "@enzoferey/ethers-error-parser";
import { GasEstimationError } from "../utils/errors";
import { Message } from "../utils/types";

export type ParsableError = EthersError | Error;
export const CUSTOM_ERROR_CODES = {
  INSUFFICIENT_FUNDS: "INSUFFICIENT_FUNDS",
  NETWORK_ERROR: "NETWORK_ERROR",
  UNPREDICTABLE_GAS_LIMIT: "UNPREDICTABLE_GAS_LIMIT",
  // Add more errors to handle in the future
};

export const ParsedErrorCodes = {
  ...RETURN_VALUE_ERROR_CODES,
  ...CUSTOM_ERROR_CODES,
};

export type Mitigation = {
  shouldRetry: boolean;
  retryWithBlocking?: boolean;
  retryPeriodInMs?: number;
  retryNumOfTime?: number;
};

export type ParsedErrorResult = {
  errorCode: string;
  context?: string;
  reason?: string;
  mitigation: Mitigation;
};

export class ErrorParser {
  public static parseErrorWithMitigation(error: ParsableError): ParsedErrorResult | null {
    if (!error) {
      return null;
    }

    const mitigation: Mitigation = { shouldRetry: false };
    const parsedErrResult: ParsedErrorResult = {
      errorCode: ParsedErrorCodes.UNKNOWN_ERROR,
      mitigation,
    };

    const parsedEthersError = getParsedEthersError(error as EthersError);
    switch (parsedEthersError.errorCode) {
      case RETURN_VALUE_ERROR_CODES.INSUFFICIENT_FUNDS_FOR_GAS:
      case RETURN_VALUE_ERROR_CODES.CALL_REVERTED:
        parsedErrResult.mitigation = {
          shouldRetry: true,
          retryWithBlocking: true,
          retryPeriodInMs: 5000,
        };
        break;
      case RETURN_VALUE_ERROR_CODES.MAX_PRIORITY_FEE_PER_GAS_HIGHER_THAN_MAX_FEE_PER_GAS:
      case RETURN_VALUE_ERROR_CODES.MAX_FEE_PER_GAS_LESS_THAN_BLOCK_BASE_FEE:
      case RETURN_VALUE_ERROR_CODES.TRANSACTION_UNDERPRICED:
      case RETURN_VALUE_ERROR_CODES.NONCE_TOO_LOW:
        parsedErrResult.mitigation = {
          shouldRetry: true,
          retryWithBlocking: true,
        };
        break;
      case RETURN_VALUE_ERROR_CODES.EXECUTION_REVERTED:
      case RETURN_VALUE_ERROR_CODES.TRANSACTION_RAN_OUT_OF_GAS:
      case RETURN_VALUE_ERROR_CODES.REJECTED_TRANSACTION:
        break;
      default:
        if (parsedEthersError.errorCode === RETURN_VALUE_ERROR_CODES.UNKNOWN_ERROR) {
          this.mapToCustomErrorWithMitigation(error, parsedEthersError, parsedErrResult);
        }
        break;
    }

    return {
      ...parsedEthersError,
      ...parsedErrResult,
    };
  }

  private static mapToCustomErrorWithMitigation(
    error: ParsableError,
    parsedEthersError: ReturnValue,
    parsedErrResult: ParsedErrorResult,
  ) {
    if (
      parsedEthersError.context?.includes(ParsedErrorCodes.UNPREDICTABLE_GAS_LIMIT) ||
      (error as GasEstimationError<Message>).stack?.includes(ParsedErrorCodes.UNPREDICTABLE_GAS_LIMIT)
    ) {
      parsedErrResult.errorCode = ParsedErrorCodes.UNPREDICTABLE_GAS_LIMIT;
      parsedErrResult.mitigation = {
        shouldRetry: false,
      };
      return;
    }

    if ((error as EthersError).code === ParsedErrorCodes.INSUFFICIENT_FUNDS) {
      parsedErrResult.errorCode = ParsedErrorCodes.INSUFFICIENT_FUNDS;
      parsedErrResult.reason = (error as EthersError).reason;
    }

    if ((error as EthersError).code === ParsedErrorCodes.NETWORK_ERROR) {
      parsedErrResult.errorCode = ParsedErrorCodes.NETWORK_ERROR;
      parsedErrResult.reason = (error as EthersError).reason;
    }

    parsedErrResult.mitigation = {
      shouldRetry: true,
      retryWithBlocking: true,
      retryPeriodInMs: 5000,
    };
    return;
  }
}
