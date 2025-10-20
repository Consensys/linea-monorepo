import { EthersError } from "ethers";
import { WinstonLogger } from "@consensys/linea-shared-utils";

export class PostmanWinstonLogger extends WinstonLogger {
  /**
   * Determines whether a given error should be logged as a `warning` instead of an `error`.
   *
   * This captures the original Postman-specific heuristics for common RPC responses.
   */
  protected shouldLogErrorAsWarning(error: unknown): boolean {
    const isEthersError = (value: unknown): value is EthersError => {
      return (value as EthersError).shortMessage !== undefined || (value as EthersError).code !== undefined;
    };

    if (!isEthersError(error)) {
      return false;
    }

    return (
      (error.shortMessage?.includes("processing response error") ||
        error.info?.error?.message?.includes("processing response error")) &&
      error.code === "SERVER_ERROR" &&
      error.info?.error?.code === -32603
    );
  }
}
