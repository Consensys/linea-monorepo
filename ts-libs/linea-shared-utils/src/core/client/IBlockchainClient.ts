import { Address, Hex } from "viem";

// TODO - Make generic and uncoupled from Viem
export interface IBlockchainClient<TClient, TTransactionReceipt> {
  getBlockchainClient(): TClient;
  estimateGasFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }>;
  getChainId(): Promise<number>;
  getBalance(address: Address): Promise<bigint>;
  sendSignedTransaction(contractAddress: Address, calldata: Hex, value?: bigint): Promise<TTransactionReceipt>;
}
