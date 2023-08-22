import { DataSource } from "typeorm";
import { LineaLogger } from "../../logger";
import { MessageRepository } from "../repositories/MessageRepository";
import { L1NetworkConfig, L2NetworkConfig } from "../utils/types";
import { Direction, MessageStatus } from "../utils/enums";
import { L1MessageServiceContract, L2MessageServiceContract } from "../../contracts";
import { DEFAULT_LISTENER_INTERVAL, DEFAULT_MAX_FETCH_MESSAGES_FROM_DB } from "../../utils/constants";
import { ErrorParser, ParsableError } from "../../errorHandlers";
import { wait } from "../utils/helpers";
import { OnChainMessageStatus } from "../../utils/enum";

export abstract class AnchoredEventListener<
  TMessageServiceContract extends L1MessageServiceContract | L2MessageServiceContract,
> {
  protected logger: LineaLogger;
  protected maxFetchMessagesFromDb: number;
  protected shouldStopListening: boolean;
  public messageRepository: MessageRepository;
  protected originContractAddress: string;
  protected pollingInterval: number;

  constructor(
    dataSource: DataSource,
    private readonly messageServiceContract: TMessageServiceContract,
    config: L1NetworkConfig | L2NetworkConfig,
    protected readonly direction: Direction,
  ) {
    this.maxFetchMessagesFromDb = Math.max(
      config.listener.maxFetchMessagesFromDb ?? DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
      0,
    );
    this.shouldStopListening = false;
    this.messageRepository = new MessageRepository(dataSource);
    this.pollingInterval = config.listener.pollingInterval ?? DEFAULT_LISTENER_INTERVAL;
  }

  public async start() {
    this.logger.info("Starting AnchoredEventListener...");
    while (!this.shouldStopListening) {
      this.listenForMessageAnchoringEvents();
      await wait(this.pollingInterval);
    }
  }

  public stop() {
    this.logger.info("Stopping AnchoredEventListener...");
    this.shouldStopListening = true;
    this.logger.info("AnchoredEventListener stopped.");
  }

  public async listenForMessageAnchoringEvents() {
    try {
      const messagesSent = await this.messageRepository.getNFirstMessageSent(
        this.direction,
        this.maxFetchMessagesFromDb,
        this.originContractAddress,
      );

      if (messagesSent.length === this.maxFetchMessagesFromDb) {
        this.logger.warn(`Limit of messages sent to listen reached (${this.maxFetchMessagesFromDb}).`);
      }

      const latestBlockNumber = await this.messageServiceContract.getCurrentBlockNumber();

      if (messagesSent.length === 0) {
        return;
      }

      for (const message of messagesSent) {
        const messageStatus = await this.messageServiceContract.getMessageStatus(message.messageHash, {
          blockTag: latestBlockNumber,
        });

        if (messageStatus === OnChainMessageStatus.CLAIMABLE) {
          await this.messageRepository.updateMessage(message.messageHash, message.direction, {
            status: MessageStatus.ANCHORED,
          });
          this.logger.info(`Message hash ${message.messageHash} has been anchored.`);
        }

        if (messageStatus === OnChainMessageStatus.CLAIMED) {
          await this.messageRepository.updateMessage(message.messageHash, message.direction, {
            status: MessageStatus.CLAIMED_SUCCESS,
          });
          this.logger.info(`Message with hash ${message.messageHash} has already been claimed.`);
        }
      }
    } catch (e) {
      const parsedError = ErrorParser.parseErrorWithMitigation(e as ParsableError);
      this.logger.error(
        `Error found in listenForMessageAnchoringEvents:\nFounded error: ${JSON.stringify(
          e,
        )}\nParsed error: ${JSON.stringify(parsedError)}`,
      );
    }
  }
}
