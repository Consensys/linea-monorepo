import { Address, Hex, WatchContractEventReturnType } from "viem";
import { LazyOracleReportData } from "../../entities";
export interface ILazyOracle<TransactionReceipt> {
  updateVaultData(
    vault: Address,
    totalValue: bigint,
    cumulativeLidoFees: bigint,
    liabilityShares: bigint,
    maxLiabilityShares: bigint,
    slashingReserve: bigint,
    proof: Hex[],
  ): Promise<TransactionReceipt | null>;
  latestReportData(): Promise<LazyOracleReportData>;
  waitForVaultsReportDataUpdatedEvent(): WaitForVaultsReportDataUpdatedEventReturnType;
}

export interface WaitForVaultsReportDataUpdatedEventReturnType {
  // Cleanup fn for event subscription
  unwatch: WatchContractEventReturnType;
  // Resolves when event detected
  waitForEvent: Promise<void>;
}
