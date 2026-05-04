/* eslint-disable @typescript-eslint/no-explicit-any */
import { inspect } from "node:util";
import { Logger as LoggerClass, LoggerOptions, createLogger, format, transports } from "winston";

import { ILogger } from "./ILogger";

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

const REDACTED_PLACEHOLDER = "[REDACTED]";

// Use Node's `util.inspect` to render Errors. It already handles `Error.cause`
// chains, `AggregateError.errors`, cycles (via `<ref *N>` markers), and any
// custom enumerable properties — so we don't reimplement any of that.
//
// `depth: 5` bounds traversal of pathological structures (e.g. very deep
// cause chains from libraries that re-wrap repeatedly). Real-world cause
// chains are 2-3 deep and AggregateError siblings render at depth 2, so this
// limit is well above what callers actually produce. Cycle detection works
// independently of depth.
const ERROR_INSPECT_OPTIONS = {
  depth: 5,
  colors: false,
  breakLength: Infinity,
} as const;

/**
 * Default keys whose values are masked in log output. Comparison is
 * case-insensitive and exact (no substring matching, to avoid e.g. "passport"
 * matching "pass"). Keys are stored lower-cased.
 */
const DEFAULT_REDACT_KEYS: ReadonlyArray<string> = [
  "privatekey",
  "private_key",
  "signerkey",
  "signer_key",
  "password",
  "passphrase",
  "apikey",
  "api_key",
  "mnemonic",
  "secret",
  "token",
  "authorization",
  "auth",
  "seed",
];

export type WinstonLoggerOptions = LoggerOptions & {
  /**
   * Additional keys (case-insensitive) whose values must be replaced with
   * `[REDACTED]` in log metadata. Merged with the built-in default list.
   */
  redact?: string[];
};

function formatValue(value: string): string {
  if (!NEEDS_QUOTE_RE.test(value)) return value;
  return `"${value.replace(ESCAPE_RE, (c) => ESCAPE_MAP[c])}"`;
}

export class WinstonLogger implements ILogger {
  private logger: LoggerClass;
  public readonly name: string;
  private readonly redactKeys: ReadonlySet<string>;

  constructor(loggerName: string, options?: WinstonLoggerOptions) {
    const { combine, colorize, timestamp, printf, errors, label } = format;
    const colorizer = colorize();

    const { redact, ...winstonOptions } = options ?? {};
    this.redactKeys = WinstonLogger.buildRedactKeys(redact);

    this.logger = createLogger({
      level: winstonOptions.level ?? "info",
      format: combine(
        timestamp(),
        // For bare-Error calls (`logger.error(err)`):
        //  - `stack: true` promotes `err.stack` so printf can emit `error=…`.
        //  - `cause: true` promotes `err.cause` so it surfaces as a separate
        //    `cause=…` metadata field. `cause` is non-enumerable when set
        //    via `new Error(msg, { cause })`, so without this flag winston's
        //    `Object.assign` would silently drop it. An Error cause is then
        //    rendered through `util.inspect` in `formatMetadataValue`, which
        //    handles its own cause chain, `AggregateError`, and cycles.
        errors({ stack: true, cause: true }),
        label({ label: loggerName }),
        printf(({ timestamp, level, label, message, stack, ...metadata }) => {
          const coloredLevel = colorizer.colorize(level, level.toUpperCase());
          let str = `time=${timestamp} level=${coloredLevel} logger=${label} msg=${formatValue(String(message))}`;

          const meta = this.formatMetadata(metadata);
          if (meta) str += ` ${meta}`;

          if (stack) str += ` error=${formatValue(String(stack))}`;

          return str;
        }),
      ),
      transports: [new transports.Console()],
      ...winstonOptions,
    });
    this.name = loggerName;
  }

  private static buildRedactKeys(extra?: string[]): ReadonlySet<string> {
    const set = new Set<string>(DEFAULT_REDACT_KEYS);
    if (extra) {
      for (const key of extra) set.add(key.toLowerCase());
    }
    return set;
  }

  private isRedacted(key: string): boolean {
    return this.redactKeys.has(key.toLowerCase());
  }

  /**
   * JSON-serializes a metadata value while:
   *  - masking any nested property whose key matches the redaction set
   *    (case-insensitive),
   *  - coercing bigints to strings so they survive `JSON.stringify`,
   *  - replacing already-visited references with `"[Circular]"` so that
   *    cyclic structures (e.g. viem/ethers provider objects) don't crash
   *    the logger via `Converting circular structure to JSON`.
   *
   * Note: the visited-set tracks every object, so a non-cyclic graph that
   * shares a sub-object via two paths will render the second path as
   * `"[Circular]"`. That's an acceptable trade-off for log output: the call
   * site still completes without error and no data is lost from the first
   * encounter.
   */
  private serializeWithRedaction(value: unknown): string {
    const seen = new WeakSet<object>();
    return JSON.stringify(value, (key: string, val: unknown) => {
      if (key && this.isRedacted(key)) return REDACTED_PLACEHOLDER;
      if (typeof val === "bigint") return val.toString();
      if (val !== null && typeof val === "object") {
        if (seen.has(val as object)) return "[Circular]";
        seen.add(val as object);
      }
      return val;
    });
  }

  private formatMetadataValue(value: unknown): string {
    if (value == null) return "null";
    if (typeof value === "string") return formatValue(value);
    if (typeof value === "number" || typeof value === "boolean") return String(value);
    if (typeof value === "bigint") return value.toString();
    if (value instanceof Error) return formatValue(inspect(value, ERROR_INSPECT_OPTIONS));
    return formatValue(this.serializeWithRedaction(value));
  }

  private formatMetadata(metadata: Record<string, unknown>): string {
    const keys = Object.keys(metadata);
    if (keys.length === 0) return "";
    const parts = new Array<string>(keys.length);
    for (let i = 0; i < keys.length; i++) {
      const key = keys[i];
      parts[i] = `${key}=${this.isRedacted(key) ? REDACTED_PLACEHOLDER : this.formatMetadataValue(metadata[key])}`;
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

  public child(context: Record<string, unknown>): WinstonLogger {
    return WinstonLogger.fromInternal(this.name, this.logger.child(context), this.redactKeys);
  }

  private static fromInternal(name: string, internal: LoggerClass, redactKeys: ReadonlySet<string>): WinstonLogger {
    const instance = Object.create(WinstonLogger.prototype) as WinstonLogger;
    Object.defineProperty(instance, "name", { value: name, enumerable: true });
    Object.defineProperty(instance, "logger", { value: internal, writable: true, enumerable: true });
    Object.defineProperty(instance, "redactKeys", { value: redactKeys, enumerable: false });
    return instance;
  }
}
