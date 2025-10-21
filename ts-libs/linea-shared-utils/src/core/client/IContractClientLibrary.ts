import { Hex } from "viem";
import { Result } from "neverthrow";

// TODO - Make generic and uncoupled from Viem
export interface IContractClientLibrary<TClient, TTransactionReceipt, TError> {
  getBlockchainClient(): TClient;
  sendSerializedTransaction(serializedTransaction: Hex): Promise<Result<TTransactionReceipt, TError>>;

  // retryTransactionWithHigherFee(
  //   transactionHash: string,
  //   priceBumpPercent: number,
  // ): Promise<void>
}
