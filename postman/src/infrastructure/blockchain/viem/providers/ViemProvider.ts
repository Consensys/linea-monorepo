import { BlockNumber, type BlockTag, type PublicClient } from "viem";

import { GasFees } from "../../../../core/clients/blockchain/IGasProvider";
import { IProvider } from "../../../../core/clients/blockchain/IProvider";
import {
  Address,
  Hash,
  Hex,
  Block,
  TransactionReceipt,
  TransactionRequest,
  TransactionSubmission,
} from "../../../../core/types";
import { mapViemBlockToCoreBlock, mapViemReceiptToCoreReceipt, mapViemTransactionToCoreSubmission } from "../mappers";

export class ViemProvider implements IProvider {
  constructor(protected readonly client: PublicClient) {}

  public async getTransactionCount(address: Address, blockTag: BlockTag): Promise<number> {
    return this.client.getTransactionCount({
      address,
      blockTag,
    });
  }

  public async getBlockNumber(): Promise<number> {
    const blockNumber = await this.client.getBlockNumber();
    return Number(blockNumber);
  }

  public async getTransactionReceipt(txHash: Hash): Promise<TransactionReceipt | null> {
    try {
      const receipt = await this.client.getTransactionReceipt({ hash: txHash });
      return mapViemReceiptToCoreReceipt(receipt);
    } catch {
      return null;
    }
  }

  public async getBlock(blockNumber: BlockNumber | BlockTag): Promise<Block | null> {
    try {
      let block;
      if (typeof blockNumber === "string") {
        block = await this.client.getBlock({ blockTag: blockNumber });
      } else {
        block = await this.client.getBlock({ blockNumber: blockNumber });
      }
      return mapViemBlockToCoreBlock(block);
    } catch {
      return null;
    }
  }

  public async estimateGas(transactionRequest: TransactionRequest): Promise<bigint> {
    return this.client.estimateGas({
      account: transactionRequest.from,
      to: transactionRequest.to,
      data: transactionRequest.data,
      value: transactionRequest.value,
      gas: transactionRequest.gasLimit,
      maxFeePerGas: transactionRequest.maxFeePerGas,
      maxPriorityFeePerGas: transactionRequest.maxPriorityFeePerGas,
    });
  }

  public async getTransaction(transactionHash: Hash): Promise<TransactionSubmission | null> {
    try {
      const tx = await this.client.getTransaction({ hash: transactionHash });
      return mapViemTransactionToCoreSubmission(tx);
    } catch {
      return null;
    }
  }

  public async call(transactionRequest: TransactionRequest): Promise<Hex> {
    const { data } = await this.client.call({
      account: transactionRequest.from,
      to: transactionRequest.to,
      data: transactionRequest.data,
      value: transactionRequest.value,
      gas: transactionRequest.gasLimit,
      maxFeePerGas: transactionRequest.maxFeePerGas,
      maxPriorityFeePerGas: transactionRequest.maxPriorityFeePerGas,
    });
    return data ?? "0x";
  }

  public async getFees(): Promise<GasFees> {
    const fees = await this.client.estimateFeesPerGas();
    return {
      maxFeePerGas: fees.maxFeePerGas,
      maxPriorityFeePerGas: fees.maxPriorityFeePerGas,
    };
  }
}
