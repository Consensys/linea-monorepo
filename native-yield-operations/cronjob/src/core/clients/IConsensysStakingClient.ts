export interface IConsensysStakingClient {
  // i.) Call GraphQL API
  // ii.) Get list of biggest n -> provide params for YieldManager.unstake()
  getActiveValidatorsByLargestBalances(): Promise<void>;
  getPendingPartialWithdrawals(): Promise<void>;
  getPendingValidatorExits(): Promise<void>;
}