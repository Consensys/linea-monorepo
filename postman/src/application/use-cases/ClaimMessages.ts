import { Message } from "../../domain/message/Message";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../domain/types/enums";
import { handleUseCaseError } from "../services/handleUseCaseError";

import type { IClaimService } from "../../domain/ports/IClaimService";
import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { IGasProvider } from "../../domain/ports/IGasProvider";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { IMessageStatusChecker } from "../../domain/ports/IMessageStatusChecker";
import type { ITransactionValidationService } from "../../domain/ports/ITransactionValidationService";
import type { MessageClaimingProcessorConfig } from "../config/PostmanConfig";
import type { NonceCoordinator } from "../services/NonceCoordinator";

export class ClaimMessages {
  constructor(
    private readonly statusChecker: IMessageStatusChecker,
    private readonly claimService: IClaimService,
    private readonly nonceCoordinator: NonceCoordinator,
    private readonly repository: IMessageRepository,
    private readonly transactionValidationService: ITransactionValidationService,
    private readonly errorParser: IErrorParser,
    private readonly config: MessageClaimingProcessorConfig,
    private readonly logger: ILogger,
    private readonly gasProvider?: IGasProvider,
  ) {}

  public async process(): Promise<void> {
    let nextMessageToClaim: Message | null = null;

    try {
      const nonce = await this.nonceCoordinator.getNextNonce(this.config.direction);

      if (!nonce && nonce !== 0) {
        this.logger.error("Nonce returned from getNonce is an invalid value (e.g. null or undefined)");
        return;
      }

      nextMessageToClaim = await this.getNextMessageToClaim();

      if (!nextMessageToClaim) {
        this.logger.info("No message to claim found");
        return;
      }

      this.logger.info("Found message to claim: messageHash=%s", nextMessageToClaim.messageHash);

      const messageStatus = await this.statusChecker.getMessageStatus({
        messageHash: nextMessageToClaim.messageHash,
        messageBlockNumber: nextMessageToClaim.sentBlockNumber,
      });

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        this.logger.info("Found already claimed message: messageHash=%s", nextMessageToClaim.messageHash);

        nextMessageToClaim.edit({ status: MessageStatus.CLAIMED_SUCCESS });
        await this.repository.updateMessage(nextMessageToClaim);
        return;
      }

      const {
        hasZeroFee,
        isUnderPriced,
        isRateLimitExceeded,
        isForSponsorship,
        estimatedGasLimit,
        threshold,
        ...claimTxFees
      } = await this.transactionValidationService.evaluateTransaction(
        nextMessageToClaim,
        this.config.feeRecipientAddress,
        this.config.claimViaAddress,
      );

      if (!isForSponsorship && (await this.handleZeroFee(hasZeroFee, nextMessageToClaim))) return;
      if (await this.handleNonExecutable(nextMessageToClaim, estimatedGasLimit)) return;

      nextMessageToClaim.edit({ claimGasEstimationThreshold: threshold, isForSponsorship });
      await this.repository.updateMessage(nextMessageToClaim);

      if (
        !isForSponsorship &&
        (await this.handleUnderpriced(nextMessageToClaim, isUnderPriced, estimatedGasLimit, claimTxFees.maxFeePerGas))
      )
        return;
      if (this.handleRateLimitExceeded(nextMessageToClaim, isRateLimitExceeded)) return;

      await this.executeClaimTransaction(
        nextMessageToClaim,
        nonce,
        estimatedGasLimit!,
        claimTxFees.maxPriorityFeePerGas,
        claimTxFees.maxFeePerGas,
      );
    } catch (e) {
      await this.handleProcessingError(e, nextMessageToClaim);
    }
  }

  private async getNextMessageToClaim(): Promise<Message | null> {
    const { direction, originContractAddress, maxNumberOfRetries, retryDelayInSeconds, profitMargin } = this.config;

    if (direction === Direction.L1_TO_L2) {
      return this.repository.getFirstMessageToClaimOnL2(
        direction,
        originContractAddress,
        [MessageStatus.TRANSACTION_SIZE_COMPUTED, MessageStatus.FEE_UNDERPRICED],
        maxNumberOfRetries,
        retryDelayInSeconds,
      );
    }

    const { maxFeePerGas } = await this.gasProvider!.getGasFees();
    return this.repository.getFirstMessageToClaimOnL1(
      direction,
      originContractAddress,
      maxFeePerGas,
      profitMargin,
      maxNumberOfRetries,
      retryDelayInSeconds,
    );
  }

  private async executeClaimTransaction(
    message: Message,
    nonce: number,
    gasLimit: bigint,
    maxPriorityFeePerGas: bigint,
    maxFeePerGas: bigint,
  ): Promise<void> {
    const claimTxFn = async () =>
      await this.claimService.claim(
        {
          ...message,
          feeRecipient: this.config.feeRecipientAddress,
          messageBlockNumber: message.sentBlockNumber,
        },
        {
          claimViaAddress: this.config.claimViaAddress,
          overrides: { nonce, gasLimit, maxPriorityFeePerGas, maxFeePerGas },
        },
      );
    await this.repository.updateMessageWithClaimTxAtomic(message, nonce, claimTxFn);
  }

  private async handleZeroFee(hasZeroFee: boolean, message: Message): Promise<boolean> {
    if (hasZeroFee) {
      this.logger.warn(
        "Found message with zero fee. This message will not be processed: messageHash=%s",
        message.messageHash,
      );
      message.edit({ status: MessageStatus.ZERO_FEE });
      await this.repository.updateMessage(message);
      return true;
    }
    return false;
  }

  private async handleNonExecutable(message: Message, estimatedGasLimit: bigint | null): Promise<boolean> {
    if (!estimatedGasLimit) {
      this.logger.warn(
        "Estimated gas limit is higher than the max allowed gas limit for this message: messageHash=%s messageInfo=%s estimatedGasLimit=%s maxAllowedGasLimit=%s",
        message.messageHash,
        message.toString(),
        estimatedGasLimit?.toString(),
        this.config.maxClaimGasLimit.toString(),
      );
      message.edit({ status: MessageStatus.NON_EXECUTABLE });
      await this.repository.updateMessage(message);
      return true;
    }
    return false;
  }

  private async handleUnderpriced(
    message: Message,
    isUnderPriced: boolean,
    estimatedGasLimit: bigint | null,
    maxFeePerGas: bigint,
  ): Promise<boolean> {
    if (isUnderPriced) {
      if (message.status !== MessageStatus.FEE_UNDERPRICED) {
        this.logger.warn(
          "Fee underpriced found in this message: messageHash=%s messageInfo=%s transactionGasLimit=%s maxFeePerGas=%s",
          message.messageHash,
          message.toString(),
          estimatedGasLimit?.toString(),
          maxFeePerGas.toString(),
        );
        message.edit({ status: MessageStatus.FEE_UNDERPRICED });
        await this.repository.updateMessage(message);
      } else {
        this.logger.warn("Message is underpriced, will retry later: messageHash=%s", message.messageHash);
      }
      return true;
    }
    return false;
  }

  private handleRateLimitExceeded(message: Message, isRateLimitExceeded: boolean): boolean {
    if (isRateLimitExceeded) {
      this.logger.warn(
        "Rate limit exceeded for this message. It will be reprocessed later: messageHash=%s",
        message.messageHash,
      );
      return true;
    }
    return false;
  }

  private async handleProcessingError(e: unknown, message: Message | null): Promise<void> {
    await handleUseCaseError({
      error: e,
      errorParser: this.errorParser,
      logger: this.logger,
      context: { operation: "ClaimMessages", direction: this.config.direction, messageHash: message?.messageHash },
      repository: this.repository,
      message: message ?? undefined,
    });
  }
}
