import { wait } from "./time";

export async function awaitUntil<T>(
  callback: () => Promise<T>,
  stopRetry: (a: T) => boolean,
  pollingIntervalMs: number = 500,
  timeoutMs: number = 2 * 60 * 1000,
): Promise<T | null> {
  let isExceedTimeOut = false;
  setTimeout(() => {
    isExceedTimeOut = true;
  }, timeoutMs);

  while (!isExceedTimeOut) {
    const result = await callback();
    if (stopRetry(result)) return result;
    await wait(pollingIntervalMs);
  }
  return null;
}
