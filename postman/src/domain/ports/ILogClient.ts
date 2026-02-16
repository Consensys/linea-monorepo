import { MessageSent, L2MessagingBlockAnchored, MessageClaimed, ServiceVersionMigrated } from "../types";

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

export interface IL1LogClient {
  getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;

  getL2MessagingBlockAnchoredEvents(params: {
    filters?: L2MessagingBlockAnchoredFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<L2MessagingBlockAnchored[]>;

  getMessageClaimedEvents(params: {
    filters?: MessageClaimedFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageClaimed[]>;
}

export interface IL2LogClient {
  getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;

  getMessageSentEventsByMessageHash(params: {
    messageHash: string;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;

  getMessageSentEventsByBlockRange(fromBlock: number, toBlock: number): Promise<MessageSent[]>;

  getServiceVersionMigratedEvents(param?: {
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<ServiceVersionMigrated[]>;
}
