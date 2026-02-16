import { Block, GasFees, TransactionReceipt, TransactionResponse } from "../types";

export interface IProvider {
  getTransactionCount(address: string, blockTag: string | number | bigint): Promise<number>;
  getBlockNumber(): Promise<number>;
  getTransactionReceipt(txHash: string): Promise<TransactionReceipt | null>;
  getBlock(blockNumber: number | bigint | string): Promise<Block | null>;
  estimateGas(transactionRequest: unknown): Promise<bigint>;
  getTransaction(transactionHash: string): Promise<TransactionResponse | null>;
  broadcastTransaction(signedTx: string): Promise<TransactionResponse>;
  call(transactionRequest: unknown): Promise<string>;
  getFees(): Promise<GasFees>;
}

export interface ILineaProvider extends IProvider {
  getBlockExtraData(blockNumber: number | bigint | string): Promise<{
    version: number;
    fixedCost: number;
    variableCost: number;
    ethGasPrice: number;
  } | null>;
}
