import { MessageProps } from "../message/Message";

import type { MessageSent, TransactionResponse } from "../types/blockchain";

export type ClaimTransactionOverrides = {
  nonce?: number;
  gasLimit?: bigint;
  maxFeePerGas?: bigint;
  maxPriorityFeePerGas?: bigint;
};

export interface IClaimService {
  claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts?: { claimViaAddress?: string; overrides?: ClaimTransactionOverrides },
  ): Promise<TransactionResponse>;
}
