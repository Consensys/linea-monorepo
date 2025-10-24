import { IRetryService } from "../core/services/IRetryService";
import { ILogger } from "../logging/ILogger";
import { wait } from "@consensys/linea-sdk";

export class ExponentialBackoffRetryService<TReturn> implements IRetryService<TReturn> {
  constructor(
    private readonly logger: ILogger,
    private readonly retryAttempts: number = 3,
    private readonly baseDelayMs: number = 1000,
  ) {
    if (retryAttempts < 1) {
      throw new Error("retryAttempts must be at least 1");
    }
    if (baseDelayMs < 0) {
      throw new Error("baseDelay must be non-negative");
    }
  }

  /**
   * @notice Retry an asynchronous operation until success or failure.
   * @param fn An async callback returning a promise of type TReturn.
   */
  public async retry(fn: () => Promise<TReturn>): Promise<TReturn> {
    let lastError: unknown;

    for (let attempt = 0; attempt <= this.retryAttempts; attempt += 1) {
      try {
        return await fn();
      } catch (error) {
        lastError = error;
        this.logger.warn("Retry attempt failed", { attempt, retryAttempts: this.retryAttempts, error });

        if (attempt >= this.retryAttempts) {
          this.logger.error("Retry attempts exhausted", { retryAttempts: this.retryAttempts, error });
          throw error;
        }

        const delayMs = this.getDelayMs(attempt);
        this.logger.debug("Retrying after delay", { attempt, delayMs });
        await wait(delayMs);
      }
    }

    throw lastError;
  }

  private getDelayMs(attempt: number): number {
    const exponentialDelay = this.baseDelayMs * 2 ** (attempt - 1);
    const jitter = Math.random() * exponentialDelay;
    return exponentialDelay + jitter;
  }
}
