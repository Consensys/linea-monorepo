import {
  Block,
  BlockTag,
  JsonRpcProvider,
  Signer,
  TransactionReceipt,
  TransactionRequest,
  TransactionResponse,
} from "ethers";
import { BaseError } from "../../core/errors/Base";
import { IChainQuerier } from "../../core/clients/blockchain/IChainQuerier";
import { GasFees } from "../../core/clients/blockchain/IGasProvider";

export class ChainQuerier
  implements IChainQuerier<TransactionReceipt, Block, TransactionRequest, TransactionResponse, JsonRpcProvider>
{
  /**
   * Creates an instance of ChainQuerier.
   *
   * @param {JsonRpcProvider} provider - The JSON RPC provider for interacting with the Ethereum network.
   * @param {Signer} [signer] - An optional Ethers.js signer object for signing transactions.
   */
  constructor(
    protected readonly provider: JsonRpcProvider,
    protected readonly signer?: Signer,
  ) {}

  /**
   * Retrieves the current nonce for a given account address.
   *
   * @param {string} [accountAddress] - The Ethereum account address to fetch the nonce for. Optional if a signer is provided.
   * @returns {Promise<number>} A promise that resolves to the current nonce of the account.
   * @throws {BaseError} If no account address is provided and no signer is available.
   */
  public async getCurrentNonce(accountAddress?: string): Promise<number> {
    if (accountAddress) {
      return this.provider.getTransactionCount(accountAddress);
    }

    if (!this.signer) {
      throw new BaseError("Please provide a signer.");
    }

    return this.provider.getTransactionCount(await this.signer.getAddress());
  }

  /**
   * Retrieves the current block number of the blockchain.
   *
   * @returns {Promise<number>} A promise that resolves to the current block number.
   */
  public async getCurrentBlockNumber(): Promise<number> {
    return this.provider.getBlockNumber();
  }

  /**
   * Retrieves the transaction receipt for a given transaction hash.
   *
   * @param {string} transactionHash - The hash of the transaction to fetch the receipt for.
   * @returns {Promise<TransactionReceipt | null>} A promise that resolves to the transaction receipt, or null if the transaction is not found.
   */
  public async getTransactionReceipt(transactionHash: string): Promise<TransactionReceipt | null> {
    return this.provider.getTransactionReceipt(transactionHash);
  }

  /**
   * Retrieves a block by its number or tag.
   *
   * @param {BlockTag} blockNumber - The block number or tag to fetch.
   * @returns {Promise<Block | null>} A promise that resolves to the block, or null if the block is not found.
   */
  public async getBlock(blockNumber: BlockTag): Promise<Block | null> {
    return this.provider.getBlock(blockNumber);
  }

  /**
   * Sends a custom JSON-RPC request.
   *
   * @param {string} methodName - The name of the JSON-RPC method to call.
   * @param {any[]} params - The parameters to pass to the JSON-RPC method.
   * @returns {Promise<any>} A promise that resolves to the result of the JSON-RPC call.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  public async sendRequest(methodName: string, params: any[]): Promise<any> {
    return this.provider.send(methodName, params);
  }

  /**
   * Estimates the gas required for a transaction.
   *
   * @param {TransactionRequest} transactionRequest - The transaction request to estimate gas for.
   * @returns {Promise<bigint>} A promise that resolves to the estimated gas.
   */
  public async estimateGas(transactionRequest: TransactionRequest): Promise<bigint> {
    return this.provider.estimateGas(transactionRequest);
  }

  /**
   * Gets the JSON RPC provider.
   *
   * @returns {JsonRpcProvider} The JSON RPC provider.
   */
  public getProvider(): JsonRpcProvider {
    return this.provider;
  }

  /**
   * Retrieves a transaction by its hash.
   *
   * @param {string} transactionHash - The hash of the transaction to fetch.
   * @returns {Promise<TransactionResponse | null>} A promise that resolves to the transaction, or null if the transaction is not found.
   */
  public async getTransaction(transactionHash: string): Promise<TransactionResponse | null> {
    return this.provider.getTransaction(transactionHash);
  }

  /**
   * Sends a signed transaction.
   *
   * @param {string} signedTx - The signed transaction to broadcast.
   * @returns {Promise<TransactionResponse>} A promise that resolves to the transaction response.
   */
  public async broadcastTransaction(signedTx: string): Promise<TransactionResponse> {
    return this.provider.broadcastTransaction(signedTx);
  }

  /**
   * Executes a call on the Ethereum network.
   *
   * @param {TransactionRequest} transactionRequest - The transaction request to execute.
   * @returns {Promise<string>} A promise that resolves to the result of the call.
   */
  public ethCall(transactionRequest: TransactionRequest): Promise<string> {
    return this.provider.call(transactionRequest);
  }

  /**
   * Retrieves the current gas fees.
   *
   * @returns {Promise<GasFees>} A promise that resolves to an object containing the current gas fees.
   * @throws {BaseError} If there is an error getting the fee data.
   */
  public async getFees(): Promise<GasFees> {
    const { maxPriorityFeePerGas, maxFeePerGas } = await this.provider.getFeeData();

    if (!maxPriorityFeePerGas || !maxFeePerGas) {
      throw new BaseError("Error getting fee data");
    }

    return { maxPriorityFeePerGas, maxFeePerGas };
  }
}
