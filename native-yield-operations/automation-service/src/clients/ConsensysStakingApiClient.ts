import { ApolloClient } from "@apollo/client";
import { IBeaconNodeAPIClient, ILogger, IRetryService, ONE_GWEI, safeSub } from "@consensys/linea-shared-utils";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient.js";
import { ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY } from "../core/entities/graphql/ActiveValidatorsByLargestBalance.js";
import { ValidatorBalance, ValidatorBalanceWithPendingWithdrawal } from "../core/entities/ValidatorBalance.js";

export class ConsensysStakingApiClient implements IValidatorDataClient {
  constructor(
    private readonly logger: ILogger,
    private readonly retryService: IRetryService,
    private readonly apolloClient: ApolloClient,
    private readonly beaconNodeApiClient: IBeaconNodeAPIClient,
  ) {}

  async getActiveValidators(): Promise<ValidatorBalance[] | undefined> {
    const { data, error } = await this.retryService.retry(() =>
      this.apolloClient.query({ query: ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY }),
    );
    if (error) {
      this.logger.error("getActiveValidators error:", { error });
      return undefined;
    }
    if (!data) {
      this.logger.error("getActiveValidators data undefined");
      return undefined;
    }
    const resp = data?.allValidators.nodes;
    this.logger.debug("getActiveValidators succeded", { resp });
    return resp;
  }

  // Return sorted in descending order of withdrawableValue
  async getActiveValidatorsWithPendingWithdrawals(): Promise<ValidatorBalanceWithPendingWithdrawal[] | undefined> {
    const [allValidators, pendingWithdrawalsQueue] = await Promise.all([
      this.getActiveValidators(),
      this.beaconNodeApiClient.getPendingPartialWithdrawals(),
    ]);
    if (allValidators === undefined || pendingWithdrawalsQueue === undefined) return undefined;

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

    this.logger.debug("getActiveValidatorsWithPendingWithdrawals return val", { joined });
    return joined;
  }

  // Should be static, but bit tricky to use static and interface together
  getTotalPendingPartialWithdrawalsWei(validatorList: ValidatorBalanceWithPendingWithdrawal[]): bigint {
    const totalGwei = validatorList.reduce((acc, v) => acc + v.pendingWithdrawalAmount, 0n);
    const totalWei = totalGwei * ONE_GWEI;
    this.logger.debug(`getTotalPendingPartialWithdrawalsWei totalWei=${totalWei}`);
    return totalWei;
  }
}
