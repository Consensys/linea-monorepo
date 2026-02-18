import { MessageStatus, OnChainMessageStatus } from "../../domain/types/enums";
import { handleUseCaseError } from "../services/handleUseCaseError";

import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { IMessageStatusChecker } from "../../domain/ports/IMessageStatusChecker";
import type { MessageAnchoringProcessorConfig } from "../config/PostmanConfig";

export class AnchorMessages {
  private readonly maxFetchMessagesFromDb: number;

  constructor(
    private readonly statusChecker: IMessageStatusChecker,
    private readonly repository: IMessageRepository,
    private readonly errorParser: IErrorParser,
    private readonly config: MessageAnchoringProcessorConfig,
    private readonly logger: ILogger,
  ) {
    this.maxFetchMessagesFromDb = Math.max(config.maxFetchMessagesFromDb, 0);
  }

  public async process(): Promise<void> {
    try {
      const messages = await this.repository.getNFirstMessagesByStatus(
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

      for (const message of messages) {
        const messageStatus = await this.statusChecker.getMessageStatus({
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

      await this.repository.saveMessages(messages);
    } catch (e) {
      await handleUseCaseError({
        error: e,
        errorParser: this.errorParser,
        logger: this.logger,
        context: { operation: "AnchorMessages", direction: this.config.direction },
      });
    }
  }
}
