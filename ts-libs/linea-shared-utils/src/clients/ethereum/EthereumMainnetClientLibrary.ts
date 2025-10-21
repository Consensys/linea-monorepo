import { IContractClientLibrary } from "../../core/client/IContractClientLibrary";
import {
  Hex,
  createPublicClient,
  http,
  PublicClient,
  TransactionReceipt,
  Address,
  TransactionSerializableEIP1559,
  serializeTransaction,
  parseSignature,
} from "viem";
import { mainnet } from "viem/chains";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";
import { IContractSignerService } from "../../core/services/IContractSignerService";

// Re-use via composition in ContractClients
// Hope that using strategy pattern like this makes us more 'viem-agnostic'
export class EthereumMainnetClientLibrary implements IContractClientLibrary<PublicClient, TransactionReceipt> {
  blockchainClient: PublicClient;

  constructor(
    rpcUrl: string,
    private readonly contractSignerService: IContractSignerService,
  ) {
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

  async estimateGasFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }> {
    const { maxFeePerGas, maxPriorityFeePerGas } = await this.blockchainClient.estimateFeesPerGas();
    return { maxFeePerGas, maxPriorityFeePerGas };
  }

  async sendSignedTransaction(contractAddress: Address, calldata: Hex): Promise<TransactionReceipt> {
    const [fees, gasLimit, chainId] = await Promise.all([
      this.estimateGasFees(),
      this.blockchainClient.estimateGas({ to: contractAddress, data: calldata }),
      this.getChainId(),
    ]);
    const { maxFeePerGas, maxPriorityFeePerGas } = fees;
    const tx: TransactionSerializableEIP1559 = {
      to: contractAddress,
      type: "eip1559",
      data: calldata,
      chainId: chainId,
      gas: gasLimit,
      maxFeePerGas,
      maxPriorityFeePerGas,
    };
    const signature = await this.contractSignerService.sign(tx);
    const serializedTransaction = serializeTransaction(tx, parseSignature(signature));
    const txHash = await sendRawTransaction(this.blockchainClient, { serializedTransaction });
    const receipt = await waitForTransactionReceipt(this.blockchainClient, { hash: txHash });
    return receipt;
  }
}
