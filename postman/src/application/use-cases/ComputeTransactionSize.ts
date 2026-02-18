import { Message } from "../../domain/message/Message";
import { MessageStatus } from "../../domain/types/MessageStatus";

import type { IErrorParser } from "../../domain/ports/IErrorParser";
import type { IL2ClaimTransactionSizeCalculator } from "../../domain/ports/IL2ClaimTransactionSizeCalculator";
import type { IL2ContractClient } from "../../domain/ports/IL2ContractClient";
import type { IPostmanLogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { L2ClaimMessageTransactionSizeProcessorConfig } from "../config/PostmanConfig";

export class ComputeTransactionSize {
  constructor(
    private readonly repository: IMessageRepository,
    private readonly l2MessageServiceClient: IL2ContractClient,
    private readonly transactionSizeCalculator: IL2ClaimTransactionSizeCalculator,
    private readonly errorParser: IErrorParser,
    private readonly config: L2ClaimMessageTransactionSizeProcessorConfig,
    private readonly logger: IPostmanLogger,
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

      const { gasLimit, maxPriorityFeePerGas, maxFeePerGas } =
        await this.l2MessageServiceClient.estimateClaimGasFees(message);

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
    const parsedError = this.errorParser.parseErrorWithMitigation(e);

    if (parsedError?.mitigation && !parsedError.mitigation.shouldRetry && message) {
      message.edit({ status: MessageStatus.NON_EXECUTABLE });
      await this.repository.updateMessage(message);
      this.logger.warnOrError("Error occurred while processing message transaction size.", {
        error: e,
        parsedError,
        messageHash: message.messageHash,
      });
      return;
    }

    this.logger.warnOrError("Error occurred while processing message transaction size.", {
      error: e,
      parsedError,
    });
  }
}
