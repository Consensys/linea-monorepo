import { ILogger } from "@consensys/linea-shared-utils";

import { ITransactionProvider, IBlockProvider } from "../../core/clients/blockchain/IProvider";
import { Message } from "../../core/entities/Message";
import { Direction, OnChainMessageStatus, MessageStatus } from "../../core/enums";
import { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "../../core/metrics";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";
import { IMessageStatusReader, IRateLimitChecker } from "../../core/services/contracts/IMessageServiceContract";
import { IReceiptPoller } from "../../core/services/IReceiptPoller";
import { ITransactionRetrier } from "../../core/services/ITransactionRetrier";
import {
  IMessageClaimingPersister,
  MessageClaimingPersisterConfig,
} from "../../core/services/processors/IMessageClaimingPersister";
import { TransactionReceipt } from "../../core/types";

export class MessageClaimingPersister implements IMessageClaimingPersister {
  constructor(
    private readonly messageRepository: IMessageRepository,
    private readonly messageServiceContract: IMessageStatusReader & IRateLimitChecker,
    private readonly sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater,
    private readonly transactionMetricsUpdater: ITransactionMetricsUpdater,
    private readonly provider: ITransactionProvider & IBlockProvider,
    private readonly transactionRetrier: ITransactionRetrier,
    private readonly receiptPoller: IReceiptPoller,
    private readonly config: MessageClaimingPersisterConfig,
    private readonly logger: ILogger,
  ) {}

  private isMessageExceededSubmissionTimeout(message: Message): boolean {
    return (
      !!message.updatedAt && new Date().getTime() - message.updatedAt.getTime() > this.config.messageSubmissionTimeout
    );
  }

  public async process(): Promise<void> {
    let firstPendingMessage: Message | null = null;
    try {
      firstPendingMessage = await this.messageRepository.getFirstPendingMessage(this.config.direction);
      if (!firstPendingMessage) {
        this.logger.debug("No pending message status to update.");
        return;
      }

      if (!firstPendingMessage.claimTxHash) {
        this.logger.warn("Found pending message without claim tx hash, resetting to allow retry.", {
          messageHash: firstPendingMessage.messageHash,
        });
        firstPendingMessage.edit({ status: MessageStatus.SENT });
        await this.messageRepository.updateMessage(firstPendingMessage);
        return;
      }

      this.logger.debug("Checking pending message.", {
        messageHash: firstPendingMessage.messageHash,
        claimTxHash: firstPendingMessage.claimTxHash,
      });

      const receipt = await this.provider.getTransactionReceipt(firstPendingMessage.claimTxHash);
      if (receipt) {
        const receiptReceivedAt = new Date();
        await this.updateReceiptStatus(firstPendingMessage, receipt, receiptReceivedAt);
        return;
      }

      if (!this.isMessageExceededSubmissionTimeout(firstPendingMessage)) return;

      // Circuit breaker: max cycles reached
      if (firstPendingMessage.claimCycleCount >= this.config.maxCycles) {
        this.logger.error("Max retry cycles exceeded. Manual intervention is needed.", {
          messageHash: firstPendingMessage.messageHash,
          claimCycleCount: firstPendingMessage.claimCycleCount,
        });
        firstPendingMessage.edit({ status: MessageStatus.NEEDS_MANUAL_INTERVENTION });
        await this.messageRepository.updateMessage(firstPendingMessage);
        return;
      }

      this.logger.warn("Retrying to claim message.", { messageHash: firstPendingMessage.messageHash });

      // Check if bumps exhausted for this cycle
      if (firstPendingMessage.claimNumberOfRetry >= this.config.maxBumpsPerCycle) {
        await this.cancelAndResetMessage(firstPendingMessage);
        return;
      }

      const retryReceipt = await this.retryWithBump(firstPendingMessage);
      if (!retryReceipt) return;

      const receiptReceivedAt = new Date();
      this.logger.warn("Retried claim message transaction succeed.", {
        messageHash: firstPendingMessage.messageHash,
        transactionHash: retryReceipt.hash,
      });
      await this.updateReceiptStatus(firstPendingMessage, retryReceipt, receiptReceivedAt);
    } catch (e) {
      this.logger.error("Error processing pending message.", {
        error: e,
        ...(firstPendingMessage ? { messageHash: firstPendingMessage.messageHash } : {}),
      });
    }
  }

  private async retryWithBump(message: Message): Promise<TransactionReceipt | null> {
    try {
      const messageStatus = await this.messageServiceContract.getMessageStatus({
        messageHash: message.messageHash,
        messageBlockNumber: message.sentBlockNumber,
      });

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        const receipt = await this.provider.getTransactionReceipt(message.claimTxHash!);
        if (!receipt) {
          this.logger.warn("Message was claimed on-chain but transaction receipt is not available yet.", {
            messageHash: message.messageHash,
            transactionHash: message.claimTxHash,
          });
        }
        return receipt;
      }

      const attempt = message.claimNumberOfRetry + 1;
      this.logger.warn("Bumping fee for claim transaction.", {
        attempt: attempt.toString(),
        messageHash: message.messageHash,
      });

      const retryCreationDate = new Date();
      const tx = await this.transactionRetrier.retryWithHigherFee(message.claimTxHash!, attempt);

      message.edit({
        claimTxCreationDate: retryCreationDate,
        claimTxGasLimit: Number(tx.gasLimit),
        claimTxMaxFeePerGas: tx.maxFeePerGas ?? undefined,
        claimTxMaxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
        claimTxHash: tx.hash,
        claimNumberOfRetry: attempt,
        claimLastRetriedAt: new Date(),
        claimTxNonce: tx.nonce,
      });
      await this.messageRepository.updateMessage(message);

      return await this.receiptPoller.poll(
        tx.hash,
        this.config.receiptPollingTimeout,
        this.config.receiptPollingInterval,
      );
    } catch (e) {
      this.logger.error("Failed to retry with bumped fee.", { error: e, messageHash: message.messageHash });
      return null;
    }
  }

  private async cancelAndResetMessage(message: Message): Promise<void> {
    try {
      const messageStatus = await this.messageServiceContract.getMessageStatus({
        messageHash: message.messageHash,
        messageBlockNumber: message.sentBlockNumber,
      });

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        const receipt = await this.provider.getTransactionReceipt(message.claimTxHash!);
        if (receipt) {
          await this.updateReceiptStatus(message, receipt, new Date());
          return;
        }
        this.logger.warn("Message claimed on-chain but receipt not available, will retry later.", {
          messageHash: message.messageHash,
        });
        return;
      }

      this.logger.warn("Max fee bumps exhausted, cancelling stuck transaction and resetting message.", {
        messageHash: message.messageHash,
        claimTxNonce: message.claimTxNonce,
        claimCycleCount: message.claimCycleCount,
      });

      if (message.claimTxNonce !== undefined) {
        await this.transactionRetrier.cancelTransaction(message.claimTxNonce);
      }

      message.edit({
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
        claimCycleCount: message.claimCycleCount + 1,
        claimTxHash: undefined,
        claimTxNonce: undefined,
        claimLastRetriedAt: new Date(),
      });
      await this.messageRepository.updateMessage(message);

      this.logger.warn("Message reset to SENT for re-claiming with fresh nonce.", {
        messageHash: message.messageHash,
        newCycleCount: message.claimCycleCount,
      });
    } catch (e) {
      this.logger.error("Failed to cancel and reset message.", {
        messageHash: message.messageHash,
        error: e,
      });
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
        processingTimeInSeconds = Math.max(0, block.timestamp - message.claimTxCreationDate.getTime() / 1_000);
        infuraConfirmationTimeInSeconds = Math.max(
          0,
          (receiptReceivedAt.getTime() - message.claimTxCreationDate.getTime()) / 1_000,
        );

        this.transactionMetricsUpdater.addTransactionProcessingTime(this.config.direction, processingTimeInSeconds);
        this.transactionMetricsUpdater.addTransactionInfuraConfirmationTime(
          this.config.direction,
          infuraConfirmationTimeInSeconds,
        );
      }
    }

    if (receipt.status === "reverted") {
      const isRateLimitExceeded = await this.messageServiceContract.isRateLimitExceededError(receipt.hash);

      if (isRateLimitExceeded) {
        message.edit({
          status: MessageStatus.SENT,
        });
        await this.messageRepository.updateMessage(message);

        this.logger.warn(
          "Claim transaction has been reverted with RateLimitExceeded error. Claiming will be retry later.",
          { messageHash: message.messageHash, transactionHash: receipt.hash },
        );
        return;
      }

      message.edit({ status: MessageStatus.CLAIMED_REVERTED });
      await this.messageRepository.updateMessage(message);
      this.logger.warn("Message claim transaction has been REVERTED.", {
        messageHash: message.messageHash,
        transactionHash: receipt.hash,
        ...(processingTimeInSeconds ? { processingTimeInSeconds } : {}),
        ...(infuraConfirmationTimeInSeconds ? { infuraConfirmationTimeInSeconds } : {}),
      });
      return;
    }

    message.edit({
      status: MessageStatus.CLAIMED_SUCCESS,
    });

    await this.messageRepository.updateMessage(message);

    if (message.isForSponsorship) {
      await this.sponsorshipMetricsUpdater.incrementSponsorshipFeePaid(
        receipt.gasPrice * receipt.gasUsed,
        message.direction,
      );
    }

    this.logger.info("Message has been SUCCESSFULLY claimed.", {
      messageHash: message.messageHash,
      transactionHash: receipt.hash,
      ...(processingTimeInSeconds ? { processingTimeInSeconds } : {}),
      ...(infuraConfirmationTimeInSeconds ? { infuraConfirmationTimeInSeconds } : {}),
    });
  }
}
