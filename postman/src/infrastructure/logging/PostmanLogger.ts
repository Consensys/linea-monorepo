/* eslint-disable @typescript-eslint/no-explicit-any */
import { WinstonLogger } from "@consensys/linea-shared-utils";
import { BaseError as ViemBaseError, HttpRequestError, RpcRequestError } from "viem";

import type { IPostmanLogger } from "../../domain/ports/ILogger";

export class PostmanLogger extends WinstonLogger implements IPostmanLogger {
  public warnOrError(message: any, ...params: any[]): void {
    if (this.shouldLogErrorAsWarning(message)) {
      this.warn(message, ...params);
    } else {
      this.error(message, ...params);
    }
  }

  protected shouldLogErrorAsWarning(error: unknown): boolean {
    if (!(error instanceof ViemBaseError)) {
      return false;
    }

    if (error instanceof HttpRequestError) {
      return error.status === 500 || error.status === 502 || error.status === 503;
    }

    if (error instanceof RpcRequestError) {
      const rpcError = error as RpcRequestError;
      return rpcError.code === -32603;
    }

    return false;
  }
}
