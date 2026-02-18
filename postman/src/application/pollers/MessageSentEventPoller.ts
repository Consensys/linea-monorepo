import { DEFAULT_INITIAL_FROM_BLOCK } from "../../domain/constants";
import { DatabaseAccessError } from "../../domain/errors/DatabaseAccessError";
import { Message } from "../../domain/message/Message";
import { wait } from "../../domain/utils/wait";

import type { IPoller } from "./Poller";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { IProvider } from "../../domain/ports/IProvider";
import type { Direction } from "../../domain/types/enums";
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
    private readonly logger: ILogger,
  ) {}

  public async start(): Promise<void> {
    if (this.isPolling) {
      this.logger.warn("%s has already started.", this.logger.name);
      return;
    }

    this.logger.info("Starting %s %s...", this.config.direction, this.logger.name);
    this.isPolling = true;

    let fromBlock: number;
    let fromBlockLogIndex: number;

    while (this.isPolling) {
      try {
        if (fromBlock! === undefined) {
          const initial = await this.getInitialFromBlock();
          fromBlock = initial.fromBlock;
          fromBlockLogIndex = initial.fromBlockLogIndex;
        }

        const result = await this.eventProcessor.process(fromBlock!, fromBlockLogIndex!);
        fromBlock = result.nextFromBlock;
        fromBlockLogIndex = result.nextFromBlockLogIndex;
      } catch (e) {
        if (e instanceof DatabaseAccessError) {
          const rejected = e.rejectedMessage as Message & { logIndex: number };
          fromBlock = rejected.sentBlockNumber;
          fromBlockLogIndex = rejected.logIndex;
          this.logger.warn(
            "Database access error, restarting from block=%s logIndex=%s: %s",
            fromBlock,
            fromBlockLogIndex,
            e.message,
          );
        } else {
          this.logger.error("MessageSentEventPoller %s encountered an error: %s", this.config.direction, e);
        }
      }

      await wait(this.config.pollingInterval);
    }
  }

  public stop(): void {
    this.logger.info("Stopping %s %s...", this.config.direction, this.logger.name);
    this.isPolling = false;
    this.logger.info("%s %s stopped.", this.config.direction, this.logger.name);
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
