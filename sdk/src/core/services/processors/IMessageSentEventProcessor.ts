export interface IMessageSentEventProcessor {
  getAndStoreMessageSentEvents(
    fromBlock: number,
    fromBlockLogIndex: number,
  ): Promise<{ nextFromBlock: number; nextFromBlockLogIndex: number }>;
}
