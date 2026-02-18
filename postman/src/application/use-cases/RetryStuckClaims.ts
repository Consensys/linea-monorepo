import { Message } from "../../domain/message/Message";
import { MessageStatus, OnChainMessageStatus } from "../../domain/types/enums";

import type { IClaimRetrier } from "../../domain/ports/IClaimRetrier";
import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { IMessageStatusChecker } from "../../domain/ports/IMessageStatusChecker";
import type { IProvider } from "../../domain/ports/IProvider";
import type { TransactionResponse } from "../../domain/types/blockchain";

export class RetryStuckClaims {
  private retryState: { message: Message | null; retries: number };

  constructor(
    private readonly statusChecker: IMessageStatusChecker,
    private readonly claimRetrier: IClaimRetrier,
    private readonly provider: IProvider,
    private readonly repository: IMessageRepository,
    private readonly errorParser: IErrorParser,
    private readonly logger: ILogger,
    private readonly maxTxRetries: number,
  ) {
    this.retryState = { message: null, retries: 0 };
  }

  public async retry(message: Message): Promise<TransactionResponse | null> {
    if (!this.retryState.message || this.retryState.message.messageHash !== message.messageHash) {
      this.retryState = { message, retries: 0 };
    }

    return this.executeRetry(message.claimTxHash!, message.messageHash, message.sentBlockNumber);
  }

  private async executeRetry(
    transactionHash: string,
    messageHash: string,
    messageBlockNumber: number,
  ): Promise<TransactionResponse | null> {
    try {
      const messageStatus = await this.statusChecker.getMessageStatus({
        messageHash,
        messageBlockNumber,
      });

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        const receipt = await this.provider.getTransactionReceipt(transactionHash);
        if (!receipt) {
          this.logger.warn(
            "Message claimed but receipt not yet available, will retry: messageHash=%s transactionHash=%s",
            messageHash,
            transactionHash,
          );
        }
        return receipt ? { hash: receipt.transactionHash, gasLimit: 0n, nonce: 0 } : null;
      }

      this.retryState.retries++;
      this.logger.warn("Retrying claim: messageHash=%s attempt=%s", messageHash, this.retryState.retries);

      const tx = await this.claimRetrier.retryTransactionWithHigherFee(transactionHash);

      this.retryState.message?.edit({
        claimTxCreationDate: new Date(),
        claimTxGasLimit: parseInt(tx.gasLimit.toString()),
        claimTxMaxFeePerGas: tx.maxFeePerGas ?? undefined,
        claimTxMaxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
        claimTxHash: tx.hash,
        claimNumberOfRetry: this.retryState.retries,
        claimLastRetriedAt: new Date(),
        claimTxNonce: tx.nonce,
      });
      await this.repository.updateMessage(this.retryState.message!);

      return tx;
    } catch (e) {
      const parsed = this.errorParser.parse(e);

      if (this.retryState.retries >= this.maxTxRetries && this.retryState.message) {
        this.retryState.message.edit({ status: MessageStatus.NON_EXECUTABLE });
        await this.repository.updateMessage(this.retryState.message);
        this.logger.error(
          "Max retries exceeded, message marked NON_EXECUTABLE: messageHash=%s retries=%s",
          messageHash,
          this.retryState.retries,
        );
      } else {
        this.logger[parsed.severity](
          "RetryStuckClaims failed: messageHash=%s errorCode=%s errorMessage=%s",
          messageHash,
          parsed.errorCode,
          parsed.errorMessage,
        );
      }

      return null;
    }
  }
}
