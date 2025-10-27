import { Counter, Gauge, Histogram } from "prom-client";
import { IMetricsService } from "../../core/services/IMetricsService";
import { SingletonMetricsService } from "../SingletonMetricsService";

export enum ExampleMetrics {
  ExampleMetrics = "ExampleMetrics",
}

describe("SingletonMetricsService", () => {
  let metricService: IMetricsService<ExampleMetrics>;

  beforeEach(() => {
    metricService = new SingletonMetricsService({ app: "app" });
  });

  it("should create a counter", () => {
    const counter = metricService.createCounter(ExampleMetrics.ExampleMetrics, "A test counter");
    expect(counter).toBeInstanceOf(Counter);
  });

  it("should increment a counter", async () => {
    const counter = metricService.createCounter(ExampleMetrics.ExampleMetrics, "A test counter");
    metricService.incrementCounter(ExampleMetrics.ExampleMetrics, {}, 1);
    expect((await counter.get()).values[0].value).toBe(1);
  });

  it("should create a gauge", () => {
    const gauge = metricService.createGauge(ExampleMetrics.ExampleMetrics, "A test gauge");
    expect(gauge).toBeInstanceOf(Gauge);
  });

  it("should increment a gauge", async () => {
    const gauge = metricService.createGauge(ExampleMetrics.ExampleMetrics, "A test gauge");
    metricService.incrementGauge(ExampleMetrics.ExampleMetrics, {}, 5);
    expect((await gauge.get()).values[0].value).toBe(5);
  });

  it("should decrement a gauge", async () => {
    metricService.createGauge(ExampleMetrics.ExampleMetrics, "A test gauge");
    metricService.incrementGauge(ExampleMetrics.ExampleMetrics, {}, 5);
    metricService.decrementGauge(ExampleMetrics.ExampleMetrics, {}, 2);
    expect(await metricService.getGaugeValue(ExampleMetrics.ExampleMetrics, {})).toBe(3);
  });

  it("should return the correct counter value", async () => {
    metricService.createCounter(ExampleMetrics.ExampleMetrics, "A test counter");
    metricService.incrementCounter(ExampleMetrics.ExampleMetrics, {}, 5);
    const counterValue = await metricService.getCounterValue(ExampleMetrics.ExampleMetrics, {});
    expect(counterValue).toBe(5);
  });

  it("should return the correct gauge value", async () => {
    metricService.createGauge(ExampleMetrics.ExampleMetrics, "A test gauge");
    metricService.incrementGauge(ExampleMetrics.ExampleMetrics, {}, 10);
    const gaugeValue = await metricService.getGaugeValue(ExampleMetrics.ExampleMetrics, {});
    expect(gaugeValue).toBe(10);
  });

  it("should create a histogram and add values", async () => {
    const histogram = metricService.createHistogram(
      ExampleMetrics.ExampleMetrics,
      [0.1, 0.5, 1, 2, 3, 5],
      "A test histogram",
    );
    expect(histogram).toBeInstanceOf(Histogram);
  });

  it("should add values to histogram and retrieve them", async () => {
    metricService.createHistogram(ExampleMetrics.ExampleMetrics, [0.1, 0.5, 1, 2, 3, 5], "A test histogram");
    metricService.addValueToHistogram(ExampleMetrics.ExampleMetrics, 0.3);
    metricService.addValueToHistogram(ExampleMetrics.ExampleMetrics, 1.5);
    const histogramValues = await metricService.getHistogramMetricsValues(ExampleMetrics.ExampleMetrics);
    expect(histogramValues?.values.length).toBe(9);
  });
});
