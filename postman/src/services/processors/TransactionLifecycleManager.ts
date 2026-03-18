import { ILogger } from "@consensys/linea-shared-utils";

import { ITransactionProvider } from "../../core/clients/blockchain/IProvider";
import { Message } from "../../core/entities/Message";
import { OnChainMessageStatus, MessageStatus } from "../../core/enums";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";
import { IMessageStatusReader } from "../../core/services/contracts/IMessageServiceContract";
import { IReceiptPoller } from "../../core/services/IReceiptPoller";
import { ITransactionRetrier } from "../../core/services/ITransactionRetrier";
import {
  ITransactionLifecycleManager,
  TransactionLifecycleConfig,
} from "../../core/services/processors/ITransactionLifecycleManager";
import { TransactionReceipt } from "../../core/types";

export class TransactionLifecycleManager implements ITransactionLifecycleManager {
  constructor(
    private readonly messageServiceContract: IMessageStatusReader,
    private readonly provider: ITransactionProvider,
    private readonly transactionRetrier: ITransactionRetrier,
    private readonly receiptPoller: IReceiptPoller,
    private readonly messageRepository: IMessageRepository,
    private readonly config: TransactionLifecycleConfig,
    private readonly logger: ILogger,
  ) {}

  public async retryWithBump(message: Message): Promise<TransactionReceipt | null> {
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

  public async cancelAndResetMessage(message: Message): Promise<TransactionReceipt | null> {
    try {
      const messageStatus = await this.messageServiceContract.getMessageStatus({
        messageHash: message.messageHash,
        messageBlockNumber: message.sentBlockNumber,
      });

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        const receipt = await this.provider.getTransactionReceipt(message.claimTxHash!);
        if (receipt) {
          return receipt;
        }
        this.logger.warn("Message claimed on-chain but receipt not available, will retry later.", {
          messageHash: message.messageHash,
        });
        return null;
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

      return null;
    } catch (e) {
      this.logger.error("Failed to cancel and reset message.", {
        messageHash: message.messageHash,
        error: e,
      });
      return null;
    }
  }
}
