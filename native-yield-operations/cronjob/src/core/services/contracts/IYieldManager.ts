import { LidoStakingVaultWithdrawalParams } from "../../entities/YieldManager";

export interface IYieldManager<TTransactionReceipt> {
  // View calls
  getTargetReserveDeficit(yieldProvider: string): Promise<bigint>;
  isStakingPaused(yieldProvider: string): Promise<boolean>;
  isOssificationInitiated(yieldProvider: string): Promise<boolean>;
  isOssified(yieldProvider: string): Promise<boolean>;
  withdrawableValue(yieldProvider: string): Promise<bigint>;
  // // Mutator calls
  // fundYieldProvider(yieldProvider: string, amount: bigint): Promise<TransactionReceipt | null>;
  // transferFundsToReserve(amount: bigint): Promise<TransactionReceipt | null>;
  // reportYield(yieldProvider: string, l2YieldRecipient: string): Promise<TransactionReceipt | null>;
  // unstake(
  //   yieldProvider: string,
  //   withdrawalParams: LidoStakingVaultWithdrawalParams,
  // ): Promise<TransactionReceipt | null>;
  // withdrawFromYieldProvider(yieldProvider: string, amount: bigint): Promise<TransactionReceipt | null>;
  // addToWithdrawalReserve(yieldProvider: string, amount: bigint): Promise<TransactionReceipt | null>;
  // pauseStaking(yieldProvider: string, amount: bigint): Promise<TransactionReceipt | null>;
  // unpauseStaking(yieldProvider: string, amount: bigint): Promise<TransactionReceipt | null>;
  // progressPendingOssification(yieldProvider: string): Promise<TransactionReceipt | null>;
}
