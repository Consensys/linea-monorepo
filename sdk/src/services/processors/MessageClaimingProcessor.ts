import { ErrorParser } from "../../utils/ErrorParser";
import {
  DEFAULT_MAX_CLAIM_GAS_LIMIT,
  DEFAULT_MAX_NONCE_DIFF,
  DEFAULT_MAX_NUMBER_OF_RETRIES,
  DEFAULT_PROFIT_MARGIN,
  DEFAULT_RETRY_DELAY_IN_SECONDS,
  PROFIT_MARGIN_MULTIPLIER,
} from "../../core/constants";
import { Overrides, TransactionResponse, ContractTransactionResponse, EthersError, TransactionReceipt } from "ethers";
import { L1NetworkConfig, L2NetworkConfig } from "../../application/postman/app/config/config";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../core/enums/MessageEnums";
import { IMessageClaimingProcessor } from "../../core/services/processors/IMessageClaimingProcessor";
import { ILogger } from "../../core/utils/logging/ILogger";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";
import { IMessageServiceContract } from "../../core/services/contracts/IMessageServiceContract";
import { IChainQuerier } from "../../core/clients/blockchain/IChainQuerier";
import { IEIP1559GasProvider } from "../../core/clients/blockchain/IEIP1559GasProvider";
import { Message } from "../../core/entities/Message";

export class MessageClaimingProcessor implements IMessageClaimingProcessor {
  private readonly maxNonceDiff: number;
  private readonly feeRecipient?: string;
  private readonly profitMargin: number;
  private readonly maxRetry: number;
  private readonly retryDelayInSeconds: number;
  private readonly maxClaimGasLimit: bigint;

  /**
   * Initializes a new instance of the `MessageClaimingProcessor`.
   *
   * @param {IMessageRepository<ContractTransactionResponse>} messageRepository - An instance of a class implementing the `IMessageRepository` interface, used for storing and retrieving message data.
   * @param {IMessageServiceContract<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse>} messageServiceContract - An instance of a class implementing the `IMessageServiceContract` interface, used to interact with the blockchain contract.
   * @param {IChainQuerier<unknown>} chainQuerier - An instance of a class implementing the `IChainQuerier` interface, used to query blockchain data.
   * @param {L1NetworkConfig | L2NetworkConfig} config - Configuration for network-specific settings, including transaction submission timeout, maximum transaction retries, and gas limit.
   * @param {Direction} direction - The direction of message flow being processed.
   * @param {string} originContractAddress - The contract address from which the messages originate.
   * @param {ILogger} logger - An instance of a class implementing the `ILogger` interface, used for logging messages.
   */
  constructor(
    private readonly messageRepository: IMessageRepository<ContractTransactionResponse>,
    private readonly messageServiceContract: IMessageServiceContract<
      Overrides,
      TransactionReceipt,
      TransactionResponse,
      ContractTransactionResponse
    > &
      IEIP1559GasProvider,
    private readonly chainQuerier: IChainQuerier<unknown>,
    config: L1NetworkConfig | L2NetworkConfig,
    private readonly direction: Direction,
    private readonly originContractAddress: string,
    private readonly logger: ILogger,
  ) {
    this.maxNonceDiff = Math.max(config.claiming.maxNonceDiff ?? DEFAULT_MAX_NONCE_DIFF, 0);
    this.feeRecipient = config.claiming.feeRecipientAddress;
    this.profitMargin = config.claiming.profitMargin ?? DEFAULT_PROFIT_MARGIN;
    this.maxRetry = config.claiming.maxNumberOfRetries ?? DEFAULT_MAX_NUMBER_OF_RETRIES;
    this.retryDelayInSeconds = config.claiming.retryDelayInSeconds ?? DEFAULT_RETRY_DELAY_IN_SECONDS;
    this.maxClaimGasLimit = BigInt(config.claiming.maxClaimGasLimit ?? DEFAULT_MAX_CLAIM_GAS_LIMIT);
  }

