import { getMessageSentEvents as sdkGetMessageSentEvents } from "@consensys/linea-sdk-viem";
import { type PublicClient, type Hex, type Address, type Client, parseAbiItem } from "viem";

import type { IL2LogClient, MessageSentEventFilters } from "../../../domain/ports/ILogClient";
import type { MessageSent, ServiceVersionMigrated } from "../../../domain/types";

const SERVICE_VERSION_MIGRATED_EVENT = parseAbiItem("event ServiceVersionMigrated(uint256 indexed version)");

export class ViemL2LogClient implements IL2LogClient {
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

  public async getMessageSentEventsByMessageHash(params: {
    messageHash: string;
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]> {
    return this.getMessageSentEvents({
      filters: { messageHash: params.messageHash },
      fromBlock: params.fromBlock,
      toBlock: params.toBlock,
      fromBlockLogIndex: params.fromBlockLogIndex,
    });
  }

  public async getMessageSentEventsByBlockRange(fromBlock: number, toBlock: number): Promise<MessageSent[]> {
    return this.getMessageSentEvents({
      fromBlock,
      toBlock,
    });
  }

  public async getServiceVersionMigratedEvents(params?: {
    fromBlock?: number;
    toBlock?: string | number;
    fromBlockLogIndex?: number;
  }): Promise<ServiceVersionMigrated[]> {
    const logs = await this.publicClient.getLogs({
      address: this.contractAddress,
      event: SERVICE_VERSION_MIGRATED_EVENT,
      fromBlock: params?.fromBlock !== undefined ? BigInt(params.fromBlock) : undefined,
      toBlock:
        params?.toBlock === "latest" ? "latest" : params?.toBlock !== undefined ? BigInt(params.toBlock) : undefined,
    });

    let events: ServiceVersionMigrated[] = logs.map((log) => ({
      version: log.args.version!,
      blockNumber: Number(log.blockNumber),
      logIndex: log.logIndex!,
      contractAddress: log.address,
      transactionHash: log.transactionHash!,
    }));

    if (params?.fromBlockLogIndex !== undefined) {
      events = events.filter(
        (e) =>
          e.blockNumber > (params.fromBlock ?? 0) ||
          (e.blockNumber === (params.fromBlock ?? 0) && e.logIndex > params.fromBlockLogIndex!),
      );
    }

    return events;
  }
}
