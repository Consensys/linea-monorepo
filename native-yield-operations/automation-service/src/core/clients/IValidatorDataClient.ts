import { PendingPartialWithdrawal } from "@consensys/linea-shared-utils";
import { ValidatorBalance, ValidatorBalanceWithPendingWithdrawal } from "../entities/ValidatorBalance.js";

export interface IValidatorDataClient {
  getActiveValidators(): Promise<ValidatorBalance[] | undefined>;
  getActiveValidatorsWithPendingWithdrawalsAscending(): Promise<ValidatorBalanceWithPendingWithdrawal[] | undefined>;
  joinValidatorsWithPendingWithdrawals(
    allValidators: ValidatorBalance[] | undefined,
    pendingWithdrawalsQueue: PendingPartialWithdrawal[] | undefined,
  ): ValidatorBalanceWithPendingWithdrawal[] | undefined;
  getTotalPendingPartialWithdrawalsWei(validatorList: ValidatorBalanceWithPendingWithdrawal[]): bigint;
}
