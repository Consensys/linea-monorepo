import { IRetryService } from "../core/services/IRetryService";
import { ILogger } from "../logging/ILogger";
import { wait } from "@consensys/linea-sdk";

export class ExponentialBackoffRetryService implements IRetryService {
  constructor(
    private readonly logger: ILogger,
    private readonly maxRetryAttempts: number = 3,
    private readonly baseDelayMs: number = 1000,
    private readonly defaultTimeoutMs: number = 10000,
  ) {
    if (maxRetryAttempts < 1) {
      throw new Error("maxRetryAttempts must be at least 1");
    }
    if (baseDelayMs < 0) {
      throw new Error("baseDelay must be non-negative");
    }
  }

  /**
   * @notice Retry an asynchronous operation until success or failure.
   * @param fn An async callback returning a promise of type TReturn.
   */
  public async retry<TReturn>(fn: () => Promise<TReturn>, timeoutMs?: number): Promise<TReturn> {
    let lastError: unknown;

    for (let attempt = 1; attempt <= this.maxRetryAttempts; attempt += 1) {
      try {
        return await this.executeWithTimeout(fn, timeoutMs ? timeoutMs : this.defaultTimeoutMs);
      } catch (error) {
        if (attempt >= this.maxRetryAttempts) {
          this.logger.error(`Retry attempts exhausted maxRetryAttempts=${this.maxRetryAttempts}`, { error });
          throw error;
        }

        lastError = error;
        this.logger.warn(`Retry attempt failed attempt=${attempt} maxRetryAttempts=${this.maxRetryAttempts}`, {
          error,
        });

        const delayMs = this.getDelayMs(attempt);
        this.logger.debug(`Retrying after delay=${delayMs}ms`);
        await wait(delayMs);
      }
    }

    throw lastError;
  }

  private executeWithTimeout<TReturn>(fn: () => Promise<TReturn>, timeoutMs: number): Promise<TReturn> {
    if (timeoutMs <= 0) {
      throw new Error("timeoutMs must be greater than 0");
    }

    return new Promise<TReturn>((resolve, reject) => {
      const timeoutHandle = setTimeout(() => {
        reject(new Error(`${fn.name} timed out after ${timeoutMs}ms`));
      }, timeoutMs);

      fn()
        .then(resolve)
        .catch(reject)
        .finally(() => clearTimeout(timeoutHandle));
    });
  }

  private getDelayMs(attempt: number): number {
    const exponentialDelay = this.baseDelayMs * 2 ** (attempt - 1);
    const jitter = Math.random() * exponentialDelay;
    return exponentialDelay + jitter;
  }
}
