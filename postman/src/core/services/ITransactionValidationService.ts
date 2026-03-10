import { Message } from "../entities/Message";

import type { Address } from "../types/hex";

export interface ITransactionValidationService {
  evaluateTransaction(
    message: Message,
    feeRecipient?: Address,
    claimViaAddress?: Address,
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
}

export type TransactionValidationServiceConfig = {
  profitMargin: number;
  maxClaimGasLimit: bigint;
  isPostmanSponsorshipEnabled: boolean;
  maxPostmanSponsorGasLimit: bigint;
};
