import { PendingPartialWithdrawal } from "@consensys/linea-shared-utils";
import {
  AggregatedPendingWithdrawal,
  ExitingValidator,
  ValidatorBalance,
  ValidatorBalanceWithPendingWithdrawal,
} from "../entities/ValidatorBalance.js";

export interface IValidatorDataClient {
  getActiveValidators(): Promise<ValidatorBalance[] | undefined>;
  getExitingValidators(): Promise<ExitingValidator[] | undefined>;
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
  getTotalValidatorBalanceGwei(validators: ValidatorBalance[] | undefined): bigint | undefined;
}
