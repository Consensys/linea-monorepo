import { Counter, Gauge, Histogram, Registry } from "prom-client";
import { IMetricService } from "../../../../core/metrics/IMetricService";

export abstract class MetricService implements IMetricService {
  private readonly registry: Registry;
  private readonly counters: Map<string, Counter<string>>;
  private readonly gauges: Map<string, Gauge<string>>;
  private readonly histograms: Map<string, Histogram<string>>;

  constructor() {
    this.registry = new Registry();
    this.registry.setDefaultLabels({ app: "postman" });

    this.counters = new Map();
    this.gauges = new Map();
    this.histograms = new Map();
  }

  public getRegistry(): Registry {
    return this.registry;
  }

  /**
   * Creates counter metric
   */
  public createCounter(name: string, help: string, labelNames: string[] = []): Counter<string> {
    if (!this.counters.has(name)) {
      this.counters.set(
        name,
        new Counter({
          name,
          help,
          labelNames,
          registers: [this.registry],
        }),
      );
    }
    return this.counters.get(name) as Counter<string>;
  }

  public async getCounterValue(name: string, labels: Record<string, string>): Promise<number | undefined> {
    const counter = this.counters.get(name);
    if (!counter) {
      return undefined;
    }

    return (
      (await counter.get()).values.find((value) => {
        return Object.entries(labels).every(([key, val]) => value.labels[key] === val);
      })?.value || 0
    );
  }

  /**
   * Creates gauge metric
   */
  public createGauge(name: string, help: string, labelNames: string[] = []): Gauge<string> {
    if (!this.gauges.has(name)) {
      this.gauges.set(
        name,
        new Gauge({
          name,
          help,
          labelNames,
          registers: [this.registry],
        }),
      );
    }
    return this.gauges.get(name) as Gauge<string>;
  }

  public async getGaugeValue(name: string, labels: Record<string, string>): Promise<number | undefined> {
    const gauge = this.gauges.get(name);

    if (!gauge) {
      return undefined;
    }

    return (await gauge.get()).values.find((value) => {
      return Object.entries(labels).every(([key, val]) => value.labels[key] === val);
    })?.value;
  }

  /**
   * Creates histogram metric
   */
  public createHistogram(
    name: string,
    help: string,
    labelNames: string[] = [],
    buckets: number[] = [0.1, 0.5, 1, 2, 5, 10],
  ): Histogram<string> {
    if (!this.histograms.has(name)) {
      this.histograms.set(
        name,
        new Histogram({
          name,
          help,
          labelNames,
          buckets,
          registers: [this.registry],
        }),
      );
    }
    return this.histograms.get(name) as Histogram<string>;
  }

  /**
   * Increments a counter metric
   */
  public incrementCounter(name: string, labels: Record<string, string> = {}, value?: number): void {
    const counter = this.counters.get(name);
    if (counter) {
      counter.inc(labels, value);
    }
  }

  /**
   * Increment a gauge metric value
   */
  public incrementGauge(name: string, value: number, labels: Record<string, string> = {}): void {
    const gauge = this.gauges.get(name);
    if (gauge) {
      gauge.inc(labels, value);
    }
  }

  /**
   * Decrement a gauge metric value
   */
  public decrementGauge(name: string, value: number, labels: Record<string, string> = {}): void {
    const gauge = this.gauges.get(name);
    if (gauge) {
      gauge.dec(labels, value);
    }
  }

  /**
   * Records a value in a histogram metric
   */
  public observeHistogram(name: string, value: number, labels: Record<string, string> = {}): void {
    const histogram = this.histograms.get(name);
    if (histogram) {
      histogram.observe(labels, value);
    }
  }
}
