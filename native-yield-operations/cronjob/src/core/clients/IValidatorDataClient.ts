import { ValidatorBalance, ValidatorBalanceWithPendingWithdrawal } from "../entities/ValidatorBalance";

export interface IValidatorDataClient {
  getActiveValidators(): Promise<ValidatorBalance[]>;
  getActiveValidatorsWithPendingWithdrawals(): Promise<ValidatorBalanceWithPendingWithdrawal[]>;
  getTotalPendingPartialWithdrawalsWei(validatorList: ValidatorBalanceWithPendingWithdrawal[]): bigint;
}
