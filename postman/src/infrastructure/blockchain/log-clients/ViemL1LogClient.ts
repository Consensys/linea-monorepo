import { getMessageSentEvents as sdkGetMessageSentEvents } from "@consensys/linea-sdk-viem";
import { type PublicClient, type Hex, type Address, BlockTag } from "viem";

import type { ILogClient, MessageSentEventFilters } from "../../../domain/ports/ILogClient";
import type { MessageSent } from "../../../domain/types";

export class ViemL1LogClient implements ILogClient {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly contractAddress: Address,
  ) {}

  public async getMessageSentEvents(params: {
    filters?: MessageSentEventFilters;
    fromBlock?: bigint;
    toBlock?: BlockTag | bigint;
    fromBlockLogIndex?: number;
  }): Promise<MessageSent[]> {
    const events = await sdkGetMessageSentEvents(this.publicClient, {
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
}
