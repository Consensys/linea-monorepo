import { Address } from "abitype";
import { Hex } from "./misc";

export type Log<TUnit = bigint> = {
  address: Address;
  topics: Hex[];
  data: Hex;
  blockNumber: TUnit;
  transactionHash: Hex | null;
  transactionIndex: number | null;
  blockHash: Hex;
  logIndex: number | null;
  removed: boolean;
};

export type TransactionReceipt<TUnit = bigint> = {
  blockHash: Hex;
  blockNumber: TUnit;
  contractAddress: Address | null | undefined;
  cumulativeGasUsed: TUnit;
  effectiveGasPrice: TUnit;
  from: Address;
  gasUsed: TUnit;
  logs: Log[];
  logsBloom: Hex;
  status: string;
  to: Address | null;
  transactionHash: Hex;
  transactionIndex: number;
  type: string;
};
