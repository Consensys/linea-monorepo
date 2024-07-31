import { Direction } from "../../enums/MessageEnums";

export interface IMessageClaimingPersister {
  process(): Promise<void>;
}

export type MessageClaimingPersisterConfig = {
  direction: Direction;
  messageSubmissionTimeout: number;
  maxTxRetries: number;
};
