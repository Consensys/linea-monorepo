import { type PublicClient, type Hex } from "viem";

import type { IProvider } from "../../../domain/ports/IProvider";
import type { TransactionReceipt, Block } from "../../../domain/types";

export class ViemProvider implements IProvider {
  constructor(protected readonly publicClient: PublicClient) {}

  public async getBlockNumber(): Promise<number> {
    const blockNumber = await this.publicClient.getBlockNumber();
    return Number(blockNumber);
  }

  public async getTransactionReceipt(txHash: Hex): Promise<TransactionReceipt | null> {
    try {
      const receipt = await this.publicClient.getTransactionReceipt({
        hash: txHash,
      });

      return {
        transactionHash: receipt.transactionHash,
        blockNumber: Number(receipt.blockNumber),
        status: receipt.status === "success" ? "success" : "reverted",
        gasPrice: receipt.effectiveGasPrice,
        gasUsed: receipt.gasUsed,
      };
    } catch {
      return null;
    }
  }

  public async getBlock(blockNumber: bigint): Promise<Block | null> {
    try {
      const block = await this.publicClient.getBlock({ blockNumber });

      return {
        number: Number(block.number),
        timestamp: Number(block.timestamp),
      };
    } catch {
      return null;
    }
  }
}
