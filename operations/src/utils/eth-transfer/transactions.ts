import { BaseError, Client } from "viem";
import { estimateGas, EstimateGasParameters, EstimateGasReturnType } from "viem/linea";
import { ethers, TransactionLike } from "ethers";

export async function estimateTransactionGas(
  client: Client,
  params: EstimateGasParameters,
): Promise<EstimateGasReturnType> {
  try {
    return await estimateGas(client, params);
  } catch (error) {
    if (error instanceof BaseError) {
      const err = (error.walk((err) => "data" in (err as BaseError)) || error.walk()) as BaseError;
      console.log("Gas estimation failed with the following error:", err.message);
    }
    throw error;
  }
}

export async function executeTransaction(
  provider: ethers.JsonRpcProvider,
  transaction: TransactionLike,
): Promise<ethers.TransactionReceipt | null> {
  try {
    const tx = await provider.broadcastTransaction(ethers.Transaction.from(transaction).serialized);
    return await tx.wait();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    throw new Error(`TransactionError: ${JSON.stringify(error.message)}`);
  }
}
