import { ResultAsync } from "neverthrow";

/**
 * @notice Wraps a potentially-throwing function (sync or async) into a `ResultAsync`.
 * @dev Useful for making code resilient to thrown errors, enabling Go-style error handling.
 *
 * Example:
 * const result = await tryResult(() => mightThrowAsync());
 * if (result.isErr()) return handle(result.error);
 * doSomething(result.value);
 *
 * @param fn A function that may throw or reject.
 * @returns A `ResultAsync<T, Error>` that resolves to `ok(value)` or `err(error)`.
 */
export const tryResult = <T>(fn: () => Promise<T> | T): ResultAsync<T, Error> => {
  return ResultAsync.fromPromise(Promise.resolve().then(fn), (e) => (e instanceof Error ? e : new Error(String(e))));
};
