import { DEFAULT_MAX_TX_RETRIES } from "../../core/constants";
import { Overrides, TransactionResponse, ContractTransactionResponse, TransactionReceipt } from "ethers";
import { L1NetworkConfig, L2NetworkConfig } from "../../application/postman/app/config/config";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../core/enums/MessageEnums";
import { ILogger } from "../../core/utils/logging/ILogger";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";
import { IMessageServiceContract } from "../../core/services/contracts/IMessageServiceContract";
import { IChainQuerier } from "../../core/clients/blockchain/IChainQuerier";
import { IEIP1559GasProvider } from "../../core/clients/blockchain/IEIP1559GasProvider";
import { Message } from "../../core/entities/Message";
import { IMessageClaimingPersister } from "../../core/services/processors/IMessageClaimingPersister";
import { BaseError } from "../../core/errors/Base";

export class MessageClaimingPersister implements IMessageClaimingPersister {
  private readonly submissionTimeout: number;
  private readonly maxTxRetries: number;
  private messageBeingRetry: { message: Message | null; retries: number };

  /**
   * Initializes a new instance of the `MessageClaimingPersister`.
   *
   * @param {IMessageRepository<unknown>} messageRepository -  An instance of a class implementing the `IMessageRepository` interface, used for storing and retrieving message data.
   * @param {IMessageServiceContract<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse>} messageServiceContract - An instance of a class implementing the `IMessageServiceContract` interface, used to interact with the blockchain contract.
   * @param {IChainQuerier<TransactionReceipt>} chainQuerier - An instance of a class implementing the `IChainQuerier` interface, used to query blockchain data.
   * @param {L1NetworkConfig | L2NetworkConfig} config - Configuration for network-specific settings, including transaction submission timeout and maximum transaction retries.
   * @param {Direction} direction - The direction of message flow being processed.
   * @param {ILogger} logger - An instance of a class implementing the `ILogger` interface, used for logging messages.
   */
  constructor(
    private readonly messageRepository: IMessageRepository<unknown>,
    private readonly messageServiceContract: IMessageServiceContract<
      Overrides,
      TransactionReceipt,
      TransactionResponse,
      ContractTransactionResponse
    > &
      IEIP1559GasProvider,
    private readonly chainQuerier: IChainQuerier<TransactionReceipt>,
    config: L1NetworkConfig | L2NetworkConfig,
    private readonly direction: Direction,
    private readonly logger: ILogger,
  ) {
    this.submissionTimeout = config.claiming.messageSubmissionTimeout;
    this.maxTxRetries = config.claiming.maxTxRetries ?? DEFAULT_MAX_TX_RETRIES;
    this.messageBeingRetry = { message: null, retries: 0 };
  }

  /**
   * Determines whether a message has exceeded the configured submission timeout.
   *
   * This method checks if the time elapsed since the last update of the message exceeds the submission timeout threshold. This is useful for identifying messages that may require action due to prolonged processing times, such as retrying the transaction with a higher fee.
   *
   * @param {Message} message - The message object to check for submission timeout exceedance.
   * @returns {boolean} `true` if the message has exceeded the submission timeout, `false` otherwise.
   */
  private isMessageExceededSubmissionTimeout(message: Message): boolean {
    return !!message.updatedAt && new Date().getTime() - message.updatedAt.getTime() > this.submissionTimeout;
  }

  /**
   * Processes the first pending message, updating its status based on the transaction receipt. If the transaction has not been mined or has failed, it attempts to retry the transaction with a higher fee.
   */
  public async updateAndPersistPendingMessage() {
    let firstPendingMessage: Message | null = null;
    try {
      firstPendingMessage = await this.messageRepository.getFirstPendingMessage(this.direction);
      if (!firstPendingMessage?.claimTxHash) {
        return;
      }

      const receipt = await this.chainQuerier.getTransactionReceipt(firstPendingMessage.claimTxHash);
      if (!receipt) {
        if (this.isMessageExceededSubmissionTimeout(firstPendingMessage)) {
          this.logger.warn("Retrying to claim message: messageHash=%s", firstPendingMessage.messageHash);

          if (
            !this.messageBeingRetry.message ||
            this.messageBeingRetry.message.messageHash !== firstPendingMessage.messageHash
          ) {
            this.messageBeingRetry = { message: firstPendingMessage, retries: 0 };
          }

          const transactionReceipt = await this.retryTransaction(
            firstPendingMessage.claimTxHash,
            firstPendingMessage.messageHash,
          );
          if (transactionReceipt) {
            this.logger.warn(
              "Retried claim message transaction succeed: messageHash=%s transactionHash=%s",
              firstPendingMessage.messageHash,
              transactionReceipt.hash,
            );
          }
        }
        return;
      }

      await this.updateReceiptStatus(firstPendingMessage, receipt);
    } catch (e) {
      this.logger.error(e);
    }
  }

