import { wait } from "./time";
import { createTestLogger } from "../../config/logger/logger";

const logger = createTestLogger();

export class AwaitUntilTimeoutError extends Error {
  constructor(
    public readonly timeoutMs: number,
    public readonly lastError?: Error,
  ) {
    const message = lastError
      ? `awaitUntil timed out after ${timeoutMs}ms. Last error: ${lastError.message}`
      : `awaitUntil timed out after ${timeoutMs}ms`;
    super(message);
    this.name = "AwaitUntilTimeoutError";
  }
}

export async function awaitUntil<T>(
  callback: () => Promise<T>,
  stopRetry: (value: T) => boolean,
  pollingIntervalMs = 500,
  timeoutMs = 2 * 60 * 1000,
): Promise<T> {
  const deadline = Date.now() + timeoutMs;
  let lastError: Error | undefined;
  let attemptCount = 0;

  while (Date.now() < deadline) {
    try {
      const result = await callback();
      lastError = undefined;

      if (stopRetry(result)) {
        return result;
      }
    } catch (error) {
      lastError = error as Error;
      attemptCount++;
      // Log error on first occurrence and then every 10 attempts to avoid spam
      if (attemptCount === 1 || attemptCount % 10 === 0) {
        logger.warn(
          `awaitUntil callback error (attempt ${attemptCount}). Will retry until timeout. error=${lastError.message}`,
        );
      }
    }

    await wait(pollingIntervalMs);
  }

  throw new AwaitUntilTimeoutError(timeoutMs, lastError);
}
