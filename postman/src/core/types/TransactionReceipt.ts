import { Log } from "./Log";

export type TransactionReceipt = {
  hash: string;
  blockNumber: number;
  status: "success" | "reverted";
  gasUsed: bigint;
  gasPrice: bigint;
  logs: Log[];
};
