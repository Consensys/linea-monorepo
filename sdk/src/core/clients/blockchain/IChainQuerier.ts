export interface IChainQuerier<TransactionReceipt> {
  getCurrentNonce(accountAddress?: string): Promise<number>;
  getCurrentBlockNumber(): Promise<number>;
  getTransactionReceipt(txHash: string): Promise<TransactionReceipt | null>;
}
