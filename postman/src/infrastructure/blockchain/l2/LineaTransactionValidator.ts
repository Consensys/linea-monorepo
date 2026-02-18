import { MINIMUM_MARGIN, PROFIT_MARGIN_MULTIPLIER } from "../../../domain/constants";
import { BaseError } from "../../../domain/errors/BaseError";
import { Message } from "../../../domain/message/Message";
import { BaseTransactionValidator } from "../shared/BaseTransactionValidator";

import type { IClaimGasEstimator } from "../../../domain/ports/IClaimGasEstimator";
import type { ILogger } from "../../../domain/ports/ILogger";
import type { ILineaProvider } from "../../../domain/ports/IProvider";
import type { IRateLimitChecker } from "../../../domain/ports/IRateLimitChecker";
import type {
  TransactionEvaluationResult,
  TransactionValidationServiceConfig,
} from "../../../domain/ports/ITransactionValidationService";
import type { LineaGasFees } from "../../../domain/types/blockchain";

export class LineaTransactionValidator extends BaseTransactionValidator {
  constructor(
    config: TransactionValidationServiceConfig,
    private readonly provider: ILineaProvider,
    private readonly gasEstimator: IClaimGasEstimator,
    private readonly rateLimitChecker: IRateLimitChecker,
    private readonly logger: ILogger,
  ) {
    super(config);
  }

  public async evaluateTransaction(
    message: Message,
    feeRecipient?: string,
    claimViaAddress?: string,
  ): Promise<TransactionEvaluationResult> {
    const fees = (await this.gasEstimator.estimateClaimGasFees(
      { ...message, feeRecipient },
      { claimViaAddress },
    )) as LineaGasFees;
    const { gasLimit, maxPriorityFeePerGas, maxFeePerGas } = fees;

    this.logger.debug(
      `Estimated gas fees for message claiming. messageHash=${message.messageHash} gasLimit=${gasLimit} maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`,
    );

    const threshold = this.calculateGasEstimationThreshold(message.fee, gasLimit);
    const estimatedGasLimit = this.getGasLimit(gasLimit);
    const isUnderPriced = await this.isUnderPricedL2(gasLimit, message.fee, message.compressedTransactionSize!);
    const hasZeroFee = this.hasZeroFee(message);
    const isRateLimitExceeded = await this.rateLimitChecker.isRateLimitExceeded(message.fee, message.value);
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

  private async isUnderPricedL2(
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
}