  /**
   * Attempts to retry a transaction with a higher fee if the original transaction has not been successfully processed.
   *
   * @param {string} transactionHash - The hash of the original transaction to retry.
   * @param {string} messageHash - The hash of the message associated with the transaction.
   * @returns {Promise<TransactionReceipt | null>} The receipt of the retried transaction, or null if the retry was unsuccessful.
   */
  private async retryTransaction(transactionHash: string, messageHash: string): Promise<TransactionReceipt | null> {
    try {
      const messageStatus = await this.messageServiceContract.getMessageStatus(messageHash, {
        blockTag: "latest",
      });

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        const receipt = await this.chainQuerier.getTransactionReceipt(transactionHash);
        if (!receipt) {
          this.logger.warn(
            "Calling retryTransaction again as message was claimed but transaction receipt is not available yet: messageHash=%s transactionHash=%s",
            messageHash,
            transactionHash,
          );
        }
        return receipt;
      }

      this.messageBeingRetry.retries++;
      this.logger.warn(
        "Retry to claim message: numberOfRetries=%s messageInfo=%s",
        this.messageBeingRetry.retries.toString(),
        this.messageBeingRetry.message?.toString(),
      );

      const tx = await this.messageServiceContract.retryTransactionWithHigherFee(transactionHash);

      const receipt = await tx.wait();
      if (!receipt) {
        throw new BaseError(
          `RetryTransaction: Transaction receipt not found after retry transaction. transactionHash=${tx.hash}`,
        );
      }

      this.messageBeingRetry.message?.edit({
        claimTxGasLimit: parseInt(tx.gasLimit.toString()),
        claimTxMaxFeePerGas: tx.maxFeePerGas ?? undefined,
        claimTxMaxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
        claimTxHash: tx.hash,
        claimNumberOfRetry: this.messageBeingRetry.retries,
        claimLastRetriedAt: new Date(),
        claimTxNonce: tx.nonce,
      });
      await this.messageRepository.updateMessage(this.messageBeingRetry.message!);

      return receipt;
    } catch (e) {
      this.logger.error(
        "Transaction retry failed: messageHash=%s error=%s",
        this.messageBeingRetry.message?.messageHash,
        e,
      );
      if (this.messageBeingRetry.retries > this.maxTxRetries) {
        this.logger.error(
          "Max number of retries exceeded. Manual intervention is needed as soon as possible: messageInfo=%s",
          this.messageBeingRetry.message?.toString(),
        );
      }
      return null;
    }
  }

  /**
   * Updates the status of a message based on the outcome of its claim transaction.
   *
   * @param {Message} message - The message object to update.
   * @param {TransactionReceipt} receipt - The receipt of the claim transaction.
   */
  private async updateReceiptStatus(message: Message, receipt: TransactionReceipt): Promise<void> {
    if (receipt.status === 0) {
      const isRateLimitExceeded = await this.messageServiceContract.isRateLimitExceededError(receipt.hash);

      if (isRateLimitExceeded) {
        message.edit({
          status: MessageStatus.SENT,
          //claimGasEstimationThreshold: undefined,
        });
        await this.messageRepository.updateMessage(message);

        this.logger.warn(
          "Claim transaction has been reverted with RateLimitExceeded error. Claiming will be retry later: messageHash=%s transactionHash=%s",
          message.messageHash,
          receipt.hash,
        );
        return;
      }

      message.edit({ status: MessageStatus.CLAIMED_REVERTED });
      await this.messageRepository.updateMessage(message);
      this.logger.warn(
        "Message claim transaction has been REVERTED: messageHash=%s transactionHash=%s",
        message.messageHash,
        receipt.hash,
      );
      return;
    }

    message.edit({ status: MessageStatus.CLAIMED_SUCCESS });
    await this.messageRepository.updateMessage(message);
    this.logger.info(
      "Message has been SUCCESSFULLY claimed: messageHash=%s transactionHash=%s",
      message.messageHash,
      receipt.hash,
    );
  }
}
