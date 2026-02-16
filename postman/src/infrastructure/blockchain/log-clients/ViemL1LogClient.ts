import { getMessageSentEvents as sdkGetMessageSentEvents } from "@consensys/linea-sdk-viem";
import { type PublicClient, type Hex, type Address, type Client, parseAbiItem } from "viem";

import type {
  IL1LogClient,
  MessageSentEventFilters,
  L2MessagingBlockAnchoredFilters,
  MessageClaimedFilters,
} from "../../../domain/ports/ILogClient";
import type { MessageSent, L2MessagingBlockAnchored, MessageClaimed } from "../../../domain/types";

const L2_MESSAGING_BLOCKS_ANCHORED_EVENT = parseAbiItem("event L2MessagingBlocksAnchored(uint256 indexed l2Block)");

const MESSAGE_CLAIMED_EVENT = parseAbiItem("event MessageClaimed(bytes32 indexed _messageHash)");

export class ViemL1LogClient implements IL1LogClient {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly contractAddress: Address,
  ) {}

  public async getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]> {
    const events = await sdkGetMessageSentEvents(this.publicClient as Client, {
      address: this.contractAddress,
      fromBlock: params.fromBlock !== undefined ? BigInt(params.fromBlock) : undefined,
      toBlock:
        params.toBlock === "latest" ? "latest" : params.toBlock !== undefined ? BigInt(params.toBlock) : undefined,
      args: {
        _from: params.filters?.from as Address | undefined,
        _to: params.filters?.to as Address | undefined,
        _messageHash: params.filters?.messageHash as Hex | undefined,
      },
    });

    let result: MessageSent[] = events.map((e) => ({
      messageSender: e.messageSender,
      destination: e.destination,
      fee: e.fee,
      value: e.value,
      messageNonce: e.messageNonce,
      calldata: e.calldata,
      messageHash: e.messageHash,
      blockNumber: Number(e.blockNumber),
      logIndex: e.logIndex,
      contractAddress: e.contractAddress,
      transactionHash: e.transactionHash,
    }));

    if (params.fromBlockLogIndex !== undefined) {
      result = result.filter(
        (e) =>
          e.blockNumber > (params.fromBlock ?? 0) ||
          (e.blockNumber === (params.fromBlock ?? 0) && e.logIndex > params.fromBlockLogIndex!),
      );
    }

    return result;
  }

  public async getL2MessagingBlockAnchoredEvents(params: {
    filters?: L2MessagingBlockAnchoredFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<L2MessagingBlockAnchored[]> {
    const logs = await this.publicClient.getLogs({
      address: this.contractAddress,
      event: L2_MESSAGING_BLOCKS_ANCHORED_EVENT,
      fromBlock: params.fromBlock !== undefined ? BigInt(params.fromBlock) : undefined,
      toBlock:
        params.toBlock === "latest" ? "latest" : params.toBlock !== undefined ? BigInt(params.toBlock) : undefined,
      args: {
        l2Block: params.filters?.l2Block,
      },
    });

    let events: L2MessagingBlockAnchored[] = logs.map((log) => ({
      l2Block: log.args.l2Block!,
      blockNumber: Number(log.blockNumber),
      logIndex: log.logIndex!,
      contractAddress: log.address,
      transactionHash: log.transactionHash!,
    }));

    if (params.fromBlockLogIndex !== undefined) {
      events = events.filter(
        (e) =>
          e.blockNumber > (params.fromBlock ?? 0) ||
          (e.blockNumber === (params.fromBlock ?? 0) && e.logIndex > params.fromBlockLogIndex!),
      );
    }

    return events;
  }

  public async getMessageClaimedEvents(params: {
    filters?: MessageClaimedFilters;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageClaimed[]> {
    const logs = await this.publicClient.getLogs({
      address: this.contractAddress,
      event: MESSAGE_CLAIMED_EVENT,
      fromBlock: params.fromBlock !== undefined ? BigInt(params.fromBlock) : undefined,
      toBlock:
        params.toBlock === "latest" ? "latest" : params.toBlock !== undefined ? BigInt(params.toBlock) : undefined,
      args: {
        _messageHash: params.filters?.messageHash as Hex | undefined,
      },
    });

    let events: MessageClaimed[] = logs.map((log) => ({
      messageHash: log.args._messageHash!,
      blockNumber: Number(log.blockNumber),
      logIndex: log.logIndex!,
      contractAddress: log.address,
      transactionHash: log.transactionHash!,
    }));

    if (params.fromBlockLogIndex !== undefined) {
      events = events.filter(
        (e) =>
          e.blockNumber > (params.fromBlock ?? 0) ||
          (e.blockNumber === (params.fromBlock ?? 0) && e.logIndex > params.fromBlockLogIndex!),
      );
    }

    return events;
  }
}
