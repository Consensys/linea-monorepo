export interface ITransactionMetricsUpdater {
  addTransactionProcessingTime(direction: string, transactionProcessingTimeInSeconds: number): void;
  addTransactionLineaInfuraLatencyTime(direction: string, transactionBroadcastTimeInSeconds: number): void;
  incrementTransactionProcessedTotal(direction: string): void;
  incrementTransactionProcessingTimeSum(direction: string, transactionProcessingTimeInSeconds: number): void;
  incrementTransactionLineaInfuraLatencyTimeSum(
    direction: string,
    transactionLineaInfuraLatencyInSeconds: number,
  ): void;
}
