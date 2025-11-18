import { SingletonMetricsService } from "../SingletonMetricsService";
import { ITransactionMetricsUpdater, LineaPostmanMetrics } from "../../../../../core/metrics";
import { TransactionMetricsUpdater } from "../TransactionMetricsUpdater";

describe("TransactionMetricsUpdater", () => {
  let transactionMetricsUpdater: ITransactionMetricsUpdater;
  let metricsService: SingletonMetricsService;

  beforeEach(() => {
    metricsService = new SingletonMetricsService();
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

  it("should increment counters correctly", async () => {
    transactionMetricsUpdater.incrementTransactionProcessedTotal("L1_TO_L2");
    transactionMetricsUpdater.incrementTransactionProcessedTotal("L1_TO_L2");
    transactionMetricsUpdater.incrementTransactionProcessingTimeSum("L1_TO_L2", 5);
    transactionMetricsUpdater.incrementTransactionLineaInfuraLatencyTimeSum("L1_TO_L2", 3);

    const processedTotal = await metricsService.getCounterValue(LineaPostmanMetrics.TransactionProcessedTotal, {
      direction: "L1_TO_L2",
    });
    expect(processedTotal).toBe(2);

    const processingTimeSum = await metricsService.getCounterValue(LineaPostmanMetrics.TransactionProcessingTimeSum, {
      direction: "L1_TO_L2",
    });
    expect(processingTimeSum).toBe(5);

    const lineaInfuraLatencyTimeSum = await metricsService.getCounterValue(
      LineaPostmanMetrics.TransactionLineaInfuraLatencyTimeSum,
      { direction: "L1_TO_L2" },
    );
    expect(lineaInfuraLatencyTimeSum).toBe(3);
  });
});
