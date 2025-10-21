import { IContractClientLibrary } from "../../core/client/IContractClientLibrary";
import { Hex, createPublicClient, http, PublicClient, TransactionReceipt } from "viem";
import { mainnet } from "viem/chains";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";

// Re-use via composition in ContractClients
// Hope that using strategy pattern like this makes us more 'viem-agnostic'
export class EthereumMainnetClientLibrary implements IContractClientLibrary<PublicClient, TransactionReceipt> {
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

  getChainId(): Promise<number> {
    return this.blockchainClient.getChainId();
  }

  async sendSerializedTransaction(serializedTransaction: Hex): Promise<TransactionReceipt> {
    const txHash = await sendRawTransaction(this.blockchainClient, { serializedTransaction });
    const receipt = await waitForTransactionReceipt(this.blockchainClient, { hash: txHash });
    return receipt;
  }

  async estimateGasFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }> {
    const { maxFeePerGas, maxPriorityFeePerGas } = await this.blockchainClient.estimateFeesPerGas();
    return { maxFeePerGas, maxPriorityFeePerGas };
  }
}
