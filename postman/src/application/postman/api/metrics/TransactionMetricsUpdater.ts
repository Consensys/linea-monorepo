import { IMetricsService, LineaPostmanMetrics, ITransactionMetricsUpdater } from "../../../../core/metrics";

export class TransactionMetricsUpdater implements ITransactionMetricsUpdater {
  constructor(private readonly metricsService: IMetricsService) {
    this.metricsService.createHistogram(
      LineaPostmanMetrics.TransactionProcessingTime,
      [0.1, 0.5, 1, 2, 3, 5, 7, 10],
      "Time taken to process a transaction",
    );

    this.metricsService.createHistogram(
      LineaPostmanMetrics.TransactionBroadcastTime,
      [0.1, 0.5, 1, 2, 3, 5, 7, 10],
      "Time taken to broadcast a transaction",
    );

    this.metricsService.createHistogram(
      LineaPostmanMetrics.TransactionInclusionTime,
      [0.1, 0.5, 1, 2, 3, 5, 7, 10, 15, 20, 30, 45, 60],
      "Time taken for a transaction to be included in a block",
    );
  }

  public addTransactionProcessingTime(transactionProcessingTimeInSeconds: number): void {
    return this.metricsService.addValueToHistogram(
      LineaPostmanMetrics.TransactionProcessingTime,
      transactionProcessingTimeInSeconds,
    );
  }

  public addTransactionBroadcastTime(transactionBroadcastTimeInSeconds: number): void {
    return this.metricsService.addValueToHistogram(
      LineaPostmanMetrics.TransactionBroadcastTime,
      transactionBroadcastTimeInSeconds,
    );
  }

  public addTransactionInclusionTime(transactionInclusionTimeInSeconds: number): void {
    return this.metricsService.addValueToHistogram(
      LineaPostmanMetrics.TransactionInclusionTime,
      transactionInclusionTimeInSeconds,
    );
  }
}
