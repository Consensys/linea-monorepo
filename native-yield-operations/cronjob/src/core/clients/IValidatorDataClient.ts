import { WithdrawalRequests } from "../entities/LidoStakingVaultWithdrawalParams";
import { ValidatorBalance } from "../entities/ValidatorBalance";

export interface IValidatorDataClient {
  // i.) Call GraphQL API
  // ii.) Get list of biggest n -> provide params for YieldManager.unstake()
  // iii.) Subtract pending withdrawals from active balances
  getActiveValidatorsByLargestBalances(): Promise<ValidatorBalance[]>;
  //   getWithdrawalRequestsToFulfilAmount(amountWei: bigint): Promise<WithdrawalRequests>;
  getTotalPendingPartialWithdrawals(): Promise<bigint>;
  //   getPendingValidatorExits(): Promise<void>;
}
