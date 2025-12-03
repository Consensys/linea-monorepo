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
      // GraphQL returns string values, which need to be converted to bigint
      const graphqlResponse = [
        { balance: "32", effectiveBalance: "32", publicKey: "validator-1", validatorIndex: "1" },
      ];
      const expectedValidators: ValidatorBalance[] = [
        { balance: 32n, effectiveBalance: 32n, publicKey: "validator-1", validatorIndex: 1n },
      ];

      apolloQueryMock.mockResolvedValue({
        data: { allValidators: { nodes: graphqlResponse } },
        error: undefined,
      });

      const result = await client.getActiveValidators();

      expect(result).toEqual(expectedValidators);
      expect(retryMock).toHaveBeenCalledTimes(1);
      expect(apolloQueryMock).toHaveBeenCalledWith({ query: ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY });
      expect(logger.info).toHaveBeenCalledWith("getActiveValidators succeeded, validatorCount=1");
      expect(logger.debug).toHaveBeenCalledWith("getActiveValidators resp", { resp: expectedValidators });
    });

    it("handles undefined nodes and logs validatorCount=0", async () => {
      const { client, logger, retryMock, apolloQueryMock } = createClient();

      apolloQueryMock.mockResolvedValue({
        data: { allValidators: { nodes: undefined } },
        error: undefined,
      });

      const result = await client.getActiveValidators();

      expect(result).toBeUndefined();
      expect(retryMock).toHaveBeenCalledTimes(1);
      expect(apolloQueryMock).toHaveBeenCalledWith({ query: ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY });
      expect(logger.info).toHaveBeenCalledWith("getActiveValidators succeeded, validatorCount=0");
      expect(logger.debug).toHaveBeenCalledWith("getActiveValidators resp", { resp: undefined });
    });
  });

  describe("getActiveValidatorsWithPendingWithdrawalsAscending", () => {
    it("returns undefined when active validator data is unavailable", async () => {
      const { client, logger, pendingWithdrawalsMock } = createClient();
      const getActiveValidatorsSpy = jest.spyOn(client, "getActiveValidators").mockResolvedValueOnce(undefined);
      pendingWithdrawalsMock.mockResolvedValueOnce([]);

      const result = await client.getActiveValidatorsWithPendingWithdrawalsAscending();

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith(
        "getActiveValidatorsWithPendingWithdrawalsAscending - failed to retrieve validators or pending withdrawals",
        { allValidators: true, pendingWithdrawalsQueue: false },
      );
      expect(pendingWithdrawalsMock).toHaveBeenCalledTimes(1);
      getActiveValidatorsSpy.mockRestore();
    });

    it("returns undefined when pending withdrawals cannot be fetched", async () => {
      const { client, logger, pendingWithdrawalsMock } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n, effectiveBalance: 32n, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const getActiveValidatorsSpy = jest.spyOn(client, "getActiveValidators").mockResolvedValueOnce(validators);
      pendingWithdrawalsMock.mockResolvedValueOnce(undefined);

      const result = await client.getActiveValidatorsWithPendingWithdrawalsAscending();

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith(
        "getActiveValidatorsWithPendingWithdrawalsAscending - failed to retrieve validators or pending withdrawals",
        { allValidators: false, pendingWithdrawalsQueue: true },
      );
      expect(getActiveValidatorsSpy).toHaveBeenCalledTimes(1);
      getActiveValidatorsSpy.mockRestore();
    });

    it("logs warning when joinValidatorsWithPendingWithdrawals returns undefined", async () => {
      const { client, logger, pendingWithdrawalsMock } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const getActiveValidatorsSpy = jest.spyOn(client, "getActiveValidators").mockResolvedValueOnce(validators);
      const joinValidatorsSpy = jest
        .spyOn(client, "joinValidatorsWithPendingWithdrawals")
        .mockReturnValueOnce(undefined);
      pendingWithdrawalsMock.mockResolvedValueOnce([
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 0 },
      ]);

      const result = await client.getActiveValidatorsWithPendingWithdrawalsAscending();

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith(
        "getActiveValidatorsWithPendingWithdrawalsAscending - joinValidatorsWithPendingWithdrawals returned undefined",
      );
      expect(joinValidatorsSpy).toHaveBeenCalledWith(validators, expect.any(Array));
      getActiveValidatorsSpy.mockRestore();
      joinValidatorsSpy.mockRestore();
    });

    it("aggregates pending withdrawals, computes withdrawable amounts, and sorts ascending", async () => {
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

      const result = await client.getActiveValidatorsWithPendingWithdrawalsAscending();

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

      // With ascending sort: B (1 GWEI) should come before A (3 GWEI)
      expect(result).toEqual([expectedValidatorB, expectedValidatorA]);
      expect(logger.info).toHaveBeenCalledWith("getActiveValidatorsWithPendingWithdrawalsAscending succeeded, validatorCount=2");
      expect(logger.debug).toHaveBeenCalledWith("getActiveValidatorsWithPendingWithdrawalsAscending joined", {
        joined: [expectedValidatorB, expectedValidatorA],
      });

      getActiveValidatorsSpy.mockRestore();
    });

    it("sorts validators ascending by withdrawable amount", async () => {
      const { client, pendingWithdrawalsMock } = createClient();

      const validatorHigh: ValidatorBalance = {
        balance: 45n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-high",
        validatorIndex: 10n,
      };

      const validatorLow: ValidatorBalance = {
        balance: 40n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-low",
        validatorIndex: 11n,
      };

      const getActiveValidatorsSpy = jest
        .spyOn(client, "getActiveValidators")
        .mockResolvedValueOnce([validatorHigh, validatorLow]);

      pendingWithdrawalsMock.mockResolvedValueOnce([
        {
          validator_index: Number(validatorHigh.validatorIndex),
          amount: 2n * ONE_GWEI,
          withdrawable_epoch: 0,
        },
        {
          validator_index: Number(validatorLow.validatorIndex),
          amount: 6n * ONE_GWEI,
          withdrawable_epoch: 0,
        },
      ]);

      const result = await client.getActiveValidatorsWithPendingWithdrawalsAscending();

      const expectedHigh: ValidatorBalanceWithPendingWithdrawal = {
        ...validatorHigh,
        pendingWithdrawalAmount: 2n * ONE_GWEI,
        withdrawableAmount: safeSub(safeSub(validatorHigh.balance, 2n * ONE_GWEI), ONE_GWEI * 32n),
      };
      const expectedLow: ValidatorBalanceWithPendingWithdrawal = {
        ...validatorLow,
        pendingWithdrawalAmount: 6n * ONE_GWEI,
        withdrawableAmount: safeSub(safeSub(validatorLow.balance, 6n * ONE_GWEI), ONE_GWEI * 32n),
      };

      // With ascending sort: Low (2 GWEI) should come before High (11 GWEI)
      expect(result).toEqual([expectedLow, expectedHigh]);

      getActiveValidatorsSpy.mockRestore();
    });

    it("handles validators with equal withdrawable amounts", async () => {
      const { client, pendingWithdrawalsMock } = createClient();

      const validatorEqualA: ValidatorBalance = {
        balance: 36n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-equal-a",
        validatorIndex: 20n,
      };
      const validatorEqualB: ValidatorBalance = {
        balance: 32n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-equal-b",
        validatorIndex: 21n,
      };

      const getActiveValidatorsSpy = jest
        .spyOn(client, "getActiveValidators")
        .mockResolvedValueOnce([validatorEqualA, validatorEqualB]);

      pendingWithdrawalsMock.mockResolvedValueOnce([
        {
          validator_index: Number(validatorEqualA.validatorIndex),
          amount: 4n * ONE_GWEI,
          withdrawable_epoch: 0,
        },
      ]);

      const result = await client.getActiveValidatorsWithPendingWithdrawalsAscending();

      const expectedEqualA: ValidatorBalanceWithPendingWithdrawal = {
        ...validatorEqualA,
        pendingWithdrawalAmount: 4n * ONE_GWEI,
        withdrawableAmount: safeSub(safeSub(validatorEqualA.balance, 4n * ONE_GWEI), ONE_GWEI * 32n),
      };

      const expectedEqualB: ValidatorBalanceWithPendingWithdrawal = {
        ...validatorEqualB,
        pendingWithdrawalAmount: 0n,
        withdrawableAmount: safeSub(safeSub(validatorEqualB.balance, 0n), ONE_GWEI * 32n),
      };

      expect(result).toEqual(expect.arrayContaining([expectedEqualA, expectedEqualB]));

      getActiveValidatorsSpy.mockRestore();
    });
  });

  describe("joinValidatorsWithPendingWithdrawals", () => {
    it("returns undefined and logs warning when validators are undefined", () => {
      const { client, logger } = createClient();
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 0 },
      ];

      const result = client.joinValidatorsWithPendingWithdrawals(undefined, pendingWithdrawals);

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - invalid inputs", {
        allValidators: true,
        pendingWithdrawalsQueue: false,
      });
    });

    it("returns undefined and logs warning when pending withdrawals are undefined", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];

      const result = client.joinValidatorsWithPendingWithdrawals(validators, undefined);

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - invalid inputs", {
        allValidators: false,
        pendingWithdrawalsQueue: true,
      });
    });

    it("aggregates pending withdrawals and computes withdrawable amounts", () => {
      const { client, logger } = createClient();

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

      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 0 },
        { validator_index: 1, amount: 3n * ONE_GWEI, withdrawable_epoch: 1 },
        { validator_index: 2, amount: 1n * ONE_GWEI, withdrawable_epoch: 0 },
      ];

      const result = client.joinValidatorsWithPendingWithdrawals([validatorB, validatorA], pendingWithdrawals);

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

      // Result maintains input order (B, then A) - no sorting
      expect(result).toEqual([expectedValidatorB, expectedValidatorA]);

      // Verify debug logs
      expect(logger.debug).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - aggregated pending withdrawals", {
        uniqueValidatorIndices: 2,
        pendingByValidator: expect.arrayContaining([
          { validator_index: 1, totalAmount: (5n * ONE_GWEI).toString() },
          { validator_index: 2, totalAmount: (1n * ONE_GWEI).toString() },
        ]),
      });
      expect(logger.debug).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - joined results", {
        joinedCount: 2,
        validatorsWithPendingWithdrawals: 2,
        allEntries: expect.arrayContaining([
          expect.objectContaining({
            validatorIndex: validatorB.validatorIndex.toString(),
            publicKey: validatorB.publicKey,
            pendingWithdrawalAmount: expectedValidatorB.pendingWithdrawalAmount.toString(),
            withdrawableAmount: expectedValidatorB.withdrawableAmount.toString(),
          }),
          expect.objectContaining({
            validatorIndex: validatorA.validatorIndex.toString(),
            publicKey: validatorA.publicKey,
            pendingWithdrawalAmount: expectedValidatorA.pendingWithdrawalAmount.toString(),
            withdrawableAmount: expectedValidatorA.withdrawableAmount.toString(),
          }),
        ]),
      });
    });

    it("handles validators with no pending withdrawals", () => {
      const { client, logger } = createClient();

      const validator: ValidatorBalance = {
        balance: 32n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-1",
        validatorIndex: 1n,
      };

      const pendingWithdrawals: PendingPartialWithdrawal[] = [];

      const result = client.joinValidatorsWithPendingWithdrawals([validator], pendingWithdrawals);

      const expected: ValidatorBalanceWithPendingWithdrawal = {
        ...validator,
        pendingWithdrawalAmount: 0n,
        withdrawableAmount: safeSub(safeSub(validator.balance, 0n), ONE_GWEI * 32n),
      };

      expect(result).toEqual([expected]);

      // Verify debug logs
      expect(logger.debug).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - aggregated pending withdrawals", {
        uniqueValidatorIndices: 0,
        pendingByValidator: [],
      });
      expect(logger.debug).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - joined results", {
        joinedCount: 1,
        validatorsWithPendingWithdrawals: 0,
        allEntries: [
          expect.objectContaining({
            validatorIndex: validator.validatorIndex.toString(),
            publicKey: validator.publicKey,
            pendingWithdrawalAmount: "0",
            withdrawableAmount: expected.withdrawableAmount.toString(),
          }),
        ],
      });
    });

    it("handles empty validators array", () => {
      const { client, logger } = createClient();
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 0 },
      ];

      const result = client.joinValidatorsWithPendingWithdrawals([], pendingWithdrawals);

      expect(result).toEqual([]);

      // Verify debug logs
      expect(logger.debug).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - aggregated pending withdrawals", {
        uniqueValidatorIndices: 1,
        pendingByValidator: [{ validator_index: 1, totalAmount: (2n * ONE_GWEI).toString() }],
      });
      expect(logger.debug).toHaveBeenCalledWith("joinValidatorsWithPendingWithdrawals - joined results", {
        joinedCount: 0,
        validatorsWithPendingWithdrawals: 0,
        allEntries: [],
      });
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
      expect(logger.info).toHaveBeenCalledWith("getTotalPendingPartialWithdrawalsWei totalWei=4000000000");
    });
  });

  describe("getFilteredAndAggregatedPendingWithdrawals", () => {
    it("returns undefined and logs warning when validators are undefined", () => {
      const { client, logger } = createClient();
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(undefined, pendingWithdrawals);

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals - invalid inputs", {
        allValidators: true,
        pendingWithdrawalsQueue: false,
      });
    });

    it("returns undefined and logs warning when pending withdrawals are undefined", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, undefined);

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals - invalid inputs", {
        allValidators: false,
        pendingWithdrawalsQueue: true,
      });
    });

    it("filters out withdrawals that do not match active validators", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
        { validator_index: 999, amount: 5n * ONE_GWEI, withdrawable_epoch: 200 }, // Not in validators
        { validator_index: 2, amount: 3n * ONE_GWEI, withdrawable_epoch: 150 }, // Not in validators
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

      expect(result).toEqual([
        {
          validator_index: 1,
          withdrawable_epoch: 100,
          amount: 2n * ONE_GWEI,
          pubkey: "validator-1",
        },
      ]);
      expect(logger.debug).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals - filtered withdrawals", {
        totalPendingWithdrawals: 3,
        filteredCount: 1,
        uniqueValidatorIndices: 1,
      });
    });

    it("aggregates amounts by validator_index and withdrawable_epoch", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-2", validatorIndex: 2n },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
        { validator_index: 1, amount: 3n * ONE_GWEI, withdrawable_epoch: 100 }, // Same validator and epoch - should aggregate
        { validator_index: 1, amount: 5n * ONE_GWEI, withdrawable_epoch: 200 }, // Same validator, different epoch
        { validator_index: 2, amount: 1n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

      expect(result).toEqual(
        expect.arrayContaining([
          {
            validator_index: 1,
            withdrawable_epoch: 100,
            amount: 5n * ONE_GWEI, // 2 + 3 aggregated
            pubkey: "validator-1",
          },
          {
            validator_index: 1,
            withdrawable_epoch: 200,
            amount: 5n * ONE_GWEI,
            pubkey: "validator-1",
          },
          {
            validator_index: 2,
            withdrawable_epoch: 100,
            amount: 1n * ONE_GWEI,
            pubkey: "validator-2",
          },
        ]),
      );
      expect(result).toHaveLength(3);
      expect(logger.info).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals succeeded, aggregatedCount=3");
    });

    it("includes pubkey from matching validators", () => {
      const { client } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "pubkey-abc", validatorIndex: 1n },
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "pubkey-xyz", validatorIndex: 2n },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
        { validator_index: 2, amount: 3n * ONE_GWEI, withdrawable_epoch: 200 },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

      expect(result).toEqual([
        {
          validator_index: 1,
          withdrawable_epoch: 100,
          amount: 2n * ONE_GWEI,
          pubkey: "pubkey-abc",
        },
        {
          validator_index: 2,
          withdrawable_epoch: 200,
          amount: 3n * ONE_GWEI,
          pubkey: "pubkey-xyz",
        },
      ]);
    });

    it("handles empty validators array", () => {
      const { client, logger } = createClient();
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals([], pendingWithdrawals);

      expect(result).toEqual([]);
      expect(logger.debug).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals - filtered withdrawals", {
        totalPendingWithdrawals: 1,
        filteredCount: 0,
        uniqueValidatorIndices: 0,
      });
      expect(logger.info).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals succeeded, aggregatedCount=0");
    });

    it("handles empty pending withdrawals array", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, []);

      expect(result).toEqual([]);
      expect(logger.debug).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals - filtered withdrawals", {
        totalPendingWithdrawals: 0,
        filteredCount: 0,
        uniqueValidatorIndices: 1,
      });
      expect(logger.info).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals succeeded, aggregatedCount=0");
    });

    it("handles multiple withdrawals with same validator_index and withdrawable_epoch", () => {
      const { client } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 1n * ONE_GWEI, withdrawable_epoch: 100 },
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
        { validator_index: 1, amount: 3n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

      expect(result).toEqual([
        {
          validator_index: 1,
          withdrawable_epoch: 100,
          amount: 6n * ONE_GWEI, // 1 + 2 + 3 aggregated
          pubkey: "validator-1",
        },
      ]);
    });

    it("handles withdrawals with different withdrawable_epochs for same validator", () => {
      const { client } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
        { validator_index: 1, amount: 3n * ONE_GWEI, withdrawable_epoch: 200 },
        { validator_index: 1, amount: 1n * ONE_GWEI, withdrawable_epoch: 100 }, // Same epoch as first
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

      expect(result).toEqual(
        expect.arrayContaining([
          {
            validator_index: 1,
            withdrawable_epoch: 100,
            amount: 3n * ONE_GWEI, // 2 + 1 aggregated
            pubkey: "validator-1",
          },
          {
            validator_index: 1,
            withdrawable_epoch: 200,
            amount: 3n * ONE_GWEI,
            pubkey: "validator-1",
          },
        ]),
      );
      expect(result).toHaveLength(2);
    });

    it("logs aggregated results with debug information", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

      expect(logger.debug).toHaveBeenCalledWith("getFilteredAndAggregatedPendingWithdrawals - aggregated results", {
        aggregatedCount: 1,
        allEntries: [
          {
            validator_index: 1,
            withdrawable_epoch: 100,
            amount: (2n * ONE_GWEI).toString(),
            pubkey: "validator-1",
          },
        ],
      });
    });

    it("handles validators with bigint validatorIndex correctly", () => {
      const { client } = createClient();
      // Use a smaller number to avoid precision loss warnings
      const validatorIndex = 12345n;
      const validators: ValidatorBalance[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: validatorIndex,
        },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: Number(validatorIndex), amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

      expect(result).toEqual([
        {
          validator_index: Number(validatorIndex),
          withdrawable_epoch: 100,
          amount: 2n * ONE_GWEI,
          pubkey: "validator-1",
        },
      ]);
    });

    it("handles missing pubkey in map by using empty string fallback", () => {
      const { client } = createClient();
      // Test the defensive fallback when pubkeyByValidatorIndex.get() returns undefined
      // This tests the ?? "" fallback on line 203
      // We need to simulate a scenario where the map lookup returns undefined
      const validators: ValidatorBalance[] = [
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n },
      ];
      const pendingWithdrawals: PendingPartialWithdrawal[] = [
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      // Create a spy that tracks calls to Map.get specifically for the pubkey lookup
      // We'll intercept the call when looking up validator_index 1 in the pubkey map
      const originalGet = Map.prototype.get;
      let getCallCount = 0;
      Map.prototype.get = function (this: Map<unknown, unknown>, key: unknown) {
        getCallCount++;
        // The aggregatedMap.get() call happens first (line 194)
        // Then pubkeyByValidatorIndex.get() happens in the map() callback (line 203)
        // We want to return undefined for the pubkey lookup specifically
        // We can detect this by checking if the key is a number (validator_index)
        // and if we're past the first get call (which is for aggregatedMap)
        if (typeof key === "number" && getCallCount > 1) {
          return undefined; // Simulate missing pubkey
        }
        return originalGet.call(this, key);
      };

      try {
        const result = client.getFilteredAndAggregatedPendingWithdrawals(validators, pendingWithdrawals);

        expect(result).toEqual([
          {
            validator_index: 1,
            withdrawable_epoch: 100,
            amount: 2n * ONE_GWEI,
            pubkey: "", // Fallback to empty string when map lookup returns undefined
          },
        ]);
      } finally {
        // Restore original Map.get
        Map.prototype.get = originalGet;
      }
    });
  });
});
