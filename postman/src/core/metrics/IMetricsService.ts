import { Counter, Gauge, Histogram, Registry } from "prom-client";

export interface IMetricsService {
  getRegistry(): Registry;
  createCounter(name: string, help: string, labelNames?: string[]): Counter<string>;
  createGauge(name: string, help: string, labelNames?: string[]): Gauge<string>;
  createHistogram(name: string, help: string, labelNames?: string[], buckets?: number[]): Histogram<string>;
  incrementCounter(name: string, labels?: Record<string, string>, value?: number): void;
  incrementGauge(name: string, value: number, labels?: Record<string, string>): void;
  decrementGauge(name: string, value: number, labels?: Record<string, string>): void;
  observeHistogram(name: string, value: number, labels?: Record<string, string>): void;
  getGaugeValue(name: string, labels: Record<string, string>): Promise<number | undefined>;
  getCounterValue(name: string, labels: Record<string, string>): Promise<number | undefined>;
}
