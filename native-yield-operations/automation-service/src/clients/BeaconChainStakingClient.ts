import { ILogger, min, ONE_GWEI, safeSub } from "@consensys/linea-shared-utils";
import { IBeaconChainStakingClient } from "../core/clients/IBeaconChainStakingClient.js";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient.js";
import { ValidatorBalanceWithPendingWithdrawal } from "../core/entities/ValidatorBalance.js";
import { WithdrawalRequests } from "../core/entities/LidoStakingVaultWithdrawalParams.js";
import { Address, maxUint256, stringToHex, TransactionReceipt } from "viem";
import { IYieldManager } from "../core/clients/contracts/IYieldManager.js";
import { INativeYieldAutomationMetricsUpdater } from "../core/metrics/INativeYieldAutomationMetricsUpdater.js";

export class BeaconChainStakingClient implements IBeaconChainStakingClient {
  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly validatorDataClient: IValidatorDataClient,
    private readonly maxValidatorWithdrawalRequestsPerTransaction: number,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly yieldProvider: Address,
  ) {}

  async submitWithdrawalRequestsToFulfilAmount(amountWei: bigint): Promise<void> {
    this.logger.debug(
      `submitWithdrawalRequestsToFulfilAmount started: amountWei=${amountWei.toString()}; validatorLimit=${this.maxValidatorWithdrawalRequestsPerTransaction}`,
    );
    const sortedValidatorList = await this.validatorDataClient.getActiveValidatorsWithPendingWithdrawals();
    if (sortedValidatorList === undefined) {
      this.logger.error(
        "submitWithdrawalRequestsToFulfilAmount failed to get sortedValidatorList with pending withdrawals",
      );
      return;
    }
    const totalPendingPartialWithdrawalsWei =
      this.validatorDataClient.getTotalPendingPartialWithdrawalsWei(sortedValidatorList);
    const requiredWithdrawalAmountWei = safeSub(amountWei, totalPendingPartialWithdrawalsWei);
    await this._submitPartialWithdrawalRequests(sortedValidatorList, requiredWithdrawalAmountWei);
  }

  async submitMaxAvailableWithdrawalRequests(): Promise<void> {
    this.logger.debug(`submitMaxAvailableWithdrawalRequests started`);
    const sortedValidatorList = await this.validatorDataClient.getActiveValidatorsWithPendingWithdrawals();
    if (sortedValidatorList === undefined) {
      this.logger.error(
        "submitMaxAvailableWithdrawalRequests failed to get sortedValidatorList with pending withdrawals",
      );
      return;
    }
    const remainingWithdrawals = await this._submitPartialWithdrawalRequests(sortedValidatorList, maxUint256);
    await this._submitValidatorExits(sortedValidatorList, remainingWithdrawals);
  }

  // Returns # of withdrawal requests remaining
  private async _submitPartialWithdrawalRequests(
    sortedValidatorList: ValidatorBalanceWithPendingWithdrawal[],
    amountWei: bigint,
  ): Promise<number> {
    this.logger.debug(`_submitPartialWithdrawalRequests started amountWei=${amountWei}`, { sortedValidatorList });
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

    // Instrument metrics after tx success
    for (let i = 0; i < withdrawalRequests.pubkeys.length; i++) {
      const pubkey = withdrawalRequests.pubkeys[i];
      const amountGwei = withdrawalRequests.amountsGwei[i];
      this.metricsUpdater.addValidatorPartialUnstakeAmount(pubkey, Number(amountGwei));
    }

    // Return # of remaining shots
    const remainingWithdrawals = this.maxValidatorWithdrawalRequestsPerTransaction - withdrawalRequests.pubkeys.length;
    this.logger.debug(`_submitPartialWithdrawalRequests remainingWithdrawal=${remainingWithdrawals}`);
    return remainingWithdrawals;
  }

  private async _submitValidatorExits(
    sortedValidatorList: ValidatorBalanceWithPendingWithdrawal[],
    remainingWithdrawals: number,
  ): Promise<void> {
    this.logger.debug(`_submitValidatorExits started remainingWithdrawals=${remainingWithdrawals}`, {
      sortedValidatorList,
    });
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

    // Instrument metrics after tx success
    for (let i = 0; i < withdrawalRequests.pubkeys.length; i++) {
      const pubkey = withdrawalRequests.pubkeys[i];
      this.metricsUpdater.incrementValidatorExit(pubkey);
    }
  }
}
