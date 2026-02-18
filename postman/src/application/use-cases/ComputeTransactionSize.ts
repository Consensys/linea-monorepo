import { Message } from "../../domain/message/Message";
import { MessageStatus } from "../../domain/types/enums";
import { handleUseCaseError } from "../services/handleUseCaseError";

import type { IClaimGasEstimator } from "../../domain/ports/IClaimGasEstimator";
import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { IL2ClaimTransactionSizeCalculator } from "../../domain/ports/IL2ClaimTransactionSizeCalculator";
import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { LineaGasFees } from "../../domain/types/blockchain";
import type { L2ClaimMessageTransactionSizeProcessorConfig } from "../config/PostmanConfig";

export class ComputeTransactionSize {
  constructor(
    private readonly repository: IMessageRepository,
    private readonly gasEstimator: IClaimGasEstimator,
    private readonly transactionSizeCalculator: IL2ClaimTransactionSizeCalculator,
    private readonly errorParser: IErrorParser,
    private readonly config: L2ClaimMessageTransactionSizeProcessorConfig,
    private readonly logger: ILogger,
  ) {}

  public async process(): Promise<void> {
    let message: Message | null = null;

    try {
      const messages = await this.repository.getNFirstMessagesByStatus(
        MessageStatus.ANCHORED,
        this.config.direction,
        1,
        this.config.originContractAddress,
      );

      if (messages.length === 0) {
        this.logger.info("No anchored messages found to compute transaction size.");
        return;
      }

      message = messages[0];

      const fees = (await this.gasEstimator.estimateClaimGasFees(message)) as LineaGasFees;
      const { gasLimit, maxPriorityFeePerGas, maxFeePerGas } = fees;

      const transactionSize = await this.transactionSizeCalculator.calculateTransactionSize(message, {
        maxPriorityFeePerGas,
        maxFeePerGas,
        gasLimit,
      });

      message.edit({
        claimTxGasLimit: Number(gasLimit),
        compressedTransactionSize: transactionSize,
        status: MessageStatus.TRANSACTION_SIZE_COMPUTED,
      });

      await this.repository.updateMessage(message);

      this.logger.info(
        "Message transaction size and gas limit have been computed: messageHash=%s transactionSize=%s gasLimit=%s",
        message.messageHash,
        transactionSize,
        gasLimit,
      );
    } catch (e) {
      await this.handleProcessingError(e, message);
    }
  }

  private async handleProcessingError(e: unknown, message: Message | null): Promise<void> {
    await handleUseCaseError({
      error: e,
      errorParser: this.errorParser,
      logger: this.logger,
      context: {
        operation: "ComputeTransactionSize",
        direction: this.config.direction,
        messageHash: message?.messageHash,
      },
      repository: this.repository,
      message: message ?? undefined,
    });
  }
}
