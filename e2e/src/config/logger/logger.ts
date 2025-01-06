import { Logger, LeveledLogMethod } from "winston";
import { baseLogger } from "./base-logger";

type LogLevel = "error" | "warn" | "info" | "verbose" | "debug" | "silly";

function getActiveTestName(): string | undefined {
  if (typeof expect !== "function" || typeof expect.getState !== "function") {
    return undefined;
  }

  const { currentTestName, currentConcurrentTestName } = expect.getState() || {};

  if (!currentConcurrentTestName) {
    return currentTestName;
  }

  return currentConcurrentTestName();
}

function appendTestName(level: LogLevel, args: unknown[]): Logger {
  const testName = getActiveTestName();

  if (!testName) {
    return baseLogger[level as keyof Logger](...args);
  }

  const [firstArg] = args;

  if (typeof firstArg === "string") {
    return baseLogger[level](firstArg, { testName });
  }

  if (firstArg && typeof firstArg === "object" && !Array.isArray(firstArg)) {
    return baseLogger[level]({ testName, ...firstArg });
  }

  return baseLogger[level as keyof Logger](...args);
}

export function createTestLogger(): Logger {
  return {
    ...baseLogger,
    error(...args: Parameters<LeveledLogMethod>) {
      return appendTestName("error", args);
    },
    warn(...args: Parameters<LeveledLogMethod>) {
      return appendTestName("warn", args);
    },
    info(...args: Parameters<LeveledLogMethod>) {
      return appendTestName("info", args);
    },
    verbose(...args: Parameters<LeveledLogMethod>) {
      return appendTestName("verbose", args);
    },
    debug(...args: Parameters<LeveledLogMethod>) {
      return appendTestName("debug", args);
    },
    silly(...args: Parameters<LeveledLogMethod>) {
      return appendTestName("silly", args);
    },
  } as Logger;
}
