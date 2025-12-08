import { BaseError } from "viem";

/**
 * Extracted error information from a viem BaseError, including revert reasons and metadata.
 */
export interface ExtractedViemError {
  name: string;
  message: string;
  shortMessage?: string;
  details?: string;
  code?: number;
  metaMessages?: string[];
  revertData?: unknown;
  revertReason?: string;
  errorName?: string;
  cause?: ExtractedViemError;
}

/**
 * Extracts useful error information from a viem BaseError instance.
 * Traverses the error chain to find revert reasons, metaMessages, and other diagnostic information.
 *
 * @param {unknown} error - The error to extract information from (may be a viem BaseError or any other error).
 * @returns {ExtractedViemError | unknown} - Extracted error information if it's a viem error, otherwise returns the original error.
 */
export function extractViemErrorInfo(error: unknown): ExtractedViemError | unknown {
  // If not a viem BaseError, return as-is
  if (!(error instanceof BaseError)) {
    return error;
  }

  const extracted: ExtractedViemError = {
    name: error.name,
    message: error.message,
  };

  // Extract shortMessage and details if available
  if ("shortMessage" in error) {
    extracted.shortMessage = (error as { shortMessage?: string }).shortMessage;
  }
  if ("details" in error) {
    extracted.details = (error as { details?: string }).details;
  }

  // Extract error code
  const errorWithCode = error.walk((err) => typeof (err as { code?: unknown }).code === "number");
  if (errorWithCode) {
    extracted.code = (errorWithCode as unknown as { code: number }).code;
  }

  // Extract metaMessages if available (often contains useful context like estimate gas arguments)
  if ("metaMessages" in error && Array.isArray((error as { metaMessages?: unknown }).metaMessages)) {
    extracted.metaMessages = (error as { metaMessages: string[] }).metaMessages;
  }

  // Try to find EstimateGasExecutionError first to get its metaMessages
  const estimateGasError = error.walk((err) => {
    return (
      err instanceof Error &&
      (err.name === "EstimateGasExecutionError" || ((err as { name?: string }).name?.includes("EstimateGas") ?? false))
    );
  });

  if (estimateGasError) {
    // Extract metaMessages from EstimateGasExecutionError if available
    if (
      "metaMessages" in estimateGasError &&
      Array.isArray((estimateGasError as { metaMessages?: unknown }).metaMessages)
    ) {
      const estimateGasMetaMessages = (estimateGasError as { metaMessages: string[] }).metaMessages;
      if (estimateGasMetaMessages && estimateGasMetaMessages.length > 0) {
        // Merge metaMessages if we don't already have them
        if (!extracted.metaMessages) {
          extracted.metaMessages = [];
        }
        // Avoid duplicates
        estimateGasMetaMessages.forEach((msg) => {
          if (!extracted.metaMessages!.includes(msg)) {
            extracted.metaMessages!.push(msg);
          }
        });
      }
    }
  }

  // Try to find ExecutionRevertedError in the error chain to extract revert data
  const executionRevertedError = error.walk((err) => {
    // Check if this error has the characteristics of ExecutionRevertedError
    return (
      err instanceof Error &&
      (err.name === "ExecutionRevertedError" ||
        err.name === "ContractFunctionExecutionError" ||
        ((err as { name?: string }).name?.includes("Reverted") ?? false))
    );
  });

  if (executionRevertedError) {
    // Extract revert data if available
    if ("data" in executionRevertedError) {
      const data = (executionRevertedError as { data?: unknown }).data;
      extracted.revertData = data;

      // Try to extract errorName or reason from the data
      if (data && typeof data === "object") {
        const dataObj = data as Record<string, unknown>;
        if ("errorName" in dataObj) {
          extracted.errorName = String(dataObj.errorName);
        }
        if ("reason" in dataObj) {
          extracted.revertReason = String(dataObj.reason);
        }
        // Sometimes the data itself is the reason
        if ("args" in dataObj && Array.isArray(dataObj.args) && dataObj.args.length > 0) {
          extracted.revertReason = String(dataObj.args[0]);
        }
      }
    }

    // Extract reason if directly available on the error
    if ("reason" in executionRevertedError) {
      extracted.revertReason = String((executionRevertedError as { reason?: unknown }).reason);
    }
  }

  // Extract cause if it exists (may be a BaseError or plain Error)
  if ("cause" in error && error.cause) {
    if (error.cause instanceof BaseError) {
      extracted.cause = extractViemErrorInfo(error.cause) as ExtractedViemError;
    } else if (error.cause instanceof Error) {
      // Extract basic info from non-BaseError causes
      extracted.cause = {
        name: error.cause.name,
        message: error.cause.message,
      };
      if ("details" in error.cause) {
        extracted.cause.details = String((error.cause as { details?: unknown }).details);
      }
      if ("shortMessage" in error.cause) {
        extracted.cause.shortMessage = String((error.cause as { shortMessage?: unknown }).shortMessage);
      }
    }
  }

  return extracted;
}
