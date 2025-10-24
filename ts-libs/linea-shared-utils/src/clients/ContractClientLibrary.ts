import { IContractClientLibrary } from "../core/client/IContractClientLibrary";
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
  Chain,
} from "viem";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";
import { IContractSignerClient } from "../core/client/IContractSignerClient";
import { ILogger } from "../logging/ILogger";

// Re-use via composition in ContractClients
// Hope that using strategy pattern like this makes us more 'viem-agnostic'
export class ContractClientLibrary implements IContractClientLibrary<PublicClient, TransactionReceipt> {
  blockchainClient: PublicClient;

  constructor(
    private readonly logger: ILogger,
    rpcUrl: string,
    chain: Chain,
    private readonly contractSignerClient: IContractSignerClient,
  ) {
    // Aim re-use single blockchain client for
    // i.) Better connection pooling
    // ii.) Memory efficient
    // iii.) Single point of configuration
    this.blockchainClient = createPublicClient({
      chain,
      transport: http(rpcUrl, { batch: true, retryCount: 3 }),
    });
  }

  getBlockchainClient(): PublicClient {
    return this.blockchainClient;
  }

  async getChainId(): Promise<number> {
    return await this.blockchainClient.getChainId();
  }

  async getBalance(address: Address): Promise<bigint> {
    return await this.blockchainClient.getBalance({
      address,
    });
  }

  async estimateGasFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }> {
    const { maxFeePerGas, maxPriorityFeePerGas } = await this.blockchainClient.estimateFeesPerGas();
    return { maxFeePerGas, maxPriorityFeePerGas };
  }

  async sendSignedTransaction(contractAddress: Address, calldata: Hex): Promise<TransactionReceipt> {
    const [fees, gasLimit, chainId, nonce] = await Promise.all([
      this.estimateGasFees(),
      this.blockchainClient.estimateGas({ to: contractAddress, data: calldata }),
      this.getChainId(),
      this.blockchainClient.getTransactionCount({ address: this.contractSignerClient.getAddress() }),
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
      nonce,
    };
    const signature = await this.contractSignerClient.sign(tx);
    const serializedTransaction = serializeTransaction(tx, parseSignature(signature));
    this.logger.debug(
      `ContractClientLibrary: sending transaction to ${contractAddress} nonce=${nonce} gas=${gasLimit}`,
    );
    const txHash = await sendRawTransaction(this.blockchainClient, { serializedTransaction });
    this.logger.debug(`ContractClientLibrary: txHash=${txHash}`);
    const receipt = await waitForTransactionReceipt(this.blockchainClient, { hash: txHash });
    return receipt;
  }
}
