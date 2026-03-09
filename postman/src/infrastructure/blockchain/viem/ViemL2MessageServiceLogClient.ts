import { type Hex, type PublicClient } from "viem";
import { getContractEvents } from "viem/actions";

import {
  IL2MessageServiceLogClient,
  MessageSentEventFilters,
} from "../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";
import { MessageSent } from "../../../core/types";
import { L2MessageServiceAbi } from "../abis/L2MessageServiceAbi";

type BlockParam = bigint | "latest" | "earliest" | "pending" | "safe" | "finalized";

function toBlockParam(block: number | string | undefined, fallback: BlockParam): BlockParam {
  if (block === undefined) return fallback;
  if (typeof block === "number") return BigInt(block);
  return block as BlockParam;
}

export class ViemL2MessageServiceLogClient implements IL2MessageServiceLogClient {
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
      abi: L2MessageServiceAbi,
      eventName: "MessageSent",
      args: {
        _from: params.filters?.from as Hex | undefined,
        _to: params.filters?.to as Hex | undefined,
        _messageHash: params.filters?.messageHash as Hex | undefined,
      },
      fromBlock: toBlockParam(params.fromBlock, "earliest"),
      toBlock: toBlockParam(params.toBlock, "latest"),
    });

    return events
      .filter((e) => !e.removed && e.blockNumber !== null && e.logIndex !== null)
      .filter((e) => {
        if (params.fromBlockLogIndex === undefined) return true;
        const sameBlock = Number(e.blockNumber) === params.fromBlock;
        return !sameBlock || e.logIndex! >= params.fromBlockLogIndex;
      })
      .map((e) => {
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
}
