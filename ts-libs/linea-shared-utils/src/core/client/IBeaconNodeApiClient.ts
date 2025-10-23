export interface IBeaconNodeAPIClient {
  getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[]>;
}

export interface PendingPartialWithdrawalResponse {
  execution_optimistic: boolean;
  finalized: boolean;
  data: PendingPartialWithdrawal[];
}

export interface PendingPartialWithdrawal {
  validator_index: number;
  amount: bigint;
  withdrawable_epoch: number;
}
