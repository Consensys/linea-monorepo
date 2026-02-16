import { MessageStatus } from "../../domain/types/MessageStatus";
import { OnChainMessageStatus } from "../../domain/types/OnChainMessageStatus";

import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageDBService } from "../../domain/ports/IMessageDBService";
import type { IMessageServiceContract } from "../../domain/ports/IMessageServiceContract";
import type { MessageAnchoringProcessorConfig } from "../config/PostmanConfig";

export class AnchorMessages {
  private readonly maxFetchMessagesFromDb: number;

  constructor(
    private readonly contractClient: IMessageServiceContract,
    private readonly databaseService: IMessageDBService,
    private readonly errorParser: IErrorParser,
    private readonly config: MessageAnchoringProcessorConfig,
    private readonly logger: ILogger,
  ) {
    this.maxFetchMessagesFromDb = Math.max(config.maxFetchMessagesFromDb, 0);
  }

  public async process(): Promise<void> {
    try {
      const messages = await this.databaseService.getNFirstMessagesSent(
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

      for (const message of messages) {
        const messageStatus = await this.contractClient.getMessageStatus({
          messageHash: message.messageHash,
          messageBlockNumber: message.sentBlockNumber,
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

      await this.databaseService.saveMessages(messages);
    } catch (e) {
      const error = this.errorParser.parseErrorWithMitigation(e);
      this.logger.error("An error occurred while processing messages.", {
        errorCode: error?.errorCode,
        errorMessage: error?.errorMessage,
        ...(error?.data ? { data: error.data } : {}),
      });
    }
  }
}
