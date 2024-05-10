import { Direction } from "../../core/enums/MessageEnums";
import { ILogger } from "../../core/utils/logging/ILogger";
import { DEFAULT_LISTENER_INTERVAL } from "../../core/constants";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";
import { IMessageSentEventProcessor } from "../../core/services/processors/IMessageSentEventProcessor";
import { Message } from "../../core/entities/Message";
import { DatabaseAccessError } from "../../core/errors/DatabaseErrors";
import { IChainQuerier } from "../../core/clients/blockchain/IChainQuerier";
import { IPoller } from "../../core/services/pollers/IPoller";
import { wait } from "../../core/utils/shared";
import { L1NetworkConfig, L2NetworkConfig } from "../../application/postman/app/config/config";

export class MessageSentEventPoller implements IPoller {
  private isPolling = false;
  private pollingInterval: number;
  private initialFromBlock?: number;
  private originContractAddress: string;

  /**
   * Constructs a new instance of the `MessageSentEventPoller`.
   *
   * @param {IMessageSentEventProcessor} eventProcessor - An instance of a class implementing the `IMessageSentEventProcessor` interface, responsible for processing message sent events.
   * @param {IChainQuerier<unknown>} chainQuerier - An instance of a class implementing the `IChainQuerier` interface, used to query blockchain data.
   * @param {IMessageRepository<unknown>} messageRepository - An instance of a class implementing the `IMessageRepository` interface, used for storing and retrieving message data.
   * @param {Direction} direction - The direction of message flow (L1 to L2 or L2 to L1) that this poller is monitoring.
   * @param {L1NetworkConfig | L2NetworkConfig} config - Configuration settings for the network, including the polling interval and the initial block number to start listening from.
   * @param {ILogger} logger - An instance of a class implementing the `ILogger` interface, used for logging messages related to the polling process.
   */
  constructor(
    private readonly eventProcessor: IMessageSentEventProcessor,
    private readonly chainQuerier: IChainQuerier<unknown>,
    private readonly messageRepository: IMessageRepository<unknown>,
    private readonly direction: Direction,
    config: L1NetworkConfig | L2NetworkConfig,
    private readonly logger: ILogger,
  ) {
    this.pollingInterval = config.listener.pollingInterval ?? DEFAULT_LISTENER_INTERVAL;
    this.initialFromBlock = config.listener.initialFromBlock;
    this.originContractAddress = config.messageServiceContractAddress;
  }

  /**
   * Starts the polling process, initiating the continuous monitoring and processing of message sent events.
   */
  public async start() {
    if (this.isPolling) {
      this.logger.warn("%s has already started.", this.logger.name);
      return;
    }

    this.logger.info("Starting %s %s...", this.direction, this.logger.name);

    this.isPolling = true;
    this.startProcessingEvents();
  }

  /**
   * Stops the polling process, halting any further monitoring and processing of message sent events.
   */
  public stop() {
    this.logger.info("Stopping %s %s...", this.direction, this.logger.name);
    this.isPolling = false;
    this.logger.info("%s %s stopped.", this.direction, this.logger.name);
  }

  /**
   * Initiates the event processing loop, fetching and processing message sent events starting from a determined block number.
   *
   * In case of errors during the initial block number retrieval or event processing, it attempts to restart the process after waiting for the specified polling interval.
   */
  private async startProcessingEvents(): Promise<void> {
    try {
      const { fromBlock, fromBlockLogIndex } = await this.getInitialFromBlock();
      this.processEvents(fromBlock, fromBlockLogIndex);
    } catch (e) {
      this.logger.error(e);
      await wait(this.pollingInterval);
      this.startProcessingEvents();
    }
  }

  /**
   * Processes message sent events starting from a specific block number and log index.
   *
   * This method continuously fetches and processes events, updating the starting point for the next fetch based on the processed events. In case of database access errors, it attempts to recover by restarting from the last successfully processed block number and log index.
   *
   * @param {number} fromBlock - The block number to start fetching events from.
   * @param {number} fromBlockLogIndex - The log index within the starting block to begin processing events from.
   */
  private async processEvents(fromBlock: number, fromBlockLogIndex: number): Promise<void> {
    if (!this.isPolling) return;

    try {
      const { nextFromBlock, nextFromBlockLogIndex } = await this.eventProcessor.getAndStoreMessageSentEvents(
        fromBlock,
        fromBlockLogIndex,
      );
      fromBlock = nextFromBlock;
      fromBlockLogIndex = nextFromBlockLogIndex;
    } catch (e) {
      if (e instanceof DatabaseAccessError) {
        fromBlock = (e.rejectedMessage as Message & { logIndex: number }).sentBlockNumber;
        fromBlockLogIndex = (e.rejectedMessage as Message & { logIndex: number }).logIndex;
        this.logger.warn(
          "Something went wrong with database access. Restarting fromBlockNum=%s and fromLogIndex=%s and errorMessage=%s",
          fromBlock,
          fromBlockLogIndex,
          e.message,
        );
      } else {
        this.logger.warnOrError(e);
      }
    } finally {
      await wait(this.pollingInterval);
      this.processEvents(fromBlock, fromBlockLogIndex);
    }
  }

  /**
   * Determines the initial block number and log index to start fetching message sent events from.
   *
   * This method considers the latest message sent block number, the configured initial block number, and the current block number to determine the most appropriate starting point for event processing.
   *
   * @returns {Promise<{ fromBlock: number; fromBlockLogIndex: number }>} An object containing the determined starting block number and log index.
   */
  private async getInitialFromBlock(): Promise<{ fromBlock: number; fromBlockLogIndex: number }> {
    let fromBlock = await this.chainQuerier.getCurrentBlockNumber();

    const fromBlockLogIndex = 0;

    const latestMessageSentBlockNumber = await this.getLatestMessageSentBlockNumber(this.direction);

    if (latestMessageSentBlockNumber) {
      fromBlock = latestMessageSentBlockNumber;
    }

    if (this.initialFromBlock || this.initialFromBlock === 0) {
      fromBlock = this.initialFromBlock;
    }

    return { fromBlock, fromBlockLogIndex };
  }

  /**
   * Retrieves the block number of the latest message sent event processed by the application, based on the specified direction and contract address.
   *
   * @param {Direction} direction - The direction of message flow to consider when retrieving the latest message sent event.
   * @returns {Promise<number | null>} The block number of the latest message sent event, or null if no such event has been processed.
   */
  private async getLatestMessageSentBlockNumber(direction: Direction): Promise<number | null> {
    const lastMessageSent = await this.messageRepository.getLatestMessageSent(direction, this.originContractAddress);

    if (!lastMessageSent) {
      return null;
    }

    return lastMessageSent.sentBlockNumber;
  }
}
