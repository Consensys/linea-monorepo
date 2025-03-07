import { Address, Log } from "viem";

export interface MessageSentEvent extends Log {
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
}

export interface BridgingInitiatedV2Event extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    sender: Address;
    recipient: Address;
    token: Address;
    amount: bigint;
  };
}

export interface USDCEvent extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    depositor: Address;
    amount: bigint;
    to: Address;
  };
}
