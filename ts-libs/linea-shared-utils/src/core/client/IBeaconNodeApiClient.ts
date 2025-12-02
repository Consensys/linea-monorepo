// https://ethereum.github.io/beacon-APIs/

export interface IBeaconNodeAPIClient {
  getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[] | undefined>;
}

export interface BeaconApiResponse {
  execution_optimistic: boolean;
  finalized: boolean;
}

export interface PendingPartialWithdrawalResponse extends BeaconApiResponse {
  data: PendingPartialWithdrawal[];
}

export interface PendingPartialWithdrawal {
  validator_index: number;
  amount: bigint;
  withdrawable_epoch: number;
}
