export interface IRetryService<TReturn> {
  /**
   * @notice Retry an asynchronous operation until success or failure.
   * @param fn An async callback returning a promise of type TReturn.
   */
  retry(fn: () => Promise<TReturn>): Promise<TReturn>;
}
