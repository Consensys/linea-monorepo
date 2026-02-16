import { type PublicClient, type Hex, numberToHex } from "viem";

import type { IProvider } from "../../../domain/ports/IProvider";
import type { TransactionReceipt, TransactionResponse, Block, GasFees } from "../../../domain/types";

export class ViemProvider implements IProvider {
  constructor(protected readonly publicClient: PublicClient) {}

  public async getTransactionCount(address: string, blockTag: string | number | bigint): Promise<number> {
    let blockTagParam: "latest" | "pending" | "earliest" | undefined;
    if (typeof blockTag === "string" && ["latest", "pending", "earliest"].includes(blockTag)) {
      blockTagParam = blockTag as "latest" | "pending" | "earliest";
    }

    return this.publicClient.getTransactionCount({
      address: address as `0x${string}`,
      ...(blockTagParam ? { blockTag: blockTagParam } : {}),
    });
  }

  public async getBlockNumber(): Promise<number> {
    const blockNumber = await this.publicClient.getBlockNumber();
    return Number(blockNumber);
  }

  public async getTransactionReceipt(txHash: string): Promise<TransactionReceipt | null> {
    try {
      const receipt = await this.publicClient.getTransactionReceipt({
        hash: txHash as Hex,
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

  public async getBlock(blockNumber: number | bigint | string): Promise<Block | null> {
    try {
      let blockTag: "latest" | "pending" | "earliest" | undefined;
      let blockNum: bigint | undefined;

      if (typeof blockNumber === "string") {
        if (["latest", "pending", "earliest"].includes(blockNumber)) {
          blockTag = blockNumber as "latest" | "pending" | "earliest";
        } else {
          blockNum = BigInt(blockNumber);
        }
      } else {
        blockNum = BigInt(blockNumber);
      }

      const block = await this.publicClient.getBlock(blockTag ? { blockTag } : { blockNumber: blockNum });

      return {
        number: Number(block.number),
        timestamp: Number(block.timestamp),
      };
    } catch {
      return null;
    }
  }

  public async estimateGas(transactionRequest: unknown): Promise<bigint> {
    return this.publicClient.estimateGas(transactionRequest as Parameters<PublicClient["estimateGas"]>[0]);
  }

  public async getTransaction(transactionHash: string): Promise<TransactionResponse | null> {
    try {
      const tx = await this.publicClient.getTransaction({
        hash: transactionHash as Hex,
      });

      return {
        hash: tx.hash,
        gasLimit: tx.gas,
        maxFeePerGas: tx.maxFeePerGas ?? undefined,
        maxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
        nonce: tx.nonce,
      };
    } catch {
      return null;
    }
  }

  public async broadcastTransaction(signedTx: string): Promise<TransactionResponse> {
    const hash = await this.publicClient.sendRawTransaction({
      serializedTransaction: signedTx as Hex,
    });

    const tx = await this.publicClient.getTransaction({ hash });
    return {
      hash: tx.hash,
      gasLimit: tx.gas,
      maxFeePerGas: tx.maxFeePerGas ?? undefined,
      maxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
      nonce: tx.nonce,
    };
  }

  public async call(transactionRequest: unknown): Promise<string> {
    const result = await this.publicClient.call(transactionRequest as Parameters<PublicClient["call"]>[0]);
    return result.data ?? "0x";
  }

  public async getFees(): Promise<GasFees> {
    const feeHistory = await this.publicClient.request({
      method: "eth_feeHistory",
      params: [numberToHex(4), "latest", [20]],
    });

    const rewards = (feeHistory as { reward: Hex[][] }).reward;
    const baseFeePerGas = (feeHistory as { baseFeePerGas: Hex[] }).baseFeePerGas;

    const avgPriorityFee = rewards.reduce((acc: bigint, r: Hex[]) => acc + BigInt(r[0]), 0n) / BigInt(rewards.length);

    const latestBaseFee = BigInt(baseFeePerGas[baseFeePerGas.length - 1]);
    const maxFeePerGas = latestBaseFee * 2n + avgPriorityFee;

    return {
      maxPriorityFeePerGas: avgPriorityFee,
      maxFeePerGas,
    };
  }
}
