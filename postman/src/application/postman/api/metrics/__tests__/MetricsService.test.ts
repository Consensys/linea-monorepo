import { MetricsService } from "../MetricsService";
import { Counter, Gauge, Histogram } from "prom-client";

class TestMetricService extends MetricsService {
  constructor() {
    super();
  }
}

describe("MetricsService", () => {
  let metricService: TestMetricService;

  beforeEach(() => {
    metricService = new TestMetricService();
  });

  it("should create a counter", () => {
    const counter = metricService.createCounter("test_counter", "A test counter");
    expect(counter).toBeInstanceOf(Counter);
  });

  it("should increment a counter", async () => {
    const counter = metricService.createCounter("test_counter", "A test counter");
    metricService.incrementCounter("test_counter", {}, 1);
    expect((await counter.get()).values[0].value).toBe(1);
  });

  it("should create a gauge", () => {
    const gauge = metricService.createGauge("test_gauge", "A test gauge");
    expect(gauge).toBeInstanceOf(Gauge);
  });

  it("should increment a gauge", async () => {
    const gauge = metricService.createGauge("test_gauge", "A test gauge");
    metricService.incrementGauge("test_gauge", 5);
    expect((await gauge.get()).values[0].value).toBe(5);
  });

  it("should decrement a gauge", async () => {
    metricService.createGauge("test_gauge", "A test gauge");
    metricService.incrementGauge("test_gauge", 5);
    metricService.decrementGauge("test_gauge", 2);
    expect(await metricService.getGaugeValue("test_gauge", {})).toBe(3);
  });

  it("should create a histogram", () => {
    const histogram = metricService.createHistogram("test_histogram", "A test histogram");
    expect(histogram).toBeInstanceOf(Histogram);
  });

  it("should return the correct counter value", async () => {
    metricService.createCounter("test_counter", "A test counter");
    metricService.incrementCounter("test_counter", {}, 5);
    const counterValue = await metricService.getCounterValue("test_counter", {});
    expect(counterValue).toBe(5);
  });

  it("should return the correct gauge value", async () => {
    metricService.createGauge("test_gauge", "A test gauge");
    metricService.incrementGauge("test_gauge", 10);
    const gaugeValue = await metricService.getGaugeValue("test_gauge", {});
    expect(gaugeValue).toBe(10);
  });

  it("should observe a value in a histogram", async () => {
    const histogram = metricService.createHistogram("test_histogram", "A test histogram");
    metricService.observeHistogram("test_histogram", 0);
    expect((await histogram.get()).values[0].value).toBe(1);
  });
});
