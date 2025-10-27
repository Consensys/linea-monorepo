import { Address } from "viem";
import { WithdrawalRequests } from "../../entities/LidoStakingVaultWithdrawalParams.js";
import { RebalanceRequirement } from "../../entities/RebalanceRequirement.js";
import { IBaseContractClient } from "@consensys/linea-shared-utils";
import { YieldReport } from "../../entities/YieldReport.js";

export interface IYieldManager<TTransactionReceipt> extends IBaseContractClient {
  // View calls
  L1_MESSAGE_SERVICE(): Promise<Address>;
  getTotalSystemBalance(): Promise<bigint>;
  getEffectiveTargetWithdrawalReserve(): Promise<bigint>;
  getTargetReserveDeficit(): Promise<bigint>;
  isStakingPaused(yieldProvider: Address): Promise<boolean>;
  isOssificationInitiated(yieldProvider: Address): Promise<boolean>;
  isOssified(yieldProvider: Address): Promise<boolean>;
  withdrawableValue(yieldProvider: Address): Promise<bigint>;
  getYieldProviderData(yieldProvider: Address): Promise<YieldProviderData>;
  // Mutator calls
  fundYieldProvider(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt>;
  transferFundsToReserve(amount: bigint): Promise<TTransactionReceipt>;
  reportYield(yieldProvider: Address, l2YieldRecipient: Address): Promise<TTransactionReceipt>;
  unstake(yieldProvider: Address, withdrawalParams: WithdrawalRequests): Promise<TTransactionReceipt>;
  withdrawFromYieldProvider(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt>;
  addToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt>;
  safeAddToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt>;
  pauseStaking(yieldProvider: Address): Promise<TTransactionReceipt>;
  unpauseStaking(yieldProvider: Address): Promise<TTransactionReceipt>;
  progressPendingOssification(yieldProvider: Address): Promise<TTransactionReceipt>;
  // Utility methods
  getRebalanceRequirements(): Promise<RebalanceRequirement>;
  getLidoStakingVaultAddress(yieldProvider: Address): Promise<Address>;
  getLidoDashboardAddress(yieldProvider: Address): Promise<Address>;
  pauseStakingIfNotAlready(yieldProvider: Address): Promise<TTransactionReceipt | undefined>;
  unpauseStakingIfNotAlready(yieldProvider: Address): Promise<TTransactionReceipt | undefined>;
  getAvailableUnstakingRebalanceBalance(yieldProvider: Address): Promise<bigint>;
  safeAddToWithdrawalReserveIfAboveThreshold(
    yieldProvider: Address,
    amount: bigint,
  ): Promise<TTransactionReceipt | undefined>;
  safeMaxAddToWithdrawalReserve(yieldProvider: Address): Promise<TTransactionReceipt | undefined>;
  getWithdrawalAmountFromTxReceipt(txReceipt: TTransactionReceipt): bigint;
  getYieldReportFromTxReceipt(txReceipt: TTransactionReceipt): YieldReport | undefined;
}

export interface YieldProviderData {
  yieldProviderVendor: number; // enum uint8
  isStakingPaused: boolean;
  isOssificationInitiated: boolean;
  isOssified: boolean;
  primaryEntrypoint: Address;
  ossifiedEntrypoint: Address;
  yieldProviderIndex: bigint; // uint96
  userFunds: bigint; // uint256
  yieldReportedCumulative: bigint; // uint256
  lstLiabilityPrincipal: bigint; // uint256
}
