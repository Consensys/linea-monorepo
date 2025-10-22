export interface ILineaRollupYieldExtension<TransactionReceipt> {
  transferFundsForNativeYield(amount: bigint): Promise<TransactionReceipt>;
}
