import { parseTransaction, type BlockTag, type Hex, type PublicClient } from "viem";

import { mapViemBlockToCoreBlock, mapViemReceiptToCoreReceipt, mapViemTransactionToCoreSubmission } from "./mappers";
import { GasFees } from "../../../core/clients/blockchain/IGasProvider";
import { IProvider } from "../../../core/clients/blockchain/IProvider";
import { Block, TransactionReceipt, TransactionRequest, TransactionSubmission } from "../../../core/types";

export class ViemProvider implements IProvider {
  constructor(protected readonly client: PublicClient) {}

  public async getTransactionCount(address: string, blockTag: string | number | bigint): Promise<number> {
    return this.client.getTransactionCount({
      address: address as Hex,
      blockTag: blockTag as BlockTag,
    });
  }

  public async getBlockNumber(): Promise<number> {
    const blockNumber = await this.client.getBlockNumber();
    return Number(blockNumber);
  }

  public async getTransactionReceipt(txHash: string): Promise<TransactionReceipt | null> {
    try {
      const receipt = await this.client.getTransactionReceipt({ hash: txHash as Hex });
      return mapViemReceiptToCoreReceipt(receipt);
    } catch {
      return null;
    }
  }

  public async getBlock(blockNumber: number | bigint | string): Promise<Block | null> {
    try {
      let block;
      if (typeof blockNumber === "string") {
        block = await this.client.getBlock({ blockTag: blockNumber as BlockTag });
      } else {
        block = await this.client.getBlock({ blockNumber: BigInt(blockNumber) });
      }
      return mapViemBlockToCoreBlock(block);
    } catch {
      return null;
    }
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  public async send(methodName: string, params: Array<any> | Record<string, any>): Promise<any> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return this.client.request({ method: methodName as any, params: params as any });
  }

  public async estimateGas(transactionRequest: TransactionRequest): Promise<bigint> {
    return this.client.estimateGas({
      account: transactionRequest.from as Hex | undefined,
      to: transactionRequest.to as Hex,
      data: transactionRequest.data as Hex | undefined,
      value: transactionRequest.value,
      gas: transactionRequest.gasLimit,
      maxFeePerGas: transactionRequest.maxFeePerGas,
      maxPriorityFeePerGas: transactionRequest.maxPriorityFeePerGas,
    });
  }

  public async getTransaction(transactionHash: string): Promise<TransactionSubmission | null> {
    try {
      const tx = await this.client.getTransaction({ hash: transactionHash as Hex });
      return mapViemTransactionToCoreSubmission(tx);
    } catch {
      return null;
    }
  }

  public async broadcastTransaction(signedTx: string): Promise<TransactionSubmission> {
    const serializedTransaction = signedTx as Hex;
    const parsed = parseTransaction(serializedTransaction);
    const hash = await this.client.sendRawTransaction({ serializedTransaction });
    return {
      hash,
      nonce: parsed.nonce ?? 0,
      gasLimit: parsed.gas ?? 0n,
      maxFeePerGas: "maxFeePerGas" in parsed ? parsed.maxFeePerGas : undefined,
      maxPriorityFeePerGas: "maxPriorityFeePerGas" in parsed ? parsed.maxPriorityFeePerGas : undefined,
    };
  }

  public async call(transactionRequest: TransactionRequest): Promise<string> {
    const { data } = await this.client.call({
      account: transactionRequest.from as Hex | undefined,
      to: transactionRequest.to as Hex,
      data: transactionRequest.data as Hex | undefined,
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
