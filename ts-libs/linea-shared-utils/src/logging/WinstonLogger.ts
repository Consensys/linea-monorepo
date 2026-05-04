/* eslint-disable @typescript-eslint/no-explicit-any */
import { Logger as LoggerClass, LoggerOptions, createLogger, format, transports } from "winston";

import { ILogger } from "./ILogger";
import { serialize } from "../utils/string";

// Logfmt: quote values that contain whitespace, double-quote, equals, or backslash.
const NEEDS_QUOTE_RE = /[ \t\n\r"=\\]/;
// Characters that must be escaped inside a quoted logfmt value.
const ESCAPE_RE = /["\\\n\r\t]/g;
const ESCAPE_MAP: Record<string, string> = {
  '"': '\\"',
  "\\": "\\\\",
  "\n": "\\n",
  "\r": "\\r",
  "\t": "\\t",
};

function formatValue(value: string): string {
  if (!NEEDS_QUOTE_RE.test(value)) return value;
  return `"${value.replace(ESCAPE_RE, (c) => ESCAPE_MAP[c])}"`;
}

export class WinstonLogger implements ILogger {
  private logger: LoggerClass;
  public readonly name: string;

  constructor(loggerName: string, options?: LoggerOptions) {
    const { combine, colorize, timestamp, printf, errors, label } = format;
    const colorizer = colorize();

    this.logger = createLogger({
      level: options?.level ?? "info",
      format: combine(
        timestamp(),
        errors({ stack: true }),
        label({ label: loggerName }),
        printf(({ timestamp, level, label, message, stack, ...metadata }) => {
          const coloredLevel = colorizer.colorize(level, level.toUpperCase());
          let str = `time=${timestamp} level=${coloredLevel} message=${formatValue(String(message))}`;

          const meta = this.formatMetadata(metadata);
          if (meta) str += ` ${meta}`;

          if (stack) str += ` error=${formatValue(String(stack))}`;

          str += ` | logger=${label}`;
          return str;
        }),
      ),
      transports: [new transports.Console()],
      ...options,
    });
    this.name = loggerName;
  }

  private formatMetadataValue(value: unknown): string {
    if (value == null) return "null";
    if (typeof value === "string") return formatValue(value);
    if (typeof value === "number" || typeof value === "boolean") return String(value);
    if (typeof value === "bigint") return value.toString();
    if (value instanceof Error) return formatValue(value.stack ?? value.message);
    return formatValue(serialize(value));
  }

  private formatMetadata(metadata: Record<string, unknown>): string {
    const keys = Object.keys(metadata);
    if (keys.length === 0) return "";
    const parts = new Array<string>(keys.length);
    for (let i = 0; i < keys.length; i++) {
      parts[i] = `${keys[i]}=${this.formatMetadataValue(metadata[keys[i]])}`;
    }
    return parts.join(" ");
  }

  private normalizeParams(params: any[]): any[] {
    if (params.length === 1 && params[0] instanceof Error) {
      return [{ error: params[0] }];
    }
    return params;
  }

  public info(message: any, ...params: any[]): void {
    this.logger.info(message, ...this.normalizeParams(params));
  }

  public error(message: any, ...params: any[]): void {
    this.logger.error(message, ...this.normalizeParams(params));
  }

  public warn(message: any, ...params: any[]): void {
    this.logger.warn(message, ...this.normalizeParams(params));
  }

  public debug(message: any, ...params: any[]): void {
    this.logger.debug(message, ...this.normalizeParams(params));
  }
}
