import { jest } from "@jest/globals";
import { ConsensysStakingApiClient } from "../ConsensysStakingApiClient.js";
import {
  IBeaconNodeAPIClient,
  ILogger,
  IRetryService,
  ONE_GWEI,
  PendingPartialWithdrawal,
  safeSub,
} from "@consensys/linea-shared-utils";
import type { ApolloClient } from "@apollo/client";
import { ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY } from "../../core/entities/graphql/ActiveValidatorsByLargestBalance.js";
import type { ValidatorBalance, ValidatorBalanceWithPendingWithdrawal } from "../../core/entities/ValidatorBalance.js";

const createLoggerMock = (): jest.Mocked<ILogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  warnOrError: jest.fn(),
});

const createClient = () => {
  const logger = createLoggerMock();
  const retryMock = jest.fn(async (fn: () => Promise<unknown>, _timeoutMs?: number) => fn());
  const retryService = { retry: retryMock } as unknown as jest.Mocked<IRetryService>;

  const apolloQueryMock = jest.fn() as jest.MockedFunction<
    (params: { query: unknown }) => Promise<{ data?: unknown; error?: unknown }>
  >;
  const apolloClient = { query: apolloQueryMock } as unknown as ApolloClient;

  const pendingWithdrawalsMock = jest.fn() as jest.MockedFunction<IBeaconNodeAPIClient["getPendingPartialWithdrawals"]>;
  const beaconNodeApiClient = {
    getPendingPartialWithdrawals: pendingWithdrawalsMock,
  } as unknown as jest.Mocked<IBeaconNodeAPIClient>;

  const client = new ConsensysStakingApiClient(logger, retryService, apolloClient, beaconNodeApiClient);

  return {
    client,
    logger,
    retryMock,
    apolloQueryMock,
    pendingWithdrawalsMock,
  };
};

describe("ConsensysStakingApiClient", () => {
  describe("getActiveValidators", () => {
    it("logs and returns undefined when the query returns an error", async () => {
      const { client, logger, retryMock } = createClient();
      const queryError = new Error("graphql failure");

      retryMock.mockImplementationOnce(async (_fn, _timeout) => ({
        data: undefined,
        error: queryError,
      }));

      const result = await client.getActiveValidators();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("getActiveValidators error:", { error: queryError });
    });

    it("logs and returns undefined when the query response lacks data", async () => {
      const { client, logger, retryMock } = createClient();

      retryMock.mockImplementationOnce(async (_fn, _timeout) => ({
        data: undefined,
        error: undefined,
      }));

      const result = await client.getActiveValidators();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("getActiveValidators data undefined");
    });

    it("returns the validator list and logs success when the query succeeds", async () => {
      const { client, logger, retryMock, apolloQueryMock } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n, effectiveBalance: 32n, publicKey: "validator-1", validatorIndex: 1n },
      ];

      apolloQueryMock.mockResolvedValue({
        data: { allValidators: { nodes: validators } },
        error: undefined,
      });

      const result = await client.getActiveValidators();

      expect(result).toEqual(validators);
      expect(retryMock).toHaveBeenCalledTimes(1);
      expect(apolloQueryMock).toHaveBeenCalledWith({ query: ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY });
      expect(logger.debug).toHaveBeenCalledWith("getActiveValidators succeded", { resp: validators });
    });
  });

  describe("getActiveValidatorsWithPendingWithdrawals", () => {
    it("returns undefined when active validator data is unavailable", async () => {
      const { client, pendingWithdrawalsMock } = createClient();
      const getActiveValidatorsSpy = jest.spyOn(client, "getActiveValidators").mockResolvedValueOnce(undefined);
      pendingWithdrawalsMock.mockResolvedValueOnce([]);

      const result = await client.getActiveValidatorsWithPendingWithdrawals();

      expect(result).toBeUndefined();
      expect(pendingWithdrawalsMock).toHaveBeenCalledTimes(1);
      getActiveValidatorsSpy.mockRestore();
    });

    it("returns undefined when pending withdrawals cannot be fetched", async () => {
      const { client, pendingWithdrawalsMock } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n, effectiveBalance: 32n, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const getActiveValidatorsSpy = jest.spyOn(client, "getActiveValidators").mockResolvedValueOnce(validators);
      pendingWithdrawalsMock.mockResolvedValueOnce(undefined);

      const result = await client.getActiveValidatorsWithPendingWithdrawals();

      expect(result).toBeUndefined();
      expect(getActiveValidatorsSpy).toHaveBeenCalledTimes(1);
      getActiveValidatorsSpy.mockRestore();
    });

    it("aggregates pending withdrawals, computes withdrawable amounts, and sorts descending", async () => {
      const { client, logger, pendingWithdrawalsMock } = createClient();

      const validatorA: ValidatorBalance = {
        balance: 40n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-a",
        validatorIndex: 1n,
      };

      const validatorB: ValidatorBalance = {
        balance: 34n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-b",
        validatorIndex: 2n,
      };

      const getActiveValidatorsSpy = jest
        .spyOn(client, "getActiveValidators")
        .mockResolvedValueOnce([validatorB, validatorA]);

      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 0 },
        { validator_index: 1, amount: 3n * ONE_GWEI, withdrawable_epoch: 1 },
        { validator_index: 2, amount: 1n * ONE_GWEI, withdrawable_epoch: 0 },
      ];
      pendingWithdrawalsMock.mockResolvedValueOnce(pendingWithdrawals);

      const result = await client.getActiveValidatorsWithPendingWithdrawals();

      const expectedValidatorA: ValidatorBalanceWithPendingWithdrawal = {
        ...validatorA,
        pendingWithdrawalAmount: 5n * ONE_GWEI,
        withdrawableAmount: safeSub(safeSub(validatorA.balance, 5n * ONE_GWEI), ONE_GWEI * 32n),
      };
      const expectedValidatorB: ValidatorBalanceWithPendingWithdrawal = {
        ...validatorB,
        pendingWithdrawalAmount: 1n * ONE_GWEI,
        withdrawableAmount: safeSub(safeSub(validatorB.balance, 1n * ONE_GWEI), ONE_GWEI * 32n),
      };

      expect(result).toEqual([expectedValidatorA, expectedValidatorB]);
      expect(logger.debug).toHaveBeenCalledWith("getActiveValidatorsWithPendingWithdrawals return val", {
        joined: [expectedValidatorA, expectedValidatorB],
      });

      getActiveValidatorsSpy.mockRestore();
    });
  });

  describe("getTotalPendingPartialWithdrawalsWei", () => {
    it("returns the total pending withdrawals converted to wei and logs it", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalanceWithPendingWithdrawal[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
          pendingWithdrawalAmount: 3n,
          withdrawableAmount: 0n,
        },
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-2",
          validatorIndex: 2n,
          pendingWithdrawalAmount: 1n,
          withdrawableAmount: 0n,
        },
      ];

      const totalWei = client.getTotalPendingPartialWithdrawalsWei(validators);

      expect(totalWei).toBe(4n * ONE_GWEI);
      expect(logger.debug).toHaveBeenCalledWith("getTotalPendingPartialWithdrawalsWei totalWei=4000000000");
    });
  });
});
