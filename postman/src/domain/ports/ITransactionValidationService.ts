import { Message } from "../message/Message";

export type TransactionEvaluationResult = {
  hasZeroFee: boolean;
  isUnderPriced: boolean;
  isRateLimitExceeded: boolean;
  isForSponsorship: boolean;
  estimatedGasLimit: bigint | null;
  threshold: number;
  maxPriorityFeePerGas: bigint;
  maxFeePerGas: bigint;
};

export interface ITransactionValidationService {
  evaluateTransaction(
    message: Message,
    feeRecipient?: string,
    claimViaAddress?: string,
  ): Promise<TransactionEvaluationResult>;
}

export type TransactionValidationServiceConfig = {
  profitMargin: number;
  maxClaimGasLimit: bigint;
  isPostmanSponsorshipEnabled: boolean;
  maxPostmanSponsorGasLimit: bigint;
};
