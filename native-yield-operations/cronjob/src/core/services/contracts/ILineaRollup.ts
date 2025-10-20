export interface ILineaRollup<TransactionReceipt> {
  transferFundsForNativeYield(amount: bigint): Promise<TransactionReceipt | null>;
}
