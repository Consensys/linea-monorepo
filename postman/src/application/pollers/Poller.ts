import { wait } from "../../domain/utils/wait";

import type { ILogger } from "../../domain/ports/ILogger";

export interface IPoller {
  start(): Promise<void>;
  stop(): void;
}

export class Poller implements IPoller {
  private isPolling = false;

  constructor(
    private readonly name: string,
    private readonly pollFn: () => Promise<void>,
    private readonly intervalMs: number,
    private readonly logger: ILogger,
  ) {}

  public async start(): Promise<void> {
    if (this.isPolling) {
      this.logger.warn("%s has already started.", this.name);
      return;
    }
    this.logger.info("Starting %s...", this.name);
    this.isPolling = true;

    while (this.isPolling) {
      try {
        await this.pollFn();
      } catch (e) {
        this.logger.error("Poller %s encountered an error: %s", this.name, e);
      }
      await wait(this.intervalMs);
    }
  }

  public stop(): void {
    this.logger.info("Stopping %s...", this.name);
    this.isPolling = false;
    this.logger.info("%s stopped.", this.name);
  }
}
