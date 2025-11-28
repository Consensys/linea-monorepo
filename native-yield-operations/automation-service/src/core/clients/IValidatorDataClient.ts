import { ValidatorBalance, ValidatorBalanceWithPendingWithdrawal } from "../entities/ValidatorBalance.js";

export interface IValidatorDataClient {
  getActiveValidators(): Promise<ValidatorBalance[] | undefined>;
  getActiveValidatorsWithPendingWithdrawalsAscending(): Promise<ValidatorBalanceWithPendingWithdrawal[] | undefined>;
  getTotalPendingPartialWithdrawalsWei(validatorList: ValidatorBalanceWithPendingWithdrawal[]): bigint;
}
