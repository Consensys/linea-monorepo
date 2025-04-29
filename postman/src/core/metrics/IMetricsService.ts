import { Counter, Gauge, Registry } from "prom-client";
import { MessageStatus } from "../enums";
import { Direction } from "@consensys/linea-sdk";

export enum MetricsOperation {
  INCREMENT = "INCREMENT",
  DECREMENT = "DECREMENT",
}

export enum LineaPostmanMetrics {
  Messages = "linea_postman_messages",
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
   */
  SponsorshipFeesWei = "linea_postman_sponsorship_fees_wei", // Represent up to ~9_007_199 GWEI
  SponsorshipFeesGwei = "linea_postman_sponsorship_fees_gwei", // Represent up to ~9_007_199 ETH
}

export type MessagesMetricsAttributes = {
  status: MessageStatus;
  direction: Direction;
  isForSponsorship: boolean;
};

export type SponsorshipFeesMetricsAttributes = {
  // Only a message with MessageStatus.CLAIMED_SUCCESS will have claimTxGasUsed and claimTxGasPrice properties
  direction: Direction;
};

export interface IMetricsService extends IMetricsServiceUtils {
  getRegistry(): Registry;
  createCounter(name: LineaPostmanMetrics, help: string, labelNames?: string[]): Counter<string>;
  createGauge(name: LineaPostmanMetrics, help: string, labelNames?: string[]): Gauge<string>;
  incrementCounter(name: LineaPostmanMetrics, labels?: Record<string, string>, value?: number): void;
  incrementGauge(name: LineaPostmanMetrics, labels?: Record<string, string>, value?: number): void;
  decrementGauge(name: LineaPostmanMetrics, labels?: Record<string, string>, value?: number): void;
  getGaugeValue(name: LineaPostmanMetrics, labels: Record<string, string>): Promise<number | undefined>;
  getCounterValue(name: LineaPostmanMetrics, labels: Record<string, string>): Promise<number | undefined>;
}

interface IMetricsServiceUtils {
  convertTxFeesToWeiAndGwei(txFees: bigint): { gwei: number; wei: number };
}
