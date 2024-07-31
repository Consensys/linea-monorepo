import { BlockTag, JsonRpcProvider, Signer, dataSlice, toNumber } from "ethers";
import { ChainQuerier } from "../ChainQuerier";
import { BlockExtraData } from "../../../core/clients/blockchain/linea/IL2ChainQuerier";

export class L2ChainQuerier extends ChainQuerier {
  private blockExtraDataCache: BlockExtraData;
  private cacheIsValidForBlockNumber: bigint;

  /**
   * Creates an instance of `L2ChainQuerier`.
   *
   * @param {JsonRpcProvider} provider - The JSON RPC provider for interacting with the Ethereum network.
   * @param {Signer} [signer] - An optional Ethers.js signer object for signing transactions.
   */
  constructor(provider: JsonRpcProvider, signer?: Signer) {
    super(provider, signer);
  }

  /**
   * Fetches and format extra data from a block.
   *
   * @param {BlockTag} blockNumber - The block number or tag to fetch extra data for.
   * @returns {Promise<BlockExtraData | null>} A promise that resolves to an object containing the formatted block's extra data, or null if the block is not found.
   */
  public async getBlockExtraData(blockNumber: BlockTag): Promise<BlockExtraData | null> {
    if (typeof blockNumber === "number" && this.isCacheValid(blockNumber)) {
      return this.blockExtraDataCache;
    }

    const block = await this.getBlock(blockNumber);

    if (!block) {
      return null;
    }

    const version = dataSlice(block.extraData, 0, 1);
    const fixedCost = dataSlice(block.extraData, 1, 5);
    const variableCost = dataSlice(block.extraData, 5, 9);
    const ethGasPrice = dataSlice(block.extraData, 9, 13);

    // original values are in Kwei and here we convert them back to wei
    const extraData = {
      version: toNumber(version),
      fixedCost: toNumber(fixedCost) * 1000,
      variableCost: toNumber(variableCost) * 1000,
      ethGasPrice: toNumber(ethGasPrice) * 1000,
    };

    if (typeof blockNumber === "number") {
      this.cacheIsValidForBlockNumber = BigInt(blockNumber);
      this.blockExtraDataCache = extraData;
    }
    return extraData;
  }

  /**
   * Checks if the cached block extra data is still valid based on the current block number.
   *
   * @private
   * @param {number} currentBlockNumber - The current block number.
   * @returns {boolean} True if the cache is valid, false otherwise.
   */
  private isCacheValid(currentBlockNumber: number): boolean {
    return this.cacheIsValidForBlockNumber >= BigInt(currentBlockNumber);
  }
}
