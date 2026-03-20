import type { Address, Hash, Hex } from "./hex";

export type Log = {
  address: Address;
  topics: Hex[];
  data: Hex;
  blockNumber: number;
  transactionHash: Hash;
  logIndex: number;
};
