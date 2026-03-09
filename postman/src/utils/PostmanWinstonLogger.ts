/* eslint-disable @typescript-eslint/no-explicit-any */
import { WinstonLogger } from "@consensys/linea-shared-utils";

import { IPostmanLogger } from "./IPostmanLogger";
import { IErrorParser } from "../core/errors/IErrorParser";

export class PostmanWinstonLogger extends WinstonLogger implements IPostmanLogger {
  constructor(
    name: string,

    options?: any,
    private readonly errorParser?: IErrorParser,
  ) {
    super(name, options);
  }

  /**
   * Decides whether to log a message as a `warning` or an `error` based on its content and severity.
   *
   * @param {any} message - The primary log message or error object.
   * @param {...any[]} params - Additional parameters or metadata to log alongside the message.
   */
  public warnOrError(message: any, ...params: any[]): void {
    if (this.errorParser && this.errorParser.parse(message).retryable) {
      this.warn(message, ...params);
    } else {
      this.error(message, ...params);
    }
  }
}
