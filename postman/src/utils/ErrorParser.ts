import { MessageProps } from "../core/entities/Message";
import { DatabaseAccessError } from "../core/errors/DatabaseErrors";
import { IErrorParser, ParsedError } from "../core/errors/IErrorParser";

export type Mitigation = {
  shouldRetry: boolean;
};

export type ParsedErrorResult = {
  errorCode: string;
  errorMessage?: string;
  data?: string;
  mitigation: Mitigation;
};

export class ErrorParser implements IErrorParser {
  public parse(error: unknown): ParsedError {
    const result = ErrorParser.parseErrorWithMitigation(error);
    return {
      retryable: result?.mitigation.shouldRetry ?? false,
      message: result?.errorMessage ?? (error instanceof Error ? error.message : String(error)),
    };
  }

  public static parseErrorWithMitigation(error: unknown): ParsedErrorResult | null {
    if (!error) {
      return null;
    }

    if (error instanceof DatabaseAccessError) {
      return {
        errorCode: "UNKNOWN_ERROR",
        errorMessage: (error as DatabaseAccessError<MessageProps>).message,
        mitigation: { shouldRetry: true },
      };
    }

    return {
      errorCode: "UNKNOWN_ERROR",
      errorMessage: error instanceof Error ? error.message : String(error),
      mitigation: { shouldRetry: false },
    };
  }
}
