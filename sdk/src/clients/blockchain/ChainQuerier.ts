import { JsonRpcProvider, Signer, TransactionReceipt } from "ethers";
import { BaseError } from "../../core/errors/Base";
import { IChainQuerier } from "../../core/clients/blockchain/IChainQuerier";

export class ChainQuerier implements IChainQuerier<TransactionReceipt> {
  constructor(
    private readonly provider: JsonRpcProvider,
    private readonly signer?: Signer,
  ) {}

  /**
   * Retrieves the current nonce for a given account address.
   *
   * @param {string} [accountAddress] - The Ethereum account address to fetch the nonce for. Optional if a signer is provided.
   * @returns {Promise<number>} A promise that resolves to the current nonce of the account.
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
}
