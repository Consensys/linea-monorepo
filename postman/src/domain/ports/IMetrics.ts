import { Direction, MessageStatus } from "../types/enums";

export type MessagesMetricsAttributes = {
  status: MessageStatus;
  direction: Direction;
};

export enum LineaPostmanMetrics {
  Messages = "linea_postman_messages",
  SponsoredMessagesTotal = "linea_postman_sponsored_messages_total",
  SponsorshipFeesWei = "linea_postman_sponsorship_fees_wei_total",
  SponsorshipFeesGwei = "linea_postman_sponsorship_fees_gwei_total",
  TransactionProcessingTime = "linea_postman_l2_transaction_tx_processing_time",
  TransactionInfuraConfirmationTime = "linea_postman_l2_transaction_tx_infura_confirmation_time",
}

export interface IMessageMetricsUpdater {
  initialize(): Promise<void>;
  getMessageCount(messageAttributes: MessagesMetricsAttributes): Promise<number | undefined>;
  incrementMessageCount(messageAttributes: MessagesMetricsAttributes, value?: number): Promise<void>;
  decrementMessageCount(messageAttributes: MessagesMetricsAttributes, value?: number): Promise<void>;
}

export interface ISponsorshipMetricsUpdater {
  getSponsoredMessagesTotal(direction: Direction): Promise<number>;
  getSponsorshipFeePaid(direction: Direction): Promise<bigint>;
  incrementSponsorshipFeePaid(txFee: bigint, direction: Direction): Promise<void>;
}

export interface ITransactionMetricsUpdater {
  addTransactionProcessingTime(direction: string, transactionProcessingTimeInSeconds: number): void;
  addTransactionInfuraConfirmationTime(direction: string, transactionBroadcastTimeInSeconds: number): void;
}
