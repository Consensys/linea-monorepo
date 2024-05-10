import { Address, Log } from 'viem';

export interface ETHEvent extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    _value: bigint;
    _to: Address;
  };
}

export interface ERC20Event extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    amount: bigint;
    recipient: Address;
    token: Address;
  };
}

export interface USDCEvent extends Log {
  blockNumber: bigint;
  transactionHash: Address;
  args: {
    amount: bigint;
    to: Address;
  };
}
