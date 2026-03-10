import { ILogger } from "@consensys/linea-shared-utils";

import { Direction } from "../../core/enums";
import { IPoller } from "../../core/services/pollers/IPoller";
import { IMessageClaimingPersister } from "../../core/services/processors/IMessageClaimingPersister";
import { wait } from "../../core/utils/shared";

type MessagePersistingPollerConfig = {
  direction: Direction;
  pollingInterval: number;
};

export class MessagePersistingPoller implements IPoller {
  private isPolling = false;

  /**
   * Constructs a new instance of the `MessagePersistingPoller`.
   *
   * @param {IMessageClaimingPersister} claimingPersister - An instance of a class implementing the `IMessageClaimingPersister` interface, responsible for updating and persisting claimed messages.
   * @param {MessagePersistingPollerConfig} config - Configuration settings for the poller, including the direction of message flow and the polling interval.
   * @param {ILogger} logger - An instance of a class implementing the `ILogger` interface, used for logging messages related to the polling process.
   */
  constructor(
    private readonly claimingPersister: IMessageClaimingPersister,
    private config: MessagePersistingPollerConfig,
    private readonly logger: ILogger,
  ) {}

  /**
   * Starts the polling process, triggering periodic execution of the message persistence logic.
   * Logs a warning if the poller is already running.
   *
   * @returns {Promise<void>} A promise that resolves when the poller is started.
   */
  public async start(): Promise<void> {
    if (this.isPolling) {
      this.logger.warn("Poller has already started.", { name: this.logger.name });
      return;
    }
    this.logger.info("Starting poller.", { direction: this.config.direction, name: this.logger.name });
    this.isPolling = true;

    while (this.isPolling) {
      await this.claimingPersister.process();
      await wait(this.config.pollingInterval);
    }
  }

  /**
   * Stops the polling process, halting any further execution of the message persistence logic.
   * Logs information about the stopping process.
   */
  public stop() {
    this.logger.info("Stopping poller.", { direction: this.config.direction, name: this.logger.name });
    this.isPolling = false;
    this.logger.info("Poller stopped.", { direction: this.config.direction, name: this.logger.name });
  }
}
