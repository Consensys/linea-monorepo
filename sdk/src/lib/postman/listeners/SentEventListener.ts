import { DataSource } from "typeorm";
import { LineaLogger } from "../../logger";
import { EventParser } from "../../contracts/EventParser";
import { L1MessageServiceContract, L2MessageServiceContract } from "../../contracts";
import { MessageRepository } from "../repositories/MessageRepository";
import { MessageInDb, L1NetworkConfig, L2NetworkConfig } from "../utils/types";
import { DatabaseAccessError } from "../utils/errors";
import { Direction, MessageStatus } from "../utils/enums";
import { ErrorParser, ParsableError } from "../../errorHandlers";
import {
  DEFAULT_LISTENER_BLOCK_CONFIRMATIONS,
  DEFAULT_LISTENER_INTERVAL,
  DEFAULT_MAX_BLOCKS_TO_FETCH_LOGS,
  DEFAULT_ONLY_EOA_TARGET,
} from "../../utils/constants";
import { isEmptyBytes, wait } from "../utils/helpers";

export abstract class SentEventListener<
  TMessageServiceContract extends L1MessageServiceContract | L2MessageServiceContract,
> {
  public logger: LineaLogger;
  protected maxBlocksToFetchLogs: number;
  protected backtrackNumberOfBlocks: number;
  public shouldStopListening: boolean;
  public messageRepository: MessageRepository;
  protected originContractAddress: string;
  protected blockConfirmation: number;
  protected pollingInterval: number;
  protected initialFromBlock?: number;
  protected onlyEOA: boolean;

  constructor(
    dataSource: DataSource,
    public readonly messageServiceContract: TMessageServiceContract,
    config: L1NetworkConfig | L2NetworkConfig,
    protected readonly direction: Direction,
  ) {
    this.maxBlocksToFetchLogs = Math.max(config.listener.maxBlocksToFetchLogs ?? DEFAULT_MAX_BLOCKS_TO_FETCH_LOGS, 0);
    this.shouldStopListening = false;
    this.messageRepository = new MessageRepository(dataSource);
    this.originContractAddress = config.messageServiceContractAddress;
    this.pollingInterval = config.listener.pollingInterval ?? DEFAULT_LISTENER_INTERVAL;
    this.blockConfirmation = config.listener.blockConfirmation ?? DEFAULT_LISTENER_BLOCK_CONFIRMATIONS;
    this.initialFromBlock = config.listener.initialFromBlock;
    this.onlyEOA = config.onlyEOATarget || DEFAULT_ONLY_EOA_TARGET;
  }

  public calculateFromBlockNumber(fromBlockNumber: number, toBlockNumber: number): number {
    if (fromBlockNumber > toBlockNumber) {
      return toBlockNumber;
    }
    return Math.max(fromBlockNumber, 0);
  }

  public async getLatestMessageSentBlockNumber(direction: Direction): Promise<number | null> {
    try {
      const lastMessageSent = await this.messageRepository.getLatestMessageSent(direction, this.originContractAddress);

      if (!lastMessageSent) {
        return null;
      }

      return lastMessageSent.sentBlockNumber;
    } catch (error) {
      this.logger.error(error);
      return null;
    }
  }

  public async start() {
    let fromBlock = await this.messageServiceContract.contract.provider.getBlockNumber();

    const fromBlockLogIndex = 0;

    const latestMessageSentBlockNumber = await this.getLatestMessageSentBlockNumber(this.direction);

    if (latestMessageSentBlockNumber) {
      fromBlock = latestMessageSentBlockNumber;
    }

    if (this.initialFromBlock || this.initialFromBlock === 0) {
      fromBlock = this.initialFromBlock;
    }

    if (!this.shouldStopListening) {
      this.listenForMessageSentEvents(this.pollingInterval, fromBlock, fromBlockLogIndex);
    }
  }

  public stop() {
    this.logger.info("Stopping SentEventListener...");
    this.shouldStopListening = true;
    this.logger.info("SentEventListener stopped.");
  }

  public async listenForMessageSentEvents(interval: number, fromBlock: number, fromBlockLogIndex: number) {
    const latestBlockNumber = Math.max(
      (await this.messageServiceContract.getCurrentBlockNumber()) - this.blockConfirmation,
      0,
    );
    const toBlock = Math.min(latestBlockNumber, fromBlock + this.maxBlocksToFetchLogs);

    fromBlock = this.calculateFromBlockNumber(fromBlock, toBlock);

    this.logger.info(`Interval reached every ${interval} ms, checking from ${fromBlock} to ${toBlock}`);

    try {
      const events = await this.messageServiceContract.getEvents(
        this.messageServiceContract.contract.filters.MessageSent(),
        fromBlock,
        toBlock,
        fromBlockLogIndex,
      );

      this.logger.info(`# of fetched MessageSent events: ${events.length}`);

      for (const event of events) {
        const shouldBeExcluded = this.shouldExcludeMessage(event.args?._calldata, event.args?._messageHash);
        const messageStatusToInsert = shouldBeExcluded ? MessageStatus.EXCLUDED : MessageStatus.SENT;

        const message = EventParser.parsedEventToMessage(event, this.direction, messageStatusToInsert);
        this.logger.info(`message: ${JSON.stringify(message.messageHash)}`);

        await this.messageRepository.insertMessage(message);
      }
      fromBlock = toBlock + 1;
      fromBlockLogIndex = 0;
    } catch (e) {
      if (e instanceof DatabaseAccessError) {
        fromBlock = (e.rejectedMessage as MessageInDb & { logIndex: number }).sentBlockNumber;
        fromBlockLogIndex = (e.rejectedMessage as MessageInDb & { logIndex: number }).logIndex;
        this.logger.warn(`${e.message} fromBlockNum = ${fromBlock} fromLogIndex = ${fromBlockLogIndex}`);
      } else {
        const parsedError = ErrorParser.parseErrorWithMitigation(e as ParsableError);
        this.logger.error(
          `Error found in listenForMessageSentEvents:\nFounded error: ${JSON.stringify(
            e,
          )}\nParsed error: ${JSON.stringify(parsedError)}`,
        );
      }
    }

    await wait(interval);

    if (!this.shouldStopListening) {
      this.listenForMessageSentEvents(interval, fromBlock, fromBlockLogIndex);
    }
  }

  protected shouldExcludeMessage(messageCalldata: string, messageHash: string): boolean {
    if (!this.onlyEOA) {
      return false;
    }

    if (isEmptyBytes(messageCalldata)) {
      return false;
    }

    this.logger.warn(
      `Message with hash ${messageHash} has been excluded because target address is not an EOA or calldata is not empty.`,
    );
    return true;
  }
}
