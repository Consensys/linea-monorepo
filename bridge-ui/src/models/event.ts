import { Address, Log } from "viem";

export interface ETHEvent extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    _from: Address;
    _value: bigint;
    _to: Address;
  };
}

export interface ERC20Event extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    sender: Address;
    token: Address;
    amount: bigint;
  };
}

export interface ERC20V2Event extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    sender: Address;
    recipient: Address;
    token: Address;
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
