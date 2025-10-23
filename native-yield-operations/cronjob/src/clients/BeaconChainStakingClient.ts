import { min, ONE_GWEI, safeSub } from "ts-libs/linea-shared-utils/src";
import { IBeaconChainStakingClient } from "../core/clients/IBeaconChainStakingClient";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient";
import { ValidatorBalanceWithPendingWithdrawal } from "../core/entities/ValidatorBalance";
import { WithdrawalRequests } from "../core/entities/LidoStakingVaultWithdrawalParams";
import { Address, maxUint256, stringToHex, TransactionReceipt } from "viem";
import { IYieldManager } from "../core/services/contracts/IYieldManager";

export class BeaconChainStakingClient implements IBeaconChainStakingClient {
  constructor(
    private readonly validatorDataClient: IValidatorDataClient,
    private readonly maxValidatorWithdrawalRequestsPerTransaction: number,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly yieldProvider: Address,
  ) {}

  async submitWithdrawalRequestsToFulfilAmount(amountWei: bigint): Promise<void> {
    const sortedValidatorList = await this.validatorDataClient.getActiveValidatorsWithPendingWithdrawals();
    const totalPendingPartialWithdrawalsWei =
      this.validatorDataClient.getTotalPendingPartialWithdrawalsWei(sortedValidatorList);
    const requiredWithdrawalAmountWei = safeSub(amountWei, totalPendingPartialWithdrawalsWei);
    await this._submitPartialWithdrawalRequests(sortedValidatorList, requiredWithdrawalAmountWei);
  }

  async submitMaxAvailableWithdrawalRequests(): Promise<void> {
    const sortedValidatorList = await this.validatorDataClient.getActiveValidatorsWithPendingWithdrawals();
    const remainingWithdrawals = await this._submitPartialWithdrawalRequests(sortedValidatorList, maxUint256);
    await this._submitValidatorExits(sortedValidatorList, remainingWithdrawals);
  }

  // Returns # of withdrawal requests remaining
  private async _submitPartialWithdrawalRequests(
    sortedValidatorList: ValidatorBalanceWithPendingWithdrawal[],
    amountWei: bigint,
  ): Promise<number> {
    const withdrawalRequests: WithdrawalRequests = {
      pubkeys: [],
      amountsGwei: [],
    };
    let totalWithdrawalRequestAmountWei = 0n;

    for (const v of sortedValidatorList) {
      if (withdrawalRequests.pubkeys.length >= this.maxValidatorWithdrawalRequestsPerTransaction) break;
      if (totalWithdrawalRequestAmountWei >= amountWei) break;

      const remainingWei = amountWei - totalWithdrawalRequestAmountWei;
      const withdrawableWei = v.withdrawableAmount * ONE_GWEI;
      const amountToWithdrawWei = min(withdrawableWei, remainingWei);
      const amountToWithdrawGwei = amountToWithdrawWei / ONE_GWEI;

      if (amountToWithdrawGwei > 0n) {
        withdrawalRequests.pubkeys.push(stringToHex(v.publicKey));
        withdrawalRequests.amountsGwei.push(amountToWithdrawGwei);
        totalWithdrawalRequestAmountWei += amountToWithdrawWei;
      }
    }

    // Do unstake
    await this.yieldManagerContractClient.unstake(this.yieldProvider, withdrawalRequests);

    // Return # of remaining shots
    return this.maxValidatorWithdrawalRequestsPerTransaction - withdrawalRequests.pubkeys.length;
  }

  private async _submitValidatorExits(
    sortedValidatorList: ValidatorBalanceWithPendingWithdrawal[],
    remainingWithdrawals: number,
  ): Promise<void> {
    const withdrawalRequests: WithdrawalRequests = {
      pubkeys: [],
      amountsGwei: [],
    };

    for (const v of sortedValidatorList) {
      if (withdrawalRequests.pubkeys.length >= remainingWithdrawals) break;
      if (withdrawalRequests.pubkeys.length >= this.maxValidatorWithdrawalRequestsPerTransaction) break;

      if (v.withdrawableAmount === 0n) {
        withdrawalRequests.pubkeys.push(stringToHex(v.publicKey));
        // 0 amount -> signal for validator exit
        withdrawalRequests.amountsGwei.push(0n);
      }
    }

    // Do unstake
    await this.yieldManagerContractClient.unstake(this.yieldProvider, withdrawalRequests);
  }
}
