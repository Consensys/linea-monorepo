import { MessageSent } from "../types";

export type MessageSentEventFilters = {
  from?: string;
  to?: string;
  messageHash?: string;
};

export type L2MessagingBlockAnchoredFilters = {
  l2Block: bigint;
};

export type MessageClaimedFilters = {
  messageHash: string;
};

export interface ILogClient {
  getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: bigint;
    toBlock?: string | bigint;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;
}
