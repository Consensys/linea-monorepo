import { Hex } from "viem";

export interface LazyOracleReportData {
  timestamp: bigint;
  refSlot: bigint;
  treeRoot: Hex;
  reportCid: string;
}
