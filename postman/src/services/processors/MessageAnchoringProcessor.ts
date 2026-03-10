import { ILogger } from "@consensys/linea-shared-utils";

import { IProvider } from "../../core/clients/blockchain/IProvider";
import { OnChainMessageStatus, MessageStatus } from "../../core/enums";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";
import { IMessageServiceContract } from "../../core/services/contracts/IMessageServiceContract";
import {
  IMessageAnchoringProcessor,
  MessageAnchoringProcessorConfig,
} from "../../core/services/processors/IMessageAnchoringProcessor";
import { ErrorParser } from "../../utils/ErrorParser";

export class MessageAnchoringProcessor implements IMessageAnchoringProcessor {
  private readonly maxFetchMessagesFromDb: number;

  /**
   * Constructs a new instance of the `MessageAnchoringProcessor`.
   *
   * @param {IMessageServiceContract} contractClient - An instance of a class implementing the `IMessageServiceContract` interface, used to interact with the blockchain contract.
   * @param {IProvider} provider - An instance of a class implementing the `IProvider` interface, used to query blockchain data.
   * @param {IMessageRepository} messageRepository - An instance of a class implementing the `IMessageRepository` interface, used for storing and retrieving message data.
   * @param {MessageAnchoringProcessorConfig} config - Configuration settings for the processor, including the maximum number of messages to fetch from the database for processing.
   * @param {ILogger} logger - An instance of a class implementing the `ILogger` interface, used for logging messages.
   */
  constructor(
    private readonly contractClient: IMessageServiceContract,
    private readonly provider: IProvider,
    private readonly messageRepository: IMessageRepository,
    private readonly config: MessageAnchoringProcessorConfig,
    private readonly logger: ILogger,
  ) {
    this.maxFetchMessagesFromDb = Math.max(config.maxFetchMessagesFromDb, 0);
  }

  /**
   * Fetches a set number of messages from the database and updates their status based on the latest anchoring information from the blockchain.
   *
   * @returns {Promise<void>} A promise that resolves when the processing is complete.
   */
  public async process(): Promise<void> {
    try {
      const messages = await this.messageRepository.getNFirstMessagesByStatus(
        MessageStatus.SENT,
        this.config.direction,
        this.maxFetchMessagesFromDb,
        this.config.originContractAddress,
      );

      if (messages.length === this.maxFetchMessagesFromDb) {
        this.logger.warn(`Limit of messages sent to listen reached (%s).`, this.maxFetchMessagesFromDb);
      }

      if (messages.length === 0) {
        this.logger.info("No messages to process for anchoring.");
        return;
      }

      const latestBlockNumber = await this.provider.getBlockNumber();

      for (const message of messages) {
        const messageStatus = await this.contractClient.getMessageStatus({
          messageHash: message.messageHash,
          messageBlockNumber: message.sentBlockNumber,
          overrides: { blockTag: latestBlockNumber },
        });

        if (messageStatus === OnChainMessageStatus.CLAIMABLE) {
          message.edit({ status: MessageStatus.ANCHORED });
          this.logger.info("Message has been anchored: messageHash=%s", message.messageHash);
        }

        if (messageStatus === OnChainMessageStatus.CLAIMED) {
          message.edit({ status: MessageStatus.CLAIMED_SUCCESS });
          this.logger.info("Message has already been claimed: messageHash=%s", message.messageHash);
        }
      }

      await this.messageRepository.saveMessages(messages);
    } catch (e) {
      const error = ErrorParser.parseErrorWithMitigation(e);
      this.logger.error("An error occurred while processing messages.", {
        errorCode: error?.errorCode,
        errorMessage: error?.errorMessage,
        ...(error?.data ? { data: error.data } : {}),
      });
    }
  }
}
