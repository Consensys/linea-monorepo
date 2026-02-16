import { BaseError } from "../../domain/errors/BaseError";
import { Message } from "../../domain/message/Message";
import { Direction } from "../../domain/types/Direction";
import { MessageStatus } from "../../domain/types/MessageStatus";
import { OnChainMessageStatus } from "../../domain/types/OnChainMessageStatus";

import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageDBService } from "../../domain/ports/IMessageDBService";
import type { IMessageServiceContract } from "../../domain/ports/IMessageServiceContract";
import type { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "../../domain/ports/IMetrics";
import type { IProvider } from "../../domain/ports/IProvider";
import type { TransactionReceipt, TransactionResponse } from "../../domain/types/BlockchainTypes";
import type { MessageClaimingPersisterConfig } from "../config/PostmanConfig";

export class PersistClaimResults {
  private messageBeingRetry: { message: Message | null; retries: number };

  constructor(
    private readonly databaseService: IMessageDBService,
    private readonly messageServiceContract: IMessageServiceContract,
    private readonly sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater,
    private readonly transactionMetricsUpdater: ITransactionMetricsUpdater,
    private readonly provider: IProvider,
    private readonly errorParser: IErrorParser,
    private readonly config: MessageClaimingPersisterConfig,
    private readonly logger: ILogger,
  ) {
    this.messageBeingRetry = { message: null, retries: 0 };
  }

  private isMessageExceededSubmissionTimeout(message: Message): boolean {
    return (
      !!message.updatedAt && new Date().getTime() - message.updatedAt.getTime() > this.config.messageSubmissionTimeout
    );
  }

  public async process(): Promise<void> {
    let firstPendingMessage: Message | null = null;
    try {
      firstPendingMessage = await this.databaseService.getFirstPendingMessage(this.config.direction);
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

        if (
          !this.messageBeingRetry.message ||
          this.messageBeingRetry.message.messageHash !== firstPendingMessage.messageHash
        ) {
          this.messageBeingRetry = { message: firstPendingMessage, retries: 0 };
        }

        const retryTransactionResponse = await this.retryTransaction(
          firstPendingMessage.claimTxHash,
          firstPendingMessage.messageHash,
          firstPendingMessage.sentBlockNumber,
        );
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
      const error = this.errorParser.parseErrorWithMitigation(e);
      this.logger.error("Error processing message.", {
        ...(firstPendingMessage ? { messageHash: firstPendingMessage.messageHash } : {}),
        ...(error?.errorCode ? { errorCode: error.errorCode } : {}),
        ...(error?.errorMessage ? { errorMessage: error.errorMessage } : {}),
        ...(error?.data ? { data: error.data } : {}),
      });
    }
  }

  private async retryTransaction(
    transactionHash: string,
    messageHash: string,
    messageBlockNumber: number,
  ): Promise<TransactionResponse | null> {
    try {
      const messageStatus = await this.messageServiceContract.getMessageStatus({
        messageHash,
        messageBlockNumber,
      });

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        const receipt = await this.provider.getTransactionReceipt(transactionHash);
        if (!receipt) {
          this.logger.warn(
            "Calling retryTransaction again as message was claimed but transaction receipt is not available yet: messageHash=%s transactionHash=%s",
            messageHash,
            transactionHash,
          );
        }
        return receipt ? { hash: receipt.transactionHash, gasLimit: 0n, nonce: 0 } : null;
      }

      this.messageBeingRetry.retries++;
      this.logger.warn(
        "Retry to claim message: numberOfRetries=%s messageInfo=%s",
        this.messageBeingRetry.retries.toString(),
        this.messageBeingRetry.message?.toString(),
      );

      const tx = await this.messageServiceContract.retryTransactionWithHigherFee(transactionHash);

      this.messageBeingRetry.message?.edit({
        claimTxCreationDate: new Date(),
        claimTxGasLimit: parseInt(tx.gasLimit.toString()),
        claimTxMaxFeePerGas: tx.maxFeePerGas ?? undefined,
        claimTxMaxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
        claimTxHash: tx.hash,
        claimNumberOfRetry: this.messageBeingRetry.retries,
        claimLastRetriedAt: new Date(),
        claimTxNonce: tx.nonce,
      });
      await this.databaseService.updateMessage(this.messageBeingRetry.message!);

      return tx;
    } catch (e) {
      this.logger.error(
        "Transaction retry failed: messageHash=%s error=%s",
        this.messageBeingRetry.message?.messageHash,
        e,
      );
      if (this.messageBeingRetry.retries > this.config.maxTxRetries) {
        this.logger.error(
          "Max number of retries exceeded. Manual intervention is needed as soon as possible: messageInfo=%s",
          this.messageBeingRetry.message?.toString(),
        );
      }
      return null;
    }
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
      const isRateLimitExceeded = await this.messageServiceContract.isRateLimitExceededError(receipt.transactionHash);

      if (isRateLimitExceeded) {
        message.edit({ status: MessageStatus.SENT });
        await this.databaseService.updateMessage(message);

        this.logger.warn(
          "Claim transaction has been reverted with RateLimitExceeded error. Claiming will be retry later: messageHash=%s transactionHash=%s",
          message.messageHash,
          receipt.transactionHash,
        );
        return;
      }

      message.edit({ status: MessageStatus.CLAIMED_REVERTED });
      await this.databaseService.updateMessage(message);
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

    await this.databaseService.updateMessage(message);

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
