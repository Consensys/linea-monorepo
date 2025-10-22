import { WithdrawalRequests } from "../entities/LidoStakingVaultWithdrawalParams";

export interface IBeaconChainStakingClient {
  submitWithdrawalRequestsToFulfilAmount(amountWei: bigint): Promise<void>;
  submitMaxAvailableWithdrawalRequests(): Promise<void>;

  // i.) Call GraphQL API
  // ii.) Get list of biggest n -> provide params for YieldManager.unstake()
  // iii.) Subtract pending withdrawals from active balances
  getActiveValidatorsByLargestBalances(): Promise<void>;
  getWithdrawalRequestsToFulfilAmount(amountWei: bigint): Promise<WithdrawalRequests>;
  getPendingPartialWithdrawals(): Promise<void>;
  getPendingValidatorExits(): Promise<void>;
}
