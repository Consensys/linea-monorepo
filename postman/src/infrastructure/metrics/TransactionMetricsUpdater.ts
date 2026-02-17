import { IMetricsService } from "@consensys/linea-shared-utils";

import { ITransactionMetricsUpdater, LineaPostmanMetrics } from "../../domain/ports/IMetrics";

export class TransactionMetricsUpdater implements ITransactionMetricsUpdater {
  constructor(private readonly metricsService: IMetricsService<LineaPostmanMetrics>) {
    this.metricsService.createHistogram(
      LineaPostmanMetrics.TransactionProcessingTime,
      [0.1, 0.5, 1, 2, 3, 5, 7, 10],
      "Time taken to process a transaction",
      ["direction"],
    );

    this.metricsService.createHistogram(
      LineaPostmanMetrics.TransactionInfuraConfirmationTime,
      [0.1, 0.5, 1, 2, 3, 5, 7, 10],
      "Time taken to receive the transaction receipt from the RPC provider",
      ["direction"],
    );
  }

  public addTransactionProcessingTime(direction: string, transactionProcessingTimeInSeconds: number): void {
    this.metricsService.addValueToHistogram(
      LineaPostmanMetrics.TransactionProcessingTime,
      transactionProcessingTimeInSeconds,
      { direction },
    );
  }

  public addTransactionInfuraConfirmationTime(
    direction: string,
    transactionInfuraConfirmationTimeInSeconds: number,
  ): void {
    this.metricsService.addValueToHistogram(
      LineaPostmanMetrics.TransactionInfuraConfirmationTime,
      transactionInfuraConfirmationTimeInSeconds,
      { direction },
    );
  }
}
