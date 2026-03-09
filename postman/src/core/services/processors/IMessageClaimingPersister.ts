import { Direction } from "../../enums";

export interface IMessageClaimingPersister {
  process(): Promise<void>;
}

export type MessageClaimingPersisterConfig = {
  direction: Direction;
  messageSubmissionTimeout: number;
  maxTxRetries: number;
  receiptPollingTimeout: number;
  receiptPollingInterval: number;
};
