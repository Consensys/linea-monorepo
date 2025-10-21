import { Hex } from "viem";
import { Result } from "neverthrow";

// TODO - Make generic and uncoupled from Viem
export interface IContractClientLibrary<TClient, TTransactionReceipt> {
  getBlockchainClient(): TClient;
  sendSerializedTransaction(serializedTransaction: Hex): Promise<TTransactionReceipt>;
  estimateGasFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }>;
  getChainId(): Promise<number>;
}
