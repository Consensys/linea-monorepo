import { Counter, Gauge, MetricObjectWithValues, MetricValue, Registry } from "prom-client";
import { IMetricsService, LineaPostmanMetrics } from "../../../../core/metrics/IMetricsService";

/**
 * Take care to instantiate as a singleton because there should be only be one instance of prom-client Registry
 * TODO - Implement Singleton pattern for this class
 *
 * MetricsService class that implements the IMetricsService interface.
 * This class provides methods to create and manage Prometheus metrics.
 */
export class SingletonMetricsService implements IMetricsService {
  private readonly registry: Registry;
  private readonly counters: Map<LineaPostmanMetrics, Counter<string>>;
  private readonly gauges: Map<LineaPostmanMetrics, Gauge<string>>;

  constructor() {
    this.registry = new Registry();
    this.registry.setDefaultLabels({ app: "postman" });

    this.counters = new Map();
    this.gauges = new Map();
  }

  /**
   * Returns the registry
   * @returns {Registry} The registry instance
   */
  public getRegistry(): Registry {
    return this.registry;
  }

  /**
   * Creates counter metric
   */
  public createCounter(name: LineaPostmanMetrics, help: string, labelNames: string[] = []): Counter<string> {
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

  /**
   * Get counter metric value
   * @param name - Name of the metric
   * @param labels - Labels for the metric
   * @returns Value of the counter metric
   */
  public async getCounterValue(name: LineaPostmanMetrics, labels: Record<string, string>): Promise<number | undefined> {
    const counter = this.counters.get(name);
    if (counter === undefined) {
      return undefined;
    }

    const metricData = await counter.get();
    const aggregatedMetricValueWithMatchingLabels = this.aggregateMetricValuesWithExactMatchingLabels(
      metricData,
      labels,
    );
    return aggregatedMetricValueWithMatchingLabels?.value;
  }

  /**
   * Creates gauge metric
   * @param name - Name of the metric
   * @param help - Help text for the metric
   * @param labelNames - Array of label names for the metric
   * @returns Gauge metric
   */
  public createGauge(name: LineaPostmanMetrics, help: string, labelNames: string[] = []): Gauge<string> {
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

  /**
   * Get gauge metric value
   * @param name - Name of the metric
   * @param labels - Labels for the metric
   * @returns Value of the gauge metric
   */
  public async getGaugeValue(name: LineaPostmanMetrics, labels: Record<string, string>): Promise<number | undefined> {
    const gauge = this.gauges.get(name);

    if (gauge === undefined) {
      return undefined;
    }

    const metricData = await gauge.get();
    const aggregatedMetricValueWithMatchingLabels = this.aggregateMetricValuesWithExactMatchingLabels(
      metricData,
      labels,
    );
    return aggregatedMetricValueWithMatchingLabels?.value;
  }

  /**
   * Increments a counter metric
   * @param name - Name of the metric
   * @param labels - Labels for the metric
   * @param value - Value to increment by (default is 1)
   * @returns void
   */
  public incrementCounter(name: LineaPostmanMetrics, labels: Record<string, string> = {}, value?: number): void {
    const counter = this.counters.get(name);
    if (counter !== undefined) {
      counter.inc(labels, value);
    }
  }

  /**
   * Increment a gauge metric value
   * @param name - Name of the metric
   * @param labels - Labels for the metric
   * @param value - Value to increment by (default is 1)
   * @returns void
   */
  public incrementGauge(name: LineaPostmanMetrics, labels: Record<string, string> = {}, value?: number): void {
    const gauge = this.gauges.get(name);
    if (gauge !== undefined) {
      gauge.inc(labels, value);
    }
  }

  /**
   * Decrement a gauge metric value
   * @param name - Name of the metric
   * @param value - Value to decrement by (default is 1)
   * @param labels - Labels for the metric
   * @returns void
   */
  public decrementGauge(name: LineaPostmanMetrics, labels: Record<string, string> = {}, value?: number): void {
    const gauge = this.gauges.get(name);
    if (gauge !== undefined) {
      gauge.dec(labels, value);
    }
  }

  private aggregateMetricValuesWithExactMatchingLabels(
    metricData: MetricObjectWithValues<MetricValue<string>>,
    labels: Record<string, string>,
  ): MetricValue<string> | undefined {
    // It is possible to have multiple metric objects with exact matching labels, e.g. if we query for 2 out of the 3 labels being used.
    // Hence we should merge all metric objects, and remove labels that were not queried from the merged metric object.
    const matchingMetricObjects = metricData.values.filter((value) =>
      Object.entries(labels).every(([key, val]) => value.labels[key] === val),
    );
    if (matchingMetricObjects.length === 0) return undefined;
    const mergedMetricObject: MetricValue<string> = {
      value: 0,
      labels,
    };
    matchingMetricObjects.forEach((m) => (mergedMetricObject.value += m.value));
    return mergedMetricObject;
  }
}
