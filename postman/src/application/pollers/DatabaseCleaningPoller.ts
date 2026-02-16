import { wait } from "../../domain/utils/wait";

import type { IPoller } from "./Poller";
import type { ILogger } from "../../domain/ports/ILogger";
import type { DBCleanerConfig } from "../config/PostmanConfig";
import type { CleanDatabase } from "../use-cases/CleanDatabase";

export class DatabaseCleaningPoller implements IPoller {
  private isPolling = false;
  private enabled: boolean;
  private pollingInterval: number;
  private msBeforeNowToDelete: number;

  constructor(
    private readonly databaseCleaner: CleanDatabase,
    private readonly logger: ILogger,
    config: DBCleanerConfig,
  ) {
    this.enabled = config.enabled;
    this.pollingInterval = config.cleaningInterval;
    this.msBeforeNowToDelete = config.daysBeforeNowToDelete * 24 * 60 * 60 * 1000;
  }

  public async start(): Promise<void> {
    if (!this.enabled) {
      this.logger.warn("%s is disabled", this.logger.name);
      return;
    }

    if (this.isPolling) {
      this.logger.warn("%s has already started.", this.logger.name);
      return;
    }
    this.logger.info("Starting %s...", this.logger.name);
    this.isPolling = true;

    while (this.isPolling) {
      await this.databaseCleaner.databaseCleanerRoutine(this.msBeforeNowToDelete);
      await wait(this.pollingInterval);
    }
  }

  public stop(): void {
    this.logger.info("Stopping %s...", this.logger.name);
    this.isPolling = false;
    this.logger.info("%s stopped.", this.logger.name);
  }
}
