import { GasFees } from "./IGasProvider";
import { Address, Hash, Hex, Block, TransactionReceipt, TransactionRequest, TransactionSubmission } from "../../types";

export interface IProvider {
  getTransactionCount(address: Address, blockTag: string | number | bigint): Promise<number>;
  getBlockNumber(): Promise<number>;
  getTransactionReceipt(txHash: Hash): Promise<TransactionReceipt | null>;
  getBlock(blockNumber: number | bigint | string): Promise<Block | null>;
  estimateGas(transactionRequest: TransactionRequest): Promise<bigint>;
  getTransaction(transactionHash: Hash): Promise<TransactionSubmission | null>;
  call(transactionRequest: TransactionRequest): Promise<Hex>;
  getFees(): Promise<GasFees>;
}
