import { ValidatorBalance, ValidatorBalanceWithPendingWithdrawal } from "../entities/ValidatorBalance.js";

export interface IValidatorDataClient {
  getActiveValidators(): Promise<ValidatorBalance[]>;
  getActiveValidatorsWithPendingWithdrawals(): Promise<ValidatorBalanceWithPendingWithdrawal[]>;
  getTotalPendingPartialWithdrawalsWei(validatorList: ValidatorBalanceWithPendingWithdrawal[]): bigint;
}
