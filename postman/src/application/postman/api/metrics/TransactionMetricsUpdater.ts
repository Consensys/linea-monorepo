import { IMetricsService, LineaPostmanMetrics, ITransactionMetricsUpdater } from "../../../../core/metrics";

export class TransactionMetricsUpdater implements ITransactionMetricsUpdater {
  constructor(private readonly metricsService: IMetricsService) {
    this.metricsService.createHistogram(
      LineaPostmanMetrics.TransactionProcessingTime,
      [0.1, 0.5, 1, 2, 3, 5, 7, 10],
      "Time taken to process a transaction",
      ["direction"],
    );

    this.metricsService.createHistogram(
      LineaPostmanMetrics.TransactionLineaInfuraLatencyTime,
      [0.1, 0.5, 1, 2, 3, 5, 7, 10],
      "Time taken to receive the transaction receipt from Infura",
      ["direction"],
    );

    this.metricsService.createCounter(
      LineaPostmanMetrics.TransactionProcessedTotal,
      "Number of transactions that have been processed",
      ["direction"],
    );

    this.metricsService.createCounter(
      LineaPostmanMetrics.TransactionLineaInfuraLatencyTimeSum,
      "Total number of seconds taken to receive the transaction receipt from Infura",
      ["direction"],
    );

    this.metricsService.createCounter(
      LineaPostmanMetrics.TransactionProcessingTimeSum,
      "Total number of seconds taken between transaction creation and receipt timestamp",
      ["direction"],
    );
  }

  public addTransactionProcessingTime(direction: string, transactionProcessingTimeInSeconds: number): void {
    return this.metricsService.addValueToHistogram(
      LineaPostmanMetrics.TransactionProcessingTime,
      transactionProcessingTimeInSeconds,
      { direction },
    );
  }

  public addTransactionLineaInfuraLatencyTime(direction: string, transactionLineaInfuraLatencyInSeconds: number): void {
    return this.metricsService.addValueToHistogram(
      LineaPostmanMetrics.TransactionLineaInfuraLatencyTime,
      transactionLineaInfuraLatencyInSeconds,
      { direction },
    );
  }

  public incrementTransactionProcessedTotal(direction: string): void {
    this.metricsService.incrementCounter(LineaPostmanMetrics.TransactionProcessedTotal, { direction }, 1);
  }

  public incrementTransactionProcessingTimeSum(direction: string, transactionProcessingTimeInSeconds: number): void {
    this.metricsService.incrementCounter(
      LineaPostmanMetrics.TransactionProcessingTimeSum,
      { direction },
      transactionProcessingTimeInSeconds,
    );
  }

  public incrementTransactionLineaInfuraLatencyTimeSum(
    direction: string,
    transactionLineaInfuraLatencyInSeconds: number,
  ): void {
    this.metricsService.incrementCounter(
      LineaPostmanMetrics.TransactionLineaInfuraLatencyTimeSum,
      { direction },
      transactionLineaInfuraLatencyInSeconds,
    );
  }
}
