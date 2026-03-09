import { type Hex, type PublicClient } from "viem";
import { getContractEvents } from "viem/actions";

import {
  ILineaRollupLogClient,
  L2MessagingBlockAnchoredFilters,
  MessageClaimedFilters,
  MessageSentEventFilters,
} from "../../../core/clients/blockchain/ethereum/ILineaRollupLogClient";
import { MessageSent } from "../../../core/types";
import { LineaRollupAbi } from "../abis/LineaRollupAbi";

type BlockParam = bigint | "latest" | "earliest" | "pending" | "safe" | "finalized";

function toBlockParam(block: number | string | undefined, fallback: BlockParam): BlockParam {
  if (block === undefined) return fallback;
  if (typeof block === "number") return BigInt(block);
  return block as BlockParam;
}

export class ViemLineaRollupLogClient implements ILineaRollupLogClient {
  constructor(
    private readonly client: PublicClient,
    private readonly contractAddress: string,
  ) {}

  public async getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]> {
    const events = await getContractEvents(this.client, {
      address: this.contractAddress as Hex,
      abi: LineaRollupAbi,
      eventName: "MessageSent",
      args: {
        _from: params.filters?.from as Hex | undefined,
        _to: params.filters?.to as Hex | undefined,
        _messageHash: params.filters?.messageHash as Hex | undefined,
      },
      fromBlock: toBlockParam(params.fromBlock, "earliest"),
      toBlock: toBlockParam(params.toBlock, "latest"),
    });

    const filtered = events
      .filter((e) => !e.removed && e.blockNumber !== null && e.logIndex !== null)
      .filter((e) => {
        if (params.fromBlockLogIndex === undefined) return true;
        const sameBlock = Number(e.blockNumber) === params.fromBlock;
        return !sameBlock || e.logIndex! >= params.fromBlockLogIndex;
      });

    return filtered.map((e) => {
      const args = e.args as {
        _messageHash: string;
        _from: string;
        _to: string;
        _fee: bigint;
        _value: bigint;
        _nonce: bigint;
        _calldata: string;
      };
      return {
        messageHash: args._messageHash,
        messageSender: args._from,
        destination: args._to,
        fee: args._fee,
        value: args._value,
        messageNonce: args._nonce,
        calldata: args._calldata,
        contractAddress: this.contractAddress,
        blockNumber: Number(e.blockNumber),
        transactionHash: e.transactionHash!,
        logIndex: e.logIndex!,
      };
    });
  }

  public async getL2MessagingBlockAnchoredEvents(params: {
    filters?: L2MessagingBlockAnchoredFilters;
    fromBlock?: number;
    toBlock?: string | number;
  }): Promise<{ l2Block: bigint; blockNumber: number; transactionHash: string; logIndex: number }[]> {
    const events = await getContractEvents(this.client, {
      address: this.contractAddress as Hex,
      abi: LineaRollupAbi,
      eventName: "L2MessagingBlocksAnchored",
      args: params.filters?.l2Block !== undefined ? { l2Block: params.filters.l2Block } : undefined,
      fromBlock: toBlockParam(params.fromBlock, "earliest"),
      toBlock: toBlockParam(params.toBlock, "latest"),
    });

    return events
      .filter((e) => !e.removed && e.blockNumber !== null && e.logIndex !== null)
      .map((e) => {
        const args = e.args as { l2Block: bigint };
        return {
          l2Block: args.l2Block,
          blockNumber: Number(e.blockNumber),
          transactionHash: e.transactionHash!,
          logIndex: e.logIndex!,
        };
      });
  }

  public async getMessageClaimedEvents(params: {
    filters?: MessageClaimedFilters;
    fromBlock?: number;
    toBlock?: string | number;
  }): Promise<{ messageHash: string; blockNumber: number; transactionHash: string; logIndex: number }[]> {
    const events = await getContractEvents(this.client, {
      address: this.contractAddress as Hex,
      abi: LineaRollupAbi,
      eventName: "MessageClaimed",
      args: params.filters?.messageHash !== undefined ? { _messageHash: params.filters.messageHash as Hex } : undefined,
      fromBlock: toBlockParam(params.fromBlock, "earliest"),
      toBlock: toBlockParam(params.toBlock, "latest"),
    });

    return events
      .filter((e) => !e.removed && e.blockNumber !== null && e.logIndex !== null)
      .map((e) => {
        const args = e.args as { _messageHash: string };
        return {
          messageHash: args._messageHash,
          blockNumber: Number(e.blockNumber),
          transactionHash: e.transactionHash!,
          logIndex: e.logIndex!,
        };
      });
  }
}
