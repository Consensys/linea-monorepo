import { Message } from "../entities/Message";

export interface ITransactionValidationService {
  evaluateTransaction(
    message: Message,
    feeRecipient?: string,
  ): Promise<{
    hasZeroFee: boolean;
    isUnderPriced: boolean;
    isRateLimitExceeded: boolean;
    isForSponsorship: boolean;
    estimatedGasLimit: bigint | null;
    threshold: number;
    maxPriorityFeePerGas: bigint;
    maxFeePerGas: bigint;
  }>;

  isForSponsorship(message: Message, feeRecipient?: string): Promise<boolean>;
}

export type TransactionValidationServiceConfig = {
  profitMargin: number;
  maxClaimGasLimit: bigint;
  isPostmanSponsorshipEnabled: boolean;
  maxPostmanSponsorGasLimit: bigint;
};
