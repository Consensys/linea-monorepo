import { Address, Hash, MessageSent } from "../../../types";

export type MessageSentEventFilters = {
  from?: Address;
  to?: Address;
  messageHash?: Hash;
};

export interface IL2MessageServiceLogClient {
  getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: bigint;
    toBlock?: string | bigint;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]>;
}
