import { Message } from "../../../domain/message/Message";

import type {
  ITransactionValidationService,
  TransactionEvaluationResult,
  TransactionValidationServiceConfig,
} from "../../../domain/ports/ITransactionValidationService";

export abstract class BaseTransactionValidator implements ITransactionValidationService {
  constructor(protected readonly config: TransactionValidationServiceConfig) {}

  abstract evaluateTransaction(
    message: Message,
    feeRecipient?: string,
    claimViaAddress?: string,
  ): Promise<TransactionEvaluationResult>;

  protected hasZeroFee(message: Message): boolean {
    return message.hasZeroFee() && this.config.profitMargin !== 0;
  }

  protected calculateGasEstimationThreshold(messageFee: bigint, gasLimit: bigint): number {
    return parseFloat((messageFee / gasLimit).toString());
  }

  protected getGasLimit(gasLimit: bigint): bigint | null {
    return gasLimit <= this.config.maxClaimGasLimit ? gasLimit : null;
  }

  protected isForSponsorship(gasLimit: bigint, hasZeroFee: boolean, isUnderPriced: boolean): boolean {
    if (!this.config.isPostmanSponsorshipEnabled) return false;
    if (gasLimit > this.config.maxPostmanSponsorGasLimit) return false;
    if (hasZeroFee) return true;
    if (isUnderPriced) return true;
    return false;
  }
}
