import { BaseError } from "../../domain/errors/BaseError";
import { Message } from "../../domain/message/Message";
import { Direction, MessageStatus } from "../../domain/types/enums";
import { handleUseCaseError } from "../services/handleUseCaseError";

import type { RetryStuckClaims } from "./RetryStuckClaims";
import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "../../domain/ports/IMetrics";
import type { IProvider } from "../../domain/ports/IProvider";
import type { IRateLimitChecker } from "../../domain/ports/IRateLimitChecker";
import type { TransactionReceipt } from "../../domain/types/blockchain";
import type { MonitorClaimReceiptsConfig } from "../config/PostmanConfig";

export class MonitorClaimReceipts {
  constructor(
    private readonly repository: IMessageRepository,
    private readonly rateLimitChecker: IRateLimitChecker,
    private readonly sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater,
    private readonly transactionMetricsUpdater: ITransactionMetricsUpdater,
    private readonly provider: IProvider,
    private readonly retryStuckClaims: RetryStuckClaims,
    private readonly errorParser: IErrorParser,
    private readonly config: MonitorClaimReceiptsConfig,
    private readonly logger: ILogger,
  ) {}

  public async process(): Promise<void> {
    let firstPendingMessage: Message | null = null;
    try {
      firstPendingMessage = await this.repository.getFirstPendingMessage(this.config.direction);
      if (!firstPendingMessage?.claimTxHash) {
        this.logger.info("No pending message status to update.");
        return;
      }

      const receipt = await this.provider.getTransactionReceipt(firstPendingMessage.claimTxHash);
      if (receipt) {
        const receiptReceivedAt = new Date();
        await this.updateReceiptStatus(firstPendingMessage, receipt, receiptReceivedAt);
      } else {
        if (!this.isMessageExceededSubmissionTimeout(firstPendingMessage)) return;
        this.logger.warn("Retrying to claim message: messageHash=%s", firstPendingMessage.messageHash);

        const retryTransactionResponse = await this.retryStuckClaims.retry(firstPendingMessage);
        if (!retryTransactionResponse) return;
        const receiptReceivedAt = new Date();
        this.logger.warn(
          "Retried claim message transaction succeed: messageHash=%s transactionHash=%s",
          firstPendingMessage.messageHash,
          retryTransactionResponse.hash,
        );

        const retryReceipt = await this.provider.getTransactionReceipt(retryTransactionResponse.hash);
        if (!retryReceipt) {
          throw new BaseError(
            `RetryTransaction: Transaction receipt not found after retry transaction. transactionHash=${retryTransactionResponse.hash}`,
          );
        }
        await this.updateReceiptStatus(firstPendingMessage, retryReceipt, receiptReceivedAt);
      }
    } catch (e) {
      await handleUseCaseError({
        error: e,
        errorParser: this.errorParser,
        logger: this.logger,
        context: {
          operation: "MonitorClaimReceipts",
          direction: this.config.direction,
          messageHash: firstPendingMessage?.messageHash,
        },
      });
    }
  }

  private isMessageExceededSubmissionTimeout(message: Message): boolean {
    return (
      !!message.updatedAt && new Date().getTime() - message.updatedAt.getTime() > this.config.messageSubmissionTimeout
    );
  }

  private async updateReceiptStatus(
    message: Message,
    receipt: TransactionReceipt,
    receiptReceivedAt: Date,
  ): Promise<void> {
    let processingTimeInSeconds: number | undefined;
    let infuraConfirmationTimeInSeconds: number | undefined;

    if (this.config.direction === Direction.L1_TO_L2 && message.claimTxCreationDate) {
      const block = await this.provider.getBlock(receipt.blockNumber);
      if (block) {
        processingTimeInSeconds = block.timestamp - message.claimTxCreationDate.getTime() / 1_000;
        infuraConfirmationTimeInSeconds = (receiptReceivedAt.getTime() - message.claimTxCreationDate.getTime()) / 1_000;

        this.transactionMetricsUpdater.addTransactionProcessingTime(this.config.direction, processingTimeInSeconds);
        this.transactionMetricsUpdater.addTransactionInfuraConfirmationTime(
          this.config.direction,
          infuraConfirmationTimeInSeconds,
        );
      }
    }

    if (receipt.status === "reverted") {
      const isRateLimitExceeded = await this.rateLimitChecker.isRateLimitExceededError(receipt.transactionHash);

      if (isRateLimitExceeded) {
        message.edit({ status: MessageStatus.SENT });
        await this.repository.updateMessage(message);

        this.logger.warn(
          "Claim transaction has been reverted with RateLimitExceeded error. Claiming will be retry later: messageHash=%s transactionHash=%s",
          message.messageHash,
          receipt.transactionHash,
        );
        return;
      }

      message.edit({ status: MessageStatus.CLAIMED_REVERTED });
      await this.repository.updateMessage(message);
      this.logger.warn(
        "Message claim transaction has been REVERTED: messageHash=%s transactionHash=%s",
        message.messageHash,
        receipt.transactionHash,
        {
          ...(processingTimeInSeconds ? { processingTimeInSeconds } : {}),
          ...(infuraConfirmationTimeInSeconds ? { infuraConfirmationTimeInSeconds } : {}),
        },
      );
      return;
    }

    message.edit({ status: MessageStatus.CLAIMED_SUCCESS });

    await this.repository.updateMessage(message);

    if (message.isForSponsorship) {
      await this.sponsorshipMetricsUpdater.incrementSponsorshipFeePaid(
        receipt.gasPrice * receipt.gasUsed,
        message.direction,
      );
    }

    this.logger.info(
      "Message has been SUCCESSFULLY claimed: messageHash=%s transactionHash=%s",
      message.messageHash,
      receipt.transactionHash,
      {
        ...(processingTimeInSeconds ? { processingTimeInSeconds } : {}),
        ...(infuraConfirmationTimeInSeconds ? { infuraConfirmationTimeInSeconds } : {}),
      },
    );
  }
}
