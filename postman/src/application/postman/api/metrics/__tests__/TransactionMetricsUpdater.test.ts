import { PostmanMetricsService } from "../PostmanMetricsService";
import { ITransactionMetricsUpdater, LineaPostmanMetrics } from "../../../../../core/metrics";
import { TransactionMetricsUpdater } from "../TransactionMetricsUpdater";
import { IMetricsService } from "@consensys/linea-shared-utils";

describe("TransactionMetricsUpdater", () => {
  let transactionMetricsUpdater: ITransactionMetricsUpdater;
  let metricsService: IMetricsService<LineaPostmanMetrics>;

  beforeEach(() => {
    metricsService = new PostmanMetricsService();
    transactionMetricsUpdater = new TransactionMetricsUpdater(metricsService);
  });

  it("should get correct values after add histogram value", async () => {
    transactionMetricsUpdater.addTransactionProcessingTime(2);
    transactionMetricsUpdater.addTransactionProcessingTime(3);

    const histogramValues = await metricsService.getHistogramMetricsValues(
      LineaPostmanMetrics.TransactionProcessingTime,
    );
    expect(histogramValues?.values.length).toBe(11);
  });
});
