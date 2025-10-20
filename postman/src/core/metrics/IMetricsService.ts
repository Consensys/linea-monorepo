import { Counter, Gauge, Histogram, MetricObjectWithValues, MetricValueWithName, Registry } from "prom-client";

export enum LineaPostmanMetrics {
  Messages = "linea_postman_messages",
  // Example PromQL query for hourly rate of sponsored messages 'rate(linea_postman_sponsored_messages_total{direction="L1_TO_L2",app="postman"}[60m]) * 3600'
  SponsoredMessagesTotal = "linea_postman_sponsored_messages_total",
  /**
   * Tx fees in wei paid by Postman for sponsored message claims
   *
   * Workaround for prom-client metrics not supporting bigint type
   * - We split txFee into GWEI and WEI components, and accumulate in two separate metrics
   * - Note JS limitation of Number.MAX_SAFE_INTEGER = 9007199254740991
   * - Given 150,000 sponsored messages a year, we should not reach overflow in <60 years
   *
   * We do not use separate labels for 'wei' and 'gwei' denominations, because metrics sharing the label should be aggregatable
   * - I.e. metric (direction: A, denomination: wei) cannot be aggregated with metric (direction: A, denomination: gwei) because they represent different units
   *
   * Example PromQL query to get hourly rate of ETH consumed for sponsoring messages - 'rate(linea_postman_sponsorship_fees_gwei_total{direction="L1_TO_L2", app="postman"}[60m]) * 3600 / 1e9 + rate(linea_postman_sponsorship_fees_wei_total{direction="L1_TO_L2", app="postman"}[60m]) * 3600 / 1e18
   */
  SponsorshipFeesWei = "linea_postman_sponsorship_fees_wei_total", // Represent up to ~9_007_199 GWEI
  SponsorshipFeesGwei = "linea_postman_sponsorship_fees_gwei_total", // Represent up to ~9_007_199 ETH

  TransactionProcessingTime = "linea_postman_l2_transaction_tx_processing_time",
}

export interface IMetricsService<TMetricName extends string = string> {
  getRegistry(): Registry;
  createCounter(name: TMetricName, help: string, labelNames?: string[]): Counter<string>;
  createGauge(name: TMetricName, help: string, labelNames?: string[]): Gauge<string>;
  incrementCounter(name: TMetricName, labels?: Record<string, string>, value?: number): void;
  incrementGauge(name: TMetricName, labels?: Record<string, string>, value?: number): void;
  decrementGauge(name: TMetricName, labels?: Record<string, string>, value?: number): void;
  getGaugeValue(name: TMetricName, labels: Record<string, string>): Promise<number | undefined>;
  getCounterValue(name: TMetricName, labels: Record<string, string>): Promise<number | undefined>;
  createHistogram(name: TMetricName, buckets: number[], help: string, labelNames?: string[]): Histogram<string>;
  addValueToHistogram(name: TMetricName, value: number, labels?: Record<string, string>): void;
  getHistogramMetricsValues(
    name: TMetricName,
  ): Promise<MetricObjectWithValues<MetricValueWithName<string>> | undefined>;
}

export type LineaPostmanMetricsService = IMetricsService<LineaPostmanMetrics>;
