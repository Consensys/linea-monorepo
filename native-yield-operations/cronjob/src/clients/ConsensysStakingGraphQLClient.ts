import { ApolloClient } from "@apollo/client";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient";
import { ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY } from "../core/entities/graphql/ActiveValidatorsByLargestBalance";
import { ILogger } from "ts-libs/linea-shared-utils";
import { ValidatorBalance } from "../core/entities/ValidatorBalance";
import { IBeaconNodeAPIClient, PendingPartialWithdrawal } from "ts-libs/linea-shared-utils/src";

export class ConsensysStakingGraphQLClient implements IValidatorDataClient {
  constructor(
    private readonly apolloClient: ApolloClient,
    private readonly beaconNodeApiClient: IBeaconNodeAPIClient,
    private readonly logger: ILogger,
  ) {}
  async getActiveValidatorsByLargestBalances(): Promise<ValidatorBalance[]> {
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
  //   getWithdrawalRequestsToFulfilAmount(amountWei: bigint): Promise<WithdrawalRequests>;
  async getTotalPendingPartialWithdrawals(): Promise<bigint> {
    const [allValidators, pendingWithdrawalsQueue] = await Promise.all([
      this.getActiveValidatorsByLargestBalances(),
      this.beaconNodeApiClient.getPendingPartialWithdrawals(),
    ]);

    // 1️⃣ Aggregate duplicate pending withdrawals by validator index
    const pendingByValidator = new Map<number, bigint>();
    for (const w of pendingWithdrawalsQueue) {
      const current = pendingByValidator.get(w.validator_index) ?? 0n;
      pendingByValidator.set(w.validator_index, current + w.amount);
    }

    // 2️⃣ Join with validators and compute total pending amount
    let totalPendingWithdrawal = 0n;

    const joined = allValidators.map((v) => {
      const pendingAmount = pendingByValidator.get(Number(v.validatorIndex)) ?? 0n;
      totalPendingWithdrawal += pendingAmount;

      return {
        balance: v.balance,
        effectiveBalance: v.effectiveBalance,
        publicKey: v.publicKey,
        validatorIndex: v.validatorIndex,
        pendingWithdrawalAmount: pendingAmount,
      };
    });
    return totalPendingWithdrawal;
  }

  //   getPendingValidatorExits(): Promise<void>;
}
