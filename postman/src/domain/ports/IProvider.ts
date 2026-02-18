import { Block, TransactionReceipt } from "../types";

export interface IProvider {
  getBlockNumber(): Promise<number>;
  getTransactionReceipt(txHash: string): Promise<TransactionReceipt | null>;
  getBlock(blockNumber: number | bigint | string): Promise<Block | null>;
}

export interface ILineaProvider extends IProvider {
  getBlockExtraData(blockNumber: number | bigint | string): Promise<{
    version: number;
    fixedCost: number;
    variableCost: number;
    ethGasPrice: number;
  } | null>;
}
