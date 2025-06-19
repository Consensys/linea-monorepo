import { MessageSent, ServiceVersionMigrated } from "../../types";

export type MessageSentEventFilters = {
  from?: string;
  to?: string;
  messageHash?: string;
};

export interface IL2MessageServiceLogClient {
  getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: string | number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;

  getMessageSentEventsByMessageHash(params: {
    messageHash: string;
    fromBlock?: string | number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;

  getMessageSentEventsByBlockRange(fromBlock: string | number, toBlock: string | number): Promise<MessageSent[]>;

  getServiceVersionMigratedEvents(param?: {
    fromBlock?: string | number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<ServiceVersionMigrated[]>;
}
