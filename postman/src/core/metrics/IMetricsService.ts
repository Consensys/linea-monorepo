import { Counter, Gauge, Registry } from "prom-client";
import { MessageStatus } from "../enums";
import { Direction } from "@consensys/linea-sdk";

export enum LineaPostmanMetrics {
  Messages = "linea_postman_messages",
  // Tx fees paid by Postman for sponsored message claims
  SponsorshipFees = "linea_postman_sponsorship_fees",
}

export type MessagesMetricsAttributes = {
  status: MessageStatus;
  direction: Direction;
  isForSponsorship: boolean;
};

export type MessagesMetricsAttributesWithCount = {
  attributes: MessagesMetricsAttributes;
  count: number;
};

export interface IMetricsService {
  getRegistry(): Registry;
  createCounter(name: LineaPostmanMetrics, help: string, labelNames?: string[]): Counter<string>;
  createGauge(name: LineaPostmanMetrics, help: string, labelNames?: string[]): Gauge<string>;
  incrementCounter(name: LineaPostmanMetrics, labels?: Record<string, string>, value?: number): void;
  incrementGauge(name: LineaPostmanMetrics, labels?: Record<string, string>, value?: number): void;
  decrementGauge(name: LineaPostmanMetrics, labels?: Record<string, string>, value?: number): void;
  getGaugeValue(name: LineaPostmanMetrics, labels: Record<string, string>): Promise<number | undefined>;
  getCounterValue(name: LineaPostmanMetrics, labels: Record<string, string>): Promise<number | undefined>;
}
