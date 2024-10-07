import axios from "axios";
import { ethers, TransactionLike } from "ethers";

export async function getWeb3SignerSignature(
  web3SignerUrl: string,
  web3SignerPublicKey: string,
  transaction: TransactionLike,
): Promise<string> {
  try {
    const { data } = await axios.post(`${web3SignerUrl}/api/v1/eth1/sign/${web3SignerPublicKey}`, {
      data: ethers.Transaction.from(transaction).unsignedSerialized,
    });
    return data;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    throw new Error(`Web3SignerError: ${JSON.stringify(error.message)}`);
  }
}

export async function estimateTransactionGas(
  provider: ethers.JsonRpcProvider,
  transaction: ethers.TransactionRequest,
): Promise<bigint> {
  try {
    return await provider.estimateGas(transaction);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    throw new Error(`GasEstimationError: ${JSON.stringify(error.message)}`);
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
