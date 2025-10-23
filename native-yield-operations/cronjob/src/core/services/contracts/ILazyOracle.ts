import { Address, Hex, WatchContractEventReturnType } from "viem";
export interface ILazyOracle<TransactionReceipt> {
  updateVaultData(params: UpdateVaultDataParams): Promise<TransactionReceipt>;
  simulateUpdateVaultData(params: UpdateVaultDataParams): Promise<void>;
  latestReportData(): Promise<LazyOracleReportData>;
  waitForVaultsReportDataUpdatedEvent(): WaitForVaultsReportDataUpdatedEventReturnType;
}

export interface UpdateVaultDataParams {
  vault: Address;
  totalValue: bigint;
  cumulativeLidoFees: bigint;
  liabilityShares: bigint;
  maxLiabilityShares: bigint;
  slashingReserve: bigint;
  proof: Hex[];
}

export interface WaitForVaultsReportDataUpdatedEventReturnType {
  // Cleanup fn for event subscription
  unwatch: WatchContractEventReturnType;
  // Resolves when event detected
  waitForEvent: Promise<void>;
}

export interface LazyOracleReportData {
  timestamp: bigint;
  refSlot: bigint;
  treeRoot: Hex;
  reportCid: string;
}
