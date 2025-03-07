import { Address, Log } from "viem";

export type MessageSentEvent = Log & {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    _from: Address;
    _to: Address;
    _fee: bigint;
    _value: bigint;
    _nonce: bigint;
    _calldata: string;
    _messageHash: string;
  };
};

export type BridgingInitiatedV2Event = Log & {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    sender: Address;
    recipient: Address;
    token: Address;
    amount: bigint;
  };
};
