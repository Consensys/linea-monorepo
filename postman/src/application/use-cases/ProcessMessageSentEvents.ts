import { MessageFactory } from "../../domain/message/MessageFactory";
import { MessageStatus } from "../../domain/types/enums";
import { serialize } from "../../domain/utils/serialize";

import type { ILogClient } from "../../domain/ports/ILogClient";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { IProvider } from "../../domain/ports/IProvider";
import type { MessageEventFilter } from "../../domain/services/MessageEventFilter";
import type { MessageSentEventProcessorConfig } from "../config/PostmanConfig";

export class ProcessMessageSentEvents {
  private readonly maxBlocksToFetchLogs: number;

  constructor(
    private readonly repository: IMessageRepository,
    private readonly logClient: ILogClient,
    private readonly provider: IProvider,
    private readonly messageFilter: MessageEventFilter,
    protected readonly config: MessageSentEventProcessorConfig,
    private readonly logger: ILogger,
  ) {
    this.maxBlocksToFetchLogs = Math.max(config.maxBlocksToFetchLogs, 0);
  }

  public async process(
    fromBlock: number,
    fromBlockLogIndex: number,
  ): Promise<{ nextFromBlock: number; nextFromBlockLogIndex: number }> {
    const latestBlockNumber = Math.max((await this.provider.getBlockNumber()) - this.config.blockConfirmation, 0);
    const toBlock = Math.min(latestBlockNumber, fromBlock + this.maxBlocksToFetchLogs);

    fromBlock = this.calculateFromBlockNumber(fromBlock, toBlock);

    this.logger.info("Getting events fromBlock=%s toBlock=%s", fromBlock, toBlock);

    const events = await this.logClient.getMessageSentEvents({
      filters: {
        from: this.config.eventFilters?.fromAddressFilter,
        to: this.config.eventFilters?.toAddressFilter,
      },
      fromBlock: BigInt(fromBlock),
      toBlock: BigInt(toBlock),
      fromBlockLogIndex,
    });

    this.logger.info("Number of fetched MessageSent events: %s", events.length);

    for (const event of events) {
      const shouldBeProcessed = this.messageFilter.shouldProcess(event, this.config.eventFilters?.calldataFilter);
      const messageStatusToInsert = shouldBeProcessed ? MessageStatus.SENT : MessageStatus.EXCLUDED;

      const message = MessageFactory.createMessage({
        ...event,
        sentBlockNumber: event.blockNumber,
        direction: this.config.direction,
        status: messageStatusToInsert,
        claimNumberOfRetry: 0,
      });

      await this.repository.insertMessage(message);
    }
    this.logger.info(`Messages hashes found: messageHashes=%s`, serialize(events.map((event) => event.messageHash)));

    return { nextFromBlock: toBlock + 1, nextFromBlockLogIndex: 0 };
  }

  private calculateFromBlockNumber(fromBlockNumber: number, toBlockNumber: number): number {
    if (fromBlockNumber > toBlockNumber) {
      return toBlockNumber;
    }
    return Math.max(fromBlockNumber, 0);
  }
}
