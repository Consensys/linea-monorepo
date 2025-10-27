export interface IBlockchainClientMetricsUpdater {
  addTransactionFees(amountGwei: number): void;
}
