import type { TransactionResponse } from "../types/blockchain";

export interface IClaimRetrier {
  retryTransactionWithHigherFee(transactionHash: string, priceBumpPercent?: number): Promise<TransactionResponse>;
}
