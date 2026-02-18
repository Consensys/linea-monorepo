import { BaseTransactionValidator } from "./BaseTransactionValidator";
import { PROFIT_MARGIN_MULTIPLIER } from "../../../domain/constants";
import { Message } from "../../../domain/message/Message";

import type { IEthereumGasProvider } from "../../../domain/ports/IGasProvider";
import type { IL1ContractClient } from "../../../domain/ports/IL1ContractClient";
import type { IPostmanLogger } from "../../../domain/ports/ILogger";
import type {
  TransactionEvaluationResult,
  TransactionValidationServiceConfig,
} from "../../../domain/ports/ITransactionValidationService";

export class EthereumTransactionValidator extends BaseTransactionValidator {
  constructor(
    private readonly lineaRollupClient: IL1ContractClient,
    private readonly gasProvider: IEthereumGasProvider,
    config: TransactionValidationServiceConfig,
    private readonly logger: IPostmanLogger,
  ) {
    super(config);
  }

  public async evaluateTransaction(
    message: Message,
    feeRecipient?: string,
    claimViaAddress?: string,
  ): Promise<TransactionEvaluationResult> {
    const [gasLimit, { maxPriorityFeePerGas, maxFeePerGas }] = await Promise.all([
      this.lineaRollupClient.estimateClaimGas(
        {
          ...message,
          feeRecipient,
          messageBlockNumber: message.sentBlockNumber,
        },
        { claimViaAddress },
      ),
      this.gasProvider.getGasFees(),
    ]);

    this.logger.debug(
      `Estimated gas fees for message claiming. messageHash=${message.messageHash} gasLimit=${gasLimit} maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`,
    );

    const threshold = this.calculateGasEstimationThreshold(message.fee, gasLimit);
    const estimatedGasLimit = this.getGasLimit(gasLimit);
    const isUnderPriced = this.isUnderPricedL1(gasLimit, message.fee, maxFeePerGas);
    const hasZeroFee = this.hasZeroFee(message);
    const isRateLimitExceeded = await this.lineaRollupClient.isRateLimitExceeded(message.fee, message.value);
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
