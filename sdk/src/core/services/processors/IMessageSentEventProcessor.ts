import { Direction } from "../../enums/MessageEnums";

export interface IMessageSentEventProcessor {
  process(
    fromBlock: number,
    fromBlockLogIndex: number,
  ): Promise<{ nextFromBlock: number; nextFromBlockLogIndex: number }>;
}
export type MessageSentEventProcessorConfig = {
  direction: Direction;
  maxBlocksToFetchLogs: number;
  blockConfirmation: number;
  isEOAEnabled: boolean;
  isCalldataEnabled: boolean;
};
