import { MINIMUM_MARGIN, PROFIT_MARGIN_MULTIPLIER } from "../../../domain/constants";
import { BaseError } from "../../../domain/errors/BaseError";
import { Message } from "../../../domain/message/Message";

import type { IL2ContractClient } from "../../../domain/ports/IL2ContractClient";
import type { IPostmanLogger } from "../../../domain/ports/ILogger";
import type { ILineaProvider } from "../../../domain/ports/IProvider";
import type {
  ITransactionValidationService,
  TransactionEvaluationResult,
  TransactionValidationServiceConfig,
} from "../../../domain/ports/ITransactionValidationService";

export class LineaTransactionValidator implements ITransactionValidationService {
  constructor(
    private readonly config: TransactionValidationServiceConfig,
    private readonly provider: ILineaProvider,
    private readonly l2MessageServiceClient: IL2ContractClient,
    private readonly logger: IPostmanLogger,
  ) {}

  public async evaluateTransaction(
    message: Message,
    feeRecipient?: string,
    claimViaAddress?: string,
  ): Promise<TransactionEvaluationResult> {
    const { gasLimit, maxPriorityFeePerGas, maxFeePerGas } = await this.l2MessageServiceClient.estimateClaimGasFees(
      { ...message, feeRecipient },
      { claimViaAddress },
    );

    this.logger.debug(
      `Estimated gas fees for message claiming. messageHash=${message.messageHash} gasLimit=${gasLimit} maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`,
    );

    const threshold = this.calculateGasEstimationThreshold(message.fee, gasLimit);
    const estimatedGasLimit = this.getGasLimit(gasLimit);
    const isUnderPriced = await this.isUnderPriced(gasLimit, message.fee, message.compressedTransactionSize!);
    const hasZeroFee = this.hasZeroFee(message);
    const isRateLimitExceeded = await this.l2MessageServiceClient.isRateLimitExceeded(message.fee, message.value);
    const isForSponsorship = this.isForSponsorship(gasLimit, hasZeroFee, isUnderPriced);

    this.logger.debug(
      `Transaction evaluation results. messageHash=${message.messageHash} hasZeroFee=${hasZeroFee} isUnderPriced=${isUnderPriced} isRateLimitExceeded=${isRateLimitExceeded} isForSponsorship=${isForSponsorship} estimatedGasLimit=${estimatedGasLimit} threshold=${threshold}`,
    );

    return {
      hasZeroFee,
      isUnderPriced,
      isRateLimitExceeded,
      isForSponsorship,
      estimatedGasLimit,
      threshold,
      maxPriorityFeePerGas,
      maxFeePerGas,
    };
  }

  private hasZeroFee(message: Message): boolean {
    return message.hasZeroFee() && this.config.profitMargin !== 0;
  }

  private async isUnderPriced(
    gasLimit: bigint,
    messageFee: bigint,
    messageCompressedTransactionSize: number,
  ): Promise<boolean> {
    const extraData = await this.provider.getBlockExtraData("latest");

    if (!extraData) {
      throw new BaseError("No extra data.");
    }

    const priorityFee =
      (BigInt(MINIMUM_MARGIN * 10) *
        ((BigInt(extraData.variableCost) * BigInt(messageCompressedTransactionSize)) / gasLimit +
          BigInt(extraData.fixedCost))) /
      10n;

    const actualCost = priorityFee * gasLimit * BigInt(Math.floor(this.config.profitMargin * PROFIT_MARGIN_MULTIPLIER));
    const maxFee = messageFee * BigInt(PROFIT_MARGIN_MULTIPLIER);
    return maxFee < actualCost;
  }

  private calculateGasEstimationThreshold(messageFee: bigint, gasLimit: bigint): number {
    return parseFloat((messageFee / gasLimit).toString());
  }

  private getGasLimit(gasLimit: bigint): bigint | null {
    return gasLimit <= this.config.maxClaimGasLimit ? gasLimit : null;
  }

  private isForSponsorship(gasLimit: bigint, hasZeroFee: boolean, isUnderPriced: boolean): boolean {
    if (!this.config.isPostmanSponsorshipEnabled) return false;
    if (gasLimit > this.config.maxPostmanSponsorGasLimit) return false;
    if (hasZeroFee) return true;
    if (isUnderPriced) return true;
    return false;
  }
}
