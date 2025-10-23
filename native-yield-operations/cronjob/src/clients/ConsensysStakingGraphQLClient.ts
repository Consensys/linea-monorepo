import { ApolloClient } from "@apollo/client";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient";
import { ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY } from "../core/entities/graphql/ActiveValidatorsByLargestBalance";
import { ILogger } from "ts-libs/linea-shared-utils";
import { ValidatorBalance, ValidatorBalanceWithPendingWithdrawal } from "../core/entities/ValidatorBalance";
import { IBeaconNodeAPIClient, ONE_GWEI, safeSub } from "ts-libs/linea-shared-utils/src";

export class ConsensysStakingGraphQLClient implements IValidatorDataClient {
  constructor(
    private readonly apolloClient: ApolloClient,
    private readonly beaconNodeApiClient: IBeaconNodeAPIClient,
    private readonly logger: ILogger,
  ) {}

  async getActiveValidators(): Promise<ValidatorBalance[]> {
    const { data, error } = await this.apolloClient.query({ query: ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY });
    if (error) {
      this.logger.error("getActiveValidatorsByLargestBalances error:", error);
      return [];
    }
    if (data === undefined) {
      this.logger.error("getActiveValidatorsByLargestBalances error:");
      return [];
    }
    return data?.allValidators.nodes;
  }

  // Return sorted in descending order of withdrawableValue
  async getActiveValidatorsWithPendingWithdrawals(): Promise<ValidatorBalanceWithPendingWithdrawal[]> {
    const [allValidators, pendingWithdrawalsQueue] = await Promise.all([
      this.getActiveValidators(),
      this.beaconNodeApiClient.getPendingPartialWithdrawals(),
    ]);

    // 1️⃣ Aggregate duplicate pending withdrawals by validator index
    const pendingByValidator = new Map<number, bigint>();
    for (const w of pendingWithdrawalsQueue) {
      const current = pendingByValidator.get(w.validator_index) ?? 0n;
      pendingByValidator.set(w.validator_index, current + w.amount);
    }

    // 2️⃣ Join with validators and compute total pending amount
    const joined = allValidators.map((v) => {
      const pendingAmount = pendingByValidator.get(Number(v.validatorIndex)) ?? 0n;

      return {
        balance: v.balance,
        effectiveBalance: v.effectiveBalance,
        publicKey: v.publicKey,
        validatorIndex: v.validatorIndex,
        pendingWithdrawalAmount: pendingAmount,
        withdrawableAmount: safeSub(safeSub(v.balance, pendingAmount), ONE_GWEI * 32n),
      };
    });

    // ✅ Sort descending (largest withdrawableAmount first)
    joined.sort((a, b) =>
      a.withdrawableAmount > b.withdrawableAmount ? -1 : a.withdrawableAmount < b.withdrawableAmount ? 1 : 0,
    );

    return joined;
  }

  // Should be static, but bit tricky to use static and interface together
  getTotalPendingPartialWithdrawalsWei(validatorList: ValidatorBalanceWithPendingWithdrawal[]): bigint {
    const totalGwei = validatorList.reduce((acc, v) => acc + v.pendingWithdrawalAmount, 0n);
    const totalWei = totalGwei * ONE_GWEI;
    return totalWei;
  }
}
