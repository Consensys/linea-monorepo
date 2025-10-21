import { IContractClientLibrary } from "../../core/client/IContractClientLibrary";
import { BaseError, Hex, createPublicClient, http, PublicClient, TransactionReceipt } from "viem";
import { mainnet } from "viem/chains";
import { err, ok, Result } from "neverthrow";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";

// Re-use via composition in ContractClients
// Hope that using strategy pattern like this makes us more 'viem-agnostic'
export class EthereumMainnetClientLibrary
  implements IContractClientLibrary<PublicClient, TransactionReceipt, BaseError>
{
  blockchainClient: PublicClient;

  constructor(rpcUrl: string) {
    // Aim re-use single blockchain client for
    // i.) Better connection pooling
    // ii.) Memory efficient
    // iii.) Single point of configuration
    this.blockchainClient = createPublicClient({
      chain: mainnet,
      transport: http(rpcUrl, { batch: true, retryCount: 3 }),
    });
  }

  getBlockchainClient(): PublicClient {
    return this.blockchainClient;
  }

  // estimateTransactionGas;

  async sendSerializedTransaction(serializedTransaction: Hex): Promise<Result<TransactionReceipt, BaseError>> {
    try {
      const txHash = await sendRawTransaction(this.blockchainClient, { serializedTransaction });
      const receipt = await waitForTransactionReceipt(this.blockchainClient, { hash: txHash });
      return ok(receipt);
    } catch (error) {
      if (error instanceof BaseError) {
        const decodedError = error.walk();
        return err(decodedError as BaseError);
      }
      return err(error as BaseError);
    }
  }

  // const gasEstimationResult = await estimateTransactionGas(client, {
  //   to: contractAddress,
  //   account: senderAddress,
  //   value: 0n,
  //   data: encodeFunctionData({
  //     abi: SUBMIT_INVOICE_ABI,
  //     functionName: "submitInvoice",
  //     args: [
  //       BigInt(Math.floor(invoicePeriod.startDate.getTime() / 1000)),
  //       BigInt(Math.floor(invoicePeriod.endDate.getTime() / 1000)),
  //       BigInt(totalCostsInEth),
  //     ],
  //   }),
  // });
}
