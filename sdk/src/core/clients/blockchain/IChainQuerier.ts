import { GasFees } from "./IGasProvider";

export interface IChainQuerier<TransactionReceipt, Block, TransactionRequest, TransactionResponse, JsonRpcProvider> {
  getCurrentNonce(accountAddress?: string): Promise<number>;
  getCurrentBlockNumber(): Promise<number>;
  getTransactionReceipt(txHash: string): Promise<TransactionReceipt | null>;
  getBlock(blockNumber: number | bigint | string): Promise<Block | null>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  sendRequest(methodName: string, params: any[]): Promise<any>;
  estimateGas(transactionRequest: TransactionRequest): Promise<bigint>;
  getProvider(): JsonRpcProvider;
  getTransaction(transactionHash: string): Promise<TransactionResponse | null>;
  broadcastTransaction(signedTx: string): Promise<TransactionResponse>;
  ethCall(transactionRequest: TransactionRequest): Promise<string>;
  getFees(): Promise<GasFees>;
}
