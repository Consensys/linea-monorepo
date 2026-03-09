import { MessageSent } from "../../../types";

export type MessageSentEventFilters = {
  from?: string;
  to?: string;
  messageHash?: string;
};

export interface IL2MessageServiceLogClient {
  getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;
}
