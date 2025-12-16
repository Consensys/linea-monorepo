import { wait } from "./time";

export class AwaitUntilTimeoutError extends Error {
  constructor(public readonly timeoutMs: number) {
    super(`awaitUntil timed out after ${timeoutMs}ms`);
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

  while (Date.now() < deadline) {
    const result = await callback();

    if (stopRetry(result)) {
      return result;
    }

    await wait(pollingIntervalMs);
  }

  throw new AwaitUntilTimeoutError(timeoutMs);
}
