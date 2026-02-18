/* eslint-disable @typescript-eslint/no-explicit-any */

export interface ILogger {
  readonly name: string;
  info(message: any, ...params: any[]): void;
  error(message: any, ...params: any[]): void;
  warn(message: any, ...params: any[]): void;
  debug(message: any, ...params: any[]): void;
}

/**
 * @deprecated Use ILogger directly. Kept for backward compatibility with legacy code.
 */
export type IPostmanLogger = ILogger;
