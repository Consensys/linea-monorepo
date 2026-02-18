import { PROFIT_MARGIN_MULTIPLIER } from "../../../domain/constants";
import { Message } from "../../../domain/message/Message";
import { BaseTransactionValidator } from "../shared/BaseTransactionValidator";

import type { IClaimGasEstimator } from "../../../domain/ports/IClaimGasEstimator";
import type { IGasProvider } from "../../../domain/ports/IGasProvider";
import type { ILogger } from "../../../domain/ports/ILogger";
import type { IRateLimitChecker } from "../../../domain/ports/IRateLimitChecker";
import type {
  TransactionEvaluationResult,
  TransactionValidationServiceConfig,
} from "../../../domain/ports/ITransactionValidationService";
import type { LineaGasFees } from "../../../domain/types/blockchain";

export class EthereumTransactionValidator extends BaseTransactionValidator {
  constructor(
    private readonly gasEstimator: IClaimGasEstimator,
    private readonly rateLimitChecker: IRateLimitChecker,
    private readonly gasProvider: IGasProvider,
    config: TransactionValidationServiceConfig,
    private readonly logger: ILogger,
  ) {
    super(config);
  }

  public async evaluateTransaction(
    message: Message,
    feeRecipient?: string,
    claimViaAddress?: string,
  ): Promise<TransactionEvaluationResult> {
    const [estimatedFees, gasFees] = await Promise.all([
      this.gasEstimator.estimateClaimGasFees(
        {
          ...message,
          feeRecipient,
          messageBlockNumber: message.sentBlockNumber,
        },
        { claimViaAddress },
      ),
      this.gasProvider.getGasFees(),
    ]);
    const { gasLimit } = estimatedFees as LineaGasFees;
    const { maxPriorityFeePerGas, maxFeePerGas } = gasFees;

    this.logger.debug(
      `Estimated gas fees for message claiming. messageHash=${message.messageHash} gasLimit=${gasLimit} maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`,
    );

    const threshold = this.calculateGasEstimationThreshold(message.fee, gasLimit);
    const estimatedGasLimit = this.getGasLimit(gasLimit);
    const isUnderPriced = this.isUnderPricedL1(gasLimit, message.fee, maxFeePerGas);
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

  private isUnderPricedL1(gasLimit: bigint, messageFee: bigint, maxFeePerGas: bigint): boolean {
    const actualCost =
      gasLimit * maxFeePerGas * BigInt(Math.floor(this.config.profitMargin * PROFIT_MARGIN_MULTIPLIER));
    const maxFee = messageFee * BigInt(PROFIT_MARGIN_MULTIPLIER);
    return actualCost > maxFee;
  }
}
