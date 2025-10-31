/* eslint-disable @typescript-eslint/no-explicit-any */
import { Logger as LoggerClass, LoggerOptions, createLogger, format, transports } from "winston";
import { ILogger } from "./ILogger";
import { isString, serialize } from "../utils/string";

export class WinstonLogger implements ILogger {
  private logger: LoggerClass;
  public readonly name: string;

  /**
   * Initializes a new instance of the `WinstonLogger`.
   *
   * @param {string} loggerName - The name of the logger, typically representing the class or module using this logger instance.
   * @param {LoggerOptions} [options] - Optional configuration settings for the logger, allowing customization of the logging behavior and output.
   */
  constructor(loggerName: string, options?: LoggerOptions) {
    const { align, combine, colorize, timestamp, printf, errors, splat, label } = format;

    this.logger = createLogger({
      level: options?.level ?? "info",
      format: combine(
        timestamp(),
        errors({ stack: true }),
        splat(),
        label({ label: loggerName }),
        printf(({ timestamp, level, label, message, stack, ...metadata }) => {
          let str = `time=${timestamp} level=${level.toUpperCase()} message=${message}`;

          str += ` | class=${label} ${this.formatMetadata(metadata)}`;

          if (stack) {
            str += ` | ${stack}`;
          }

          return colorize().colorize(level, str);
        }),
        align(),
      ),
      transports: [new transports.Console()],
      ...options,
    });
    this.name = loggerName;
  }

  /**
   * Formats metadata for logging, ensuring that non-string values are serialized for readability.
   *
   * @param {any} metadata - The metadata object to format.
   * @returns {string} A string representation of the metadata, suitable for inclusion in a log message.
   */
  private formatMetadata(metadata: any): string {
    if (isString(metadata)) {
      return `metadata=${metadata}`;
    }

    let str = "";

    for (const key of Object.keys(metadata)) {
      if (isString(metadata[key])) {
        str += ` ${key}=${metadata[key]}`;
        continue;
      }
      str += ` ${key}=${serialize(metadata[key])}`;
    }
    return str.trim();
  }

  /**
   * Logs a message at the `info` level.
   *
   * @param {any} message - The primary log message.
   * @param {...any[]} params - Additional parameters or metadata to log alongside the message.
   */
  public info(message: any, ...params: any[]): void {
    this.logger.info(message, ...params);
  }

  /**
   * Logs a message at the `error` level.
   *
   * @param {any} message - The primary log message.
   * @param {...any[]} params - Additional parameters or metadata to log alongside the message.
   */
  public error(message: any, ...params: any[]): void {
    this.logger.error(message, ...params);
  }

  /**
   * Logs a message at the `warn` level.
   *
   * @param {any} message - The primary log message.
   * @param {...any[]} params - Additional parameters or metadata to log alongside the message.
   */
  public warn(message: any, ...params: any[]): void {
    this.logger.warn(message, ...params);
  }

  /**
   * Logs a message at the `debug` level.
   *
   * @param {any} message - The primary log message.
   * @param {...any[]} params - Additional parameters or metadata to log alongside the message.
   */
  public debug(message: any, ...params: any[]): void {
    this.logger.debug(message, ...params);
  }

  /**
   * Decides whether to log a message as a `warning` or an `error` based on its content and severity.
   *
   * This method is particularly useful for handling errors that may not always require immediate attention or could be retried successfully.
   *
   * @param {any} message - The primary log message or error object.
   * @param {...any[]} params - Additional parameters or metadata to log alongside the message.
   */
  public warnOrError(message: any, ...params: any[]): void {
    if (this.shouldLogErrorAsWarning(message)) {
      this.warn(message, ...params);
    } else {
      this.error(message, ...params);
    }
  }

  /**
   * Hook that determines whether a given payload should be logged as a warning instead of an error.
   *
   * Packages extending the base logger can override this method to provide domain-specific logic.
   *
   * @param {unknown} _error - The value provided to `warnOrError`.
   * @returns {boolean} `true` to downgrade the entry to `warn`, `false` otherwise.
   */
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  protected shouldLogErrorAsWarning(_error: unknown): boolean {
    return false;
  }
}
