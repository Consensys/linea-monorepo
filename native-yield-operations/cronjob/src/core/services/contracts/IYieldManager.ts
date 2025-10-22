import { Address } from "viem";
import { LidoStakingVaultWithdrawalParams } from "../../entities/LidoStakingVaultWithdrawalParams";
import { RebalanceRequirement } from "../../entities/RebalanceRequirement";

export interface IYieldManager<TTransactionReceipt> {
  // View calls
  L1_MESSAGE_SERVICE(): Promise<Address>;
  getTotalSystemBalance(): Promise<bigint>;
  getEffectiveTargetWithdrawalReserve(): Promise<bigint>;
  getTargetReserveDeficit(): Promise<bigint>;
  isStakingPaused(yieldProvider: Address): Promise<boolean>;
  isOssificationInitiated(yieldProvider: Address): Promise<boolean>;
  isOssified(yieldProvider: Address): Promise<boolean>;
  withdrawableValue(yieldProvider: Address): Promise<bigint>;
  // Mutator calls
  fundYieldProvider(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt | null>;
  transferFundsToReserve(amount: bigint): Promise<TTransactionReceipt | null>;
  reportYield(yieldProvider: Address, l2YieldRecipient: Address): Promise<TTransactionReceipt | null>;
  unstake(
    yieldProvider: Address,
    withdrawalParams: LidoStakingVaultWithdrawalParams,
  ): Promise<TTransactionReceipt | null>;
  withdrawFromYieldProvider(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt | null>;
  addToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt | null>;
  safeAddToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TTransactionReceipt | null>;
  pauseStaking(yieldProvider: Address): Promise<TTransactionReceipt | null>;
  unpauseStaking(yieldProvider: Address): Promise<TTransactionReceipt | null>;
  progressPendingOssification(yieldProvider: Address): Promise<TTransactionReceipt | null>;
  // Utility methods
  getRebalanceRequirements(): Promise<RebalanceRequirement>;
}
