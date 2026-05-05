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
const URL_REPLACEMENT = "[REDACTED_URL]";

// `inspect` depth bounds traversal of pathological structures (nested causes,
// deeply nested objects). Cycle detection is handled internally by `inspect`.
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
  "secretkey",
  "secret_key",
  "clientsecret",
  "client_secret",
  "password",
  "passphrase",
  "keystorepassword",
  "keystore_password",
  "truststorepassword",
  "truststore_password",
  "apikey",
  "api_key",
  "mnemonic",
  "secret",
  "accesstoken",
  "access_token",
  "bearertoken",
  "bearer_token",
  "refreshtoken",
  "refresh_token",
  "idtoken",
  "id_token",
  "authorization",
  "seed",
];

/**
 * Matches URL substrings inside free-form text. Used to scrub URLs from
 * Error messages and stacks (e.g. viem's
 * `RpcRequestError: RPC Request failed.\n\n  URL: https://rpc.example/path`),
 * where per-key redaction can't reach.
 *
 * Limited to schemes actually used in this monorepo so it does not match
 * `file:` paths or arbitrary `foo://bar` identifiers. Match stops at
 * whitespace/quotes/angle-brackets so the surrounding text is preserved
 * verbatim.
 */
const URL_PATTERN_RE = /\b(?:https?|wss?|postgres(?:ql)?|mysql|mongodb|redis):\/\/[^\s"'<>`\\]+/gi;

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

/**
 * Replaces every URL substring in `text` with `[REDACTED_URL]`. Used on
 * Error `message` and `stack` strings, so URLs embedded by HTTP/RPC libraries
 * (viem, ethers, etc.) — which often carry userinfo or `?apiKey=…` — never
 * reach the log output. `String#replace` with a global regex returns the
 * original string when there are no matches, so no fast-path needed.
 */
function stripUrlsFromText(text: string): string {
  return text.replace(URL_PATTERN_RE, URL_REPLACEMENT);
}

export class WinstonLogger implements ILogger {
  private logger: LoggerClass;
  public readonly name: string;
  private readonly redactKeys: ReadonlySet<string>;

  constructor(loggerName: string, options?: WinstonLoggerOptions) {
    const { combine, colorize, timestamp, printf, errors, label, splat } = format;
    const colorizer = colorize();

    const { redact, ...winstonOptions } = options ?? {};
    const keys = new Set<string>(DEFAULT_REDACT_KEYS);
    if (redact) for (const k of redact) keys.add(k.toLowerCase());
    this.redactKeys = keys;

    // Spread caller options FIRST, then override with our defaults / required
    // fields. `format` is non-overridable: it owns the redaction and URL-strip
    // pipeline, so a caller-supplied format must never replace it (otherwise
    // every secret leaks). `transports` IS overridable so tests can swap in a
    // capturing stream — transports receive already-formatted strings and
    // therefore cannot bypass redaction. `level` falls back to "info" only if
    // the caller didn't set one.
    this.logger = createLogger({
      ...winstonOptions,
      level: winstonOptions.level ?? "info",
      transports: winstonOptions.transports ?? [new transports.Console()],
      format: combine(
        timestamp(),
        // For bare-Error calls (`logger.error(err)`):
        //  - `stack: true` promotes `err.stack` so printf can emit `error=…`.
        //  - `cause: true` promotes `err.cause` so it surfaces as a separate
        //    `cause=…` metadata field. `cause` is non-enumerable when set
        //    via `new Error(msg, { cause })`, so without this flag winston's
        //    `Object.assign` would silently drop it.
        errors({ stack: true, cause: true }),
        splat(),
        label({ label: loggerName }),
        printf(({ timestamp, level, label, message, stack, ...metadata }) => {
          const coloredLevel = colorizer.colorize(level, level.toUpperCase());
          // `message` is free-form text. Strip URLs so `RpcRequestError`-style
          // errors (which embed the request URL into `error.message`, and
          // therefore into the first line of `error.stack`) cannot leak them.
          const safeMessage = stripUrlsFromText(String(message));
          let str = `time=${timestamp} level=${coloredLevel} logger=${label} msg=${formatValue(safeMessage)}`;

          const meta = this.formatMetadata(metadata);
          if (meta) str += ` ${meta}`;

          if (stack) {
            const safeStack = stripUrlsFromText(String(stack));
            str += ` error=${formatValue(safeStack)}`;
          }

          return str;
        }),
      ),
    });
    this.name = loggerName;
  }

  private isRedacted(key: string): boolean {
    return this.redactKeys.has(key.toLowerCase());
  }

  /**
   * Walks a plain object/array and returns a structurally equivalent copy
   * with values under any redacted key replaced with `[REDACTED]`. Bigints
   * are coerced to strings (so `JSON.stringify` doesn't throw); cycles are
   * broken with `"[Circular]"`.
   *
   * Scope is intentionally narrow: per-key redaction for caller-supplied
   * config dumps. URLs embedded inside string values are NOT scrubbed here —
   * those are caught at the Error-message/stack layer in the printf.
   */
  private redactValue(value: unknown, key: string, seen: WeakSet<object>): unknown {
    if (key && this.isRedacted(key)) return REDACTED_PLACEHOLDER;
    if (typeof value === "bigint") return value.toString();
    if (value === null || typeof value !== "object") return value;
    if (seen.has(value)) return "[Circular]";
    seen.add(value);

    if (Array.isArray(value)) {
      return value.map((item, i) => this.redactValue(item, String(i), seen));
    }
    const out: Record<string, unknown> = {};
    for (const propKey of Object.keys(value)) {
      out[propKey] = this.redactValue((value as Record<string, unknown>)[propKey], propKey, seen);
    }
    return out;
  }

  private formatMetadataValue(value: unknown): string {
    if (value == null) return "null";
    if (typeof value === "string") return formatValue(value);
    if (typeof value === "number" || typeof value === "boolean") return String(value);
    if (typeof value === "bigint") return value.toString();
    if (value instanceof Error) {
      // Render the Error via util.inspect (preserves `Error: <msg>\n at...`
      // formatting, cause chains, AggregateError siblings) but strip URLs
      // from the rendered text afterwards. Error own-properties (e.g.
      // `error.code`) are dumped by inspect verbatim — they are not redacted
      // here. If a caller attaches a secret to an Error, log it under a
      // metadata key instead of as the Error itself.
      return formatValue(stripUrlsFromText(inspect(value, ERROR_INSPECT_OPTIONS)));
    }
    return formatValue(JSON.stringify(this.redactValue(value, "", new WeakSet<object>())));
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
    // Skip the constructor (which would build a fresh winston pipeline) and
    // wrap the parent's already-configured `winston.Logger.child()` directly.
    // Plain assignment matches the constructor's class-field semantics
    // (enumerable + writable); TypeScript's `readonly` is compile-time only.
    return Object.assign(Object.create(WinstonLogger.prototype) as WinstonLogger, {
      name: this.name,
      logger: this.logger.child(context),
      redactKeys: this.redactKeys,
    });
  }
}
