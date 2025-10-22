export interface IBeaconNodeAPIClient {
  getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[]>;
}

export interface PendingPartialWithdrawal {
  validator_index: number;
  amount: bigint;
  withdrawable_epoch: number;
}
