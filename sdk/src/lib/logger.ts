/* eslint-disable @typescript-eslint/no-explicit-any */
import { Logger, LoggerOptions, createLogger } from "winston";

export class LineaLogger {
  private logger: Logger;

  constructor(loggerName: string, options?: LoggerOptions) {
    this.logger = createLogger({
      ...options,
      defaultMeta: { module: loggerName },
    });
  }

  public info(message: any): void;
  public info(message: string, ...meta: any[]): void {
    if (typeof message === "string") {
      this.logger.info(message, meta);
    } else {
      this.logger.info(message);
    }
  }

  public error(message: any): void;
  public error(message: string, ...meta: any[]): void {
    if (typeof message === "string") {
      this.logger.error(message, meta);
    } else {
      this.logger.error(message);
    }
  }

  public warn(message: any): void;
  public warn(message: string, ...meta: any[]): void {
    if (typeof message === "string") {
      this.logger.warn(message, meta);
    } else {
      this.logger.warn(message);
    }
  }
}

export const getLogger = (loggerName: string, options?: LoggerOptions) => {
  return new LineaLogger(loggerName, options);
};