  /**
   * Identifies the next message eligible for claiming and attempts to execute the claim transaction. It considers various factors such as gas estimation, profit margin, and rate limits to decide whether to proceed with the claim.
   */
  public async getAndClaimAnchoredMessage() {
    let nextMessageToClaim: Message | null = null;

    try {
      const nonce = await this.getNonce();

      if (!nonce && nonce !== 0) {
        this.logger.error("Nonce returned from getNonce is an invalid value (e.g. null or undefined)");
        return;
      }

      const { maxFeePerGas } = await this.messageServiceContract.get1559Fees();
      nextMessageToClaim = await this.messageRepository.getFirstMessageToClaim(
        this.direction,
        this.originContractAddress,
        maxFeePerGas,
        this.profitMargin,
        this.maxRetry,
        this.retryDelayInSeconds,
      );

      if (!nextMessageToClaim) {
        return;
      }

      if (BigInt(nextMessageToClaim.fee) === 0n && this.profitMargin !== 0) {
        this.logger.warn(
          "Found message with zero fee. This message will not be processed: messageHash=%s",
          nextMessageToClaim.messageHash,
        );

        nextMessageToClaim.edit({ status: MessageStatus.ZERO_FEE });
        await this.messageRepository.updateMessage(nextMessageToClaim);
        return;
      }

      const messageStatus = await this.messageServiceContract.getMessageStatus(nextMessageToClaim.messageHash);

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        this.logger.info("Found already claimed message: messageHash=%s", nextMessageToClaim.messageHash);

        nextMessageToClaim.edit({ status: MessageStatus.CLAIMED_SUCCESS });
        await this.messageRepository.updateMessage(nextMessageToClaim);
        return;
      }

      const { estimatedGasLimit, threshold } = await this.calculateGasEstimationAndThresHold(nextMessageToClaim);
      const txGasLimit = this.getGasLimit(estimatedGasLimit);

      if (!txGasLimit) {
        this.logger.warn(
          "Estimated gas limit is higher than the max allowed gas limit for this message: messageHash=%s messageInfo=%s estimatedGasLimit=%s maxAllowedGasLimit=%s",
          nextMessageToClaim.messageHash,
          nextMessageToClaim.toString(),
          estimatedGasLimit.toString(),
          this.maxClaimGasLimit.toString(),
        );
        nextMessageToClaim.edit({ status: MessageStatus.NON_EXECUTABLE });
        await this.messageRepository.updateMessage(nextMessageToClaim);
        return;
      }

      nextMessageToClaim.edit({ claimGasEstimationThreshold: threshold });
      await this.messageRepository.updateMessage(nextMessageToClaim);

      const isTxUnderPriced = await this.isTransactionUnderPriced(txGasLimit, nextMessageToClaim.fee, maxFeePerGas);

      if (isTxUnderPriced) {
        this.logger.warn(
          "Fee underpriced found in this message: messageHash=%s messageInfo=%s transactionGasLimit=%s maxFeePerGas=%s",
          nextMessageToClaim.messageHash,
          nextMessageToClaim.toString(),
          txGasLimit.toString(),
          maxFeePerGas.toString(),
        );
        nextMessageToClaim.edit({ status: MessageStatus.FEE_UNDERPRICED });
        await this.messageRepository.updateMessage(nextMessageToClaim);
        return;
      }

      if (await this.messageServiceContract.isRateLimitExceeded(nextMessageToClaim.fee, nextMessageToClaim.value)) {
        this.logger.warn(
          "Rate limit exceeded for this message. It will be reprocessed later: messageHash=%s",
          nextMessageToClaim.messageHash,
        );
        return;
      }

      await this.executeClaimTransaction(nextMessageToClaim, nonce, txGasLimit);
    } catch (e) {
      const parsedError = ErrorParser.parseErrorWithMitigation(e as EthersError);
      if (parsedError?.mitigation && !parsedError.mitigation.shouldRetry) {
        if (nextMessageToClaim) {
          nextMessageToClaim.edit({ status: MessageStatus.NON_EXECUTABLE });
          await this.messageRepository.updateMessage(nextMessageToClaim);
        }
      }
      this.logger.warnOrError(e, {
        parsedError,
      });
    }
  }

  /**
   * Retrieves the current nonce for the claiming transactions, ensuring it is within an acceptable range compared to the last recorded nonce.
   *
   * @returns {Promise<number | null>} The nonce to use for the next transaction, or null if the nonce difference exceeds the configured maximum.
   */
  private async getNonce(): Promise<number | null> {
    const lastTxNonce = await this.messageRepository.getLastClaimTxNonce(this.direction);

    let nonce = await this.chainQuerier.getCurrentNonce();
    if (lastTxNonce) {
      if (lastTxNonce - nonce > this.maxNonceDiff) {
        this.logger.warn(
          "Last recorded nonce in db is higher than the latest nonce from blockchain and exceeds the diff limit, paused the claim message process now: nonceInDb=%s nonceOnChain=%s maxAllowedNonceDiff=%s",
          lastTxNonce,
          nonce,
          this.maxNonceDiff,
        );
        return null;
      }
      nonce = Math.max(nonce, lastTxNonce + 1);
    }
    return nonce;
  }

  /**
   * Calculates the gas estimation for claiming a message and determines if the message fee meets the profit margin threshold.
   *
   * @param {Message} message - The message object to calculate gas estimation for.
   * @returns {Promise<{ threshold: number; estimatedGasLimit: bigint }>} The gas estimation and profit margin threshold for the message.
   */
  private async calculateGasEstimationAndThresHold(
    message: Message,
  ): Promise<{ threshold: number; estimatedGasLimit: bigint }> {
    const gasEstimation = await this.messageServiceContract.estimateClaimGas(
      {
        ...message,
        feeRecipient: this.feeRecipient,
      },
      { ...(await this.messageServiceContract.get1559Fees()) },
    );

    return {
      threshold: parseFloat((BigInt(message.fee) / gasEstimation).toString()),
      estimatedGasLimit: gasEstimation,
    };
  }

  /**
   * Determines the appropriate gas limit for a claim transaction based on the estimated gas and the configured maximum gas limit.
   *
   * @param {bigint} gasLimit - The estimated gas limit for the transaction.
   * @returns {bigint | null} The gas limit to use for the transaction, or null if it exceeds the maximum allowed gas limit.
   */
  private getGasLimit(gasLimit: bigint): bigint | null {
    if (gasLimit <= this.maxClaimGasLimit) {
      return gasLimit;
    } else {
      return null;
    }
  }

  /**
   * Checks if a transaction is underpriced based on the gas limit, message fee, and maximum fee per gas.
   *
   * @param {bigint} gasLimit - The gas limit for the transaction.
   * @param {bigint} messageFee - The fee associated with the message.
   * @param {bigint} maxFeePerGas - The maximum fee per gas willing to be paid.
   * @returns {boolean} `true` if the transaction is underpriced, `false` otherwise.
   */
  private isTransactionUnderPriced(gasLimit: bigint, messageFee: bigint, maxFeePerGas: bigint): boolean {
    if (
      gasLimit * maxFeePerGas * BigInt(Math.floor(this.profitMargin * PROFIT_MARGIN_MULTIPLIER)) >
      messageFee * BigInt(PROFIT_MARGIN_MULTIPLIER)
    ) {
      return true;
    }
    return false;
  }

  /**
   * Executes the claim transaction for a message, updating the message repository with transaction details.
   *
   * @param {Message} message - The message object to claim.
   * @param {number} nonce - The nonce to use for the transaction.
   * @param {bigint} gasLimit - The gas limit for the transaction.
   */
  private async executeClaimTransaction(message: Message, nonce: number, gasLimit: bigint) {
    const claimTxResponsePromise = this.messageServiceContract.claim(
      {
        ...message,
        feeRecipient: this.feeRecipient,
      },
      { nonce, gasLimit, ...(await this.messageServiceContract.get1559Fees()) },
    );
    await this.messageRepository.updateMessageWithClaimTxAtomic(message, nonce, claimTxResponsePromise);
  }
}
