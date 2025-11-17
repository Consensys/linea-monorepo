export interface ITransactionMetricsUpdater {
  addTransactionProcessingTime(transactionProcessingTimeInSeconds: number): void;
  addTransactionBroadcastTime(transactionBroadcastTimeInSeconds: number): void;
  addTransactionInclusionTime(transactionInclusionTimeInSeconds: number): void;
}
