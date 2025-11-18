export interface ITransactionMetricsUpdater {
  addTransactionProcessingTime(transactionProcessingTimeInSeconds: number): void;
  addTransactionLineaInfuraLatencyTime(transactionBroadcastTimeInSeconds: number): void;
  incrementTransactionProcessedTotal(direction: string): void;
  incrementTransactionProcessingTimeSum(direction: string, transactionProcessingTimeInSeconds: number): void;
  incrementTransactionLineaInfuraLatencyTimeSum(
    direction: string,
    transactionLineaInfuraLatencyInSeconds: number,
  ): void;
}
