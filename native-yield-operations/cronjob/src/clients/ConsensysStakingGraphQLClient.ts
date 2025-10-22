import { ApolloClient } from "@apollo/client";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient";
import { ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY } from "../core/entities/graphql/ActiveValidatorsByLargestBalance";
import { ILogger } from "ts-libs/linea-shared-utils";
import { ValidatorBalance } from "../core/entities/ValidatorBalance";

export class ConsensysStakingGraphQLClient implements IValidatorDataClient {
  constructor(
    private readonly apolloClient: ApolloClient,
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
  //   getPendingPartialWithdrawals(): Promise<void>;
  //   getPendingValidatorExits(): Promise<void>;
}
