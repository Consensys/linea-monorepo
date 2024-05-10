import { Direction } from "../../core/enums/MessageEnums";
import { ILogger } from "../../core/utils/logging/ILogger";
import { DEFAULT_LISTENER_INTERVAL } from "../../core/constants";
import { IPoller } from "../../core/services/pollers/IPoller";
import { IMessageClaimingProcessor } from "../../core/services/processors/IMessageClaimingProcessor";
import { L1NetworkConfig, L2NetworkConfig } from "../../application/postman/app/config/config";
import { wait } from "../../core/utils/shared";

export class MessageClaimingPoller implements IPoller {
  private isPolling = false;
  private pollingInterval: number;

  /**
   * Constructs a new instance of the `MessageClaimingPoller`.
   *
   * @param {IMessageClaimingProcessor} claimingProcessor - An instance of a class implementing the `IMessageClaimingProcessor` interface, responsible for the message claiming logic.
   * @param {Direction} direction - The direction of message flow (L1 to L2 or L2 to L1) that this poller is handling.
   * @param {L1NetworkConfig | L2NetworkConfig} config - Configuration settings for the network, including the polling interval for the listener.
   * @param {ILogger} logger - An instance of a class implementing the `ILogger` interface, used for logging messages related to the polling process.
   */
  constructor(
    private readonly claimingProcessor: IMessageClaimingProcessor,
    private readonly direction: Direction,
    config: L1NetworkConfig | L2NetworkConfig,
    private readonly logger: ILogger,
  ) {
    this.pollingInterval = config.listener.pollingInterval ?? DEFAULT_LISTENER_INTERVAL;
  }

  /**
   * Starts the polling process, triggering periodic execution of the message claiming logic.
   */
  public async start() {
    if (this.isPolling) {
      this.logger.warn("%s has already started.", this.logger.name);
      return;
    }
    this.logger.info("Starting %s %s...", this.direction, this.logger.name);
    this.isPolling = true;

    while (this.isPolling) {
      await this.claimingProcessor.getAndClaimAnchoredMessage();
      await wait(this.pollingInterval);
    }
  }
  /**
   * Stops the polling process, halting any further execution of the message claiming logic.
   */
  public stop() {
    this.logger.info("Stopping %s %s...", this.direction, this.logger.name);
    this.isPolling = false;
    this.logger.info("%s %s stopped.", this.direction, this.logger.name);
  }
}
