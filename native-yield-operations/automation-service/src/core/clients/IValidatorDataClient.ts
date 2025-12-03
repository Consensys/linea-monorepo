import { PendingPartialWithdrawal } from "@consensys/linea-shared-utils";
import {
  AggregatedPendingWithdrawal,
  ValidatorBalance,
  ValidatorBalanceWithPendingWithdrawal,
} from "../entities/ValidatorBalance.js";

export interface IValidatorDataClient {
  getActiveValidators(): Promise<ValidatorBalance[] | undefined>;
  getActiveValidatorsWithPendingWithdrawalsAscending(): Promise<ValidatorBalanceWithPendingWithdrawal[] | undefined>;
  joinValidatorsWithPendingWithdrawals(
    allValidators: ValidatorBalance[] | undefined,
    pendingWithdrawalsQueue: PendingPartialWithdrawal[] | undefined,
  ): ValidatorBalanceWithPendingWithdrawal[] | undefined;
  getFilteredAndAggregatedPendingWithdrawals(
    allValidators: ValidatorBalance[] | undefined,
    pendingWithdrawalsQueue: PendingPartialWithdrawal[] | undefined,
  ): AggregatedPendingWithdrawal[] | undefined;
  getTotalPendingPartialWithdrawalsWei(validatorList: ValidatorBalanceWithPendingWithdrawal[]): bigint;
}
