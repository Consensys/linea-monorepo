import axios from "axios";
import { Agent } from "https";
import { Hex, serializeTransaction, TransactionSerializable } from "viem";

export async function getWeb3SignerSignature(
  web3SignerUrl: string,
  web3SignerPublicKey: string,
  transaction: TransactionSerializable,
  agent?: Agent,
): Promise<Hex> {
  try {
    const { data } = await axios.post(
      `${web3SignerUrl}/api/v1/eth1/sign/${web3SignerPublicKey}`,
      {
        data: serializeTransaction(transaction),
      },
      { httpsAgent: agent },
    );
    return data;
  } catch (error) {
    if (error instanceof Error) {
      console.log(`Web3SignerError: ${JSON.stringify(error.message)}`);
    }
    throw error;
  }
}
