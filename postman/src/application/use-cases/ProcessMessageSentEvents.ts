import { compileExpression, useDotAccessOperator } from "filtrex";

import { MessageFactory } from "../../domain/message/MessageFactory";
import { MessageStatus } from "../../domain/types/MessageStatus";
import { isEmptyBytes } from "../../domain/utils/isEmptyBytes";
import { serialize } from "../../domain/utils/serialize";

import type { ICalldataDecoder } from "../../domain/ports/ICalldataDecoder";
import type { ILogClient } from "../../domain/ports/ILogClient";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { IProvider } from "../../domain/ports/IProvider";
import type { MessageSent } from "../../domain/types/Events";
import type { MessageSentEventProcessorConfig } from "../config/PostmanConfig";

export class ProcessMessageSentEvents {
  private readonly maxBlocksToFetchLogs: number;

  constructor(
    private readonly repository: IMessageRepository,
    private readonly logClient: ILogClient,
    private readonly provider: IProvider,
    private readonly calldataDecoder: ICalldataDecoder | null,
    protected readonly config: MessageSentEventProcessorConfig,
    private readonly logger: ILogger,
  ) {
    this.maxBlocksToFetchLogs = Math.max(config.maxBlocksToFetchLogs, 0);
  }

  private calculateFromBlockNumber(fromBlockNumber: number, toBlockNumber: number): number {
    if (fromBlockNumber > toBlockNumber) {
      return toBlockNumber;
    }
    return Math.max(fromBlockNumber, 0);
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
      const shouldBeProcessed = this.shouldProcessMessage(
        event,
        event.messageHash,
        this.config.eventFilters?.calldataFilter,
      );
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

  protected shouldProcessMessage(
    event: MessageSent,
    messageHash: string,
    filters?: {
      criteriaExpression: string;
      calldataFunctionInterface: string;
    },
  ): boolean {
    const hasEmptyCalldata = isEmptyBytes(event.calldata);
    let basicProcess = false;

    if (hasEmptyCalldata) {
      basicProcess = this.config.isEOAEnabled;
    } else {
      basicProcess = this.config.isCalldataEnabled;
    }

    if (!basicProcess) {
      this.logger.debug(
        "Message has been excluded because target address is not an EOA or calldata is not empty: messageHash=%s",
        messageHash,
      );
      return false;
    }

    if (!hasEmptyCalldata && this.config.isCalldataEnabled && !this.isMessageMatchingCriteria(event, filters)) {
      return false;
    }

    return true;
  }

  private isMessageMatchingCriteria(
    event: MessageSent,
    filters?: { criteriaExpression: string; calldataFunctionInterface: string },
  ): boolean {
    if (!filters) {
      return true;
    }

    if (!this.calldataDecoder) {
      return true;
    }

    const decoded = this.calldataDecoder.decode(event.calldata);
    if (!decoded) {
      return false;
    }

    const funcSignature = event.calldata.slice(0, 10);

    const context = {
      calldata: {
        funcSignature,
        ...this.convertBigInts(decoded.args),
      },
    };

    const passesFilter = this.evaluateExpression(filters.criteriaExpression, context);

    if (!passesFilter) {
      this.logger.debug(
        "Message has been excluded because it does not match the criteria: criteria=%s messageHash=%s transactionHash=%s",
        filters.criteriaExpression,
        event.messageHash,
        event.transactionHash,
      );
      return false;
    }

    return true;
  }

  private evaluateExpression(expression: string, context: unknown): boolean {
    try {
      const compiledFilter = compileExpression(expression, { customProp: useDotAccessOperator });
      const passesFilter = compiledFilter(context);
      return passesFilter === true;
    } catch {
      return false;
    }
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private convertBigInts(data: any): any {
    if (typeof data === "bigint") {
      return Number(data);
    }

    if (Array.isArray(data)) {
      return data.map((item) => this.convertBigInts(item));
    }

    if (data !== null && typeof data === "object") {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const result: Record<string, any> = {};
      for (const key in data) {
        if (Object.prototype.hasOwnProperty.call(data, key)) {
          result[key] = this.convertBigInts(data[key]);
        }
      }
      return result;
    }

    return data;
  }
}
