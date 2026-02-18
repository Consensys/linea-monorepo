import { DEFAULT_INITIAL_FROM_BLOCK } from "../../domain/constants";
import { DatabaseAccessError } from "../../domain/errors/DatabaseAccessError";
import { Message } from "../../domain/message/Message";
import { wait } from "../../domain/utils/wait";

import type { IPoller } from "./Poller";
import type { IPostmanLogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { IProvider } from "../../domain/ports/IProvider";
import type { Direction } from "../../domain/types/Direction";
import type { ProcessMessageSentEvents } from "../use-cases/ProcessMessageSentEvents";

export type MessageSentEventPollerConfig = {
  direction: Direction;
  pollingInterval: number;
  initialFromBlock: number;
  originContractAddress: string;
};

export class MessageSentEventPoller implements IPoller {
  private isPolling = false;

  constructor(
    private readonly eventProcessor: ProcessMessageSentEvents,
    private readonly provider: IProvider,
    private readonly repository: IMessageRepository,
    private readonly config: MessageSentEventPollerConfig,
    private readonly logger: IPostmanLogger,
  ) {}

  public async start(): Promise<void> {
    if (this.isPolling) {
      this.logger.warn("%s has already started.", this.logger.name);
      return;
    }

    this.logger.info("Starting %s %s...", this.config.direction, this.logger.name);

    this.isPolling = true;
    this.startProcessingEvents();
  }

  public stop(): void {
    this.logger.info("Stopping %s %s...", this.config.direction, this.logger.name);
    this.isPolling = false;
    this.logger.info("%s %s stopped.", this.config.direction, this.logger.name);
  }

  private async startProcessingEvents(): Promise<void> {
    try {
      const { fromBlock, fromBlockLogIndex } = await this.getInitialFromBlock();
      this.processEvents(fromBlock, fromBlockLogIndex);
    } catch (e) {
      this.logger.error(e);
      await wait(this.config.pollingInterval);
      this.startProcessingEvents();
    }
  }

  private async processEvents(fromBlock: number, fromBlockLogIndex: number): Promise<void> {
    if (!this.isPolling) return;

    try {
      const { nextFromBlock, nextFromBlockLogIndex } = await this.eventProcessor.process(fromBlock, fromBlockLogIndex);
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
      await wait(this.config.pollingInterval);
      this.processEvents(fromBlock, fromBlockLogIndex);
    }
  }

  private async getInitialFromBlock(): Promise<{ fromBlock: number; fromBlockLogIndex: number }> {
    let fromBlock = await this.provider.getBlockNumber();

    const fromBlockLogIndex = 0;

    const latestMessageSentBlockNumber = await this.getLatestMessageSentBlockNumber(this.config.direction);

    if (latestMessageSentBlockNumber) {
      fromBlock = latestMessageSentBlockNumber;
    }

    if (this.config.initialFromBlock > DEFAULT_INITIAL_FROM_BLOCK) {
      fromBlock = this.config.initialFromBlock;
    }

    return { fromBlock, fromBlockLogIndex };
  }

  private async getLatestMessageSentBlockNumber(direction: Direction): Promise<number | null> {
    const lastMessageSent = await this.repository.getLatestMessageSent(direction, this.config.originContractAddress);

    if (!lastMessageSent) {
      return null;
    }

    return lastMessageSent.sentBlockNumber;
  }
}
