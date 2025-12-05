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
import { EXITING_VALIDATORS_QUERY } from "../../core/entities/graphql/ExitingValidators.js";
import type {
  ExitingValidator,
  ValidatorBalance,
  ValidatorBalanceWithPendingWithdrawal,
} from "../../core/entities/ValidatorBalance.js";

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
        { balance: "32", effectiveBalance: "32", publicKey: "validator-1", validatorIndex: "1", activationEpoch: "0" },
      ];
      const expectedValidators: ValidatorBalance[] = [
        { balance: 32n, effectiveBalance: 32n, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
      ];

      apolloQueryMock.mockResolvedValue({
        data: { allHeadValidators: { nodes: graphqlResponse } },
        error: undefined,
      });

      const result = await client.getActiveValidators();

      expect(result).toEqual(expectedValidators);
      expect(retryMock).toHaveBeenCalledTimes(1);
      expect(apolloQueryMock).toHaveBeenCalledWith({
        query: ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY,
        fetchPolicy: "network-only",
      });
      expect(logger.info).toHaveBeenCalledWith("getActiveValidators succeeded, validatorCount=1");
      expect(logger.debug).toHaveBeenCalledWith("getActiveValidators resp", { resp: expectedValidators });
    });

    it("handles undefined nodes and logs validatorCount=0", async () => {
      const { client, logger, retryMock, apolloQueryMock } = createClient();

      apolloQueryMock.mockResolvedValue({
        data: { allHeadValidators: { nodes: undefined } },
        error: undefined,
      });

      const result = await client.getActiveValidators();

      expect(result).toBeUndefined();
      expect(retryMock).toHaveBeenCalledTimes(1);
      expect(apolloQueryMock).toHaveBeenCalledWith({
        query: ALL_VALIDATORS_BY_LARGEST_BALANCE_QUERY,
        fetchPolicy: "network-only",
      });
      expect(logger.info).toHaveBeenCalledWith("getActiveValidators succeeded, validatorCount=0");
      expect(logger.debug).toHaveBeenCalledWith("getActiveValidators resp", { resp: undefined });
    });
  });

  describe("getExitingValidators", () => {
    it("logs and returns undefined when the query returns an error", async () => {
      const { client, logger, retryMock } = createClient();
      const queryError = new Error("graphql failure");

      retryMock.mockImplementationOnce(async (_fn, _timeout) => ({
        data: undefined,
        error: queryError,
      }));

      const result = await client.getExitingValidators();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("getExitingValidators error:", { error: queryError });
    });

    it("logs and returns undefined when the query response lacks data", async () => {
      const { client, logger, retryMock } = createClient();

      retryMock.mockImplementationOnce(async (_fn, _timeout) => ({
        data: undefined,
        error: undefined,
      }));

      const result = await client.getExitingValidators();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("getExitingValidators data undefined");
    });

    it("returns the validator list and logs success when the query succeeds", async () => {
      const { client, logger, retryMock, apolloQueryMock } = createClient();
      // GraphQL returns string values, which need to be converted to appropriate types
      const exitDateString = "2024-01-15T10:30:00Z";
      const graphqlResponse = [
        {
          balance: "32",
          effectiveBalance: "32",
          publicKey: "validator-1",
          validatorIndex: "1",
          exitEpoch: "100",
          exitDate: exitDateString,
          slashed: false,
        },
      ];
      const expectedValidators: ExitingValidator[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      apolloQueryMock.mockResolvedValue({
        data: { allHeadValidators: { nodes: graphqlResponse } },
        error: undefined,
      });

      const result = await client.getExitingValidators();

      expect(result).toEqual(expectedValidators);
      expect(retryMock).toHaveBeenCalledTimes(1);
      expect(apolloQueryMock).toHaveBeenCalledWith({
        query: EXITING_VALIDATORS_QUERY,
        fetchPolicy: "network-only",
      });
      expect(logger.info).toHaveBeenCalledWith("getExitingValidators succeeded, validatorCount=1");
      expect(logger.debug).toHaveBeenCalledWith("getExitingValidators resp", { resp: expectedValidators });
    });

    it("handles undefined nodes and logs validatorCount=0", async () => {
      const { client, logger, retryMock, apolloQueryMock } = createClient();

      apolloQueryMock.mockResolvedValue({
        data: { allHeadValidators: { nodes: undefined } },
        error: undefined,
      });

      const result = await client.getExitingValidators();

      expect(result).toBeUndefined();
      expect(retryMock).toHaveBeenCalledTimes(1);
      expect(apolloQueryMock).toHaveBeenCalledWith({
        query: EXITING_VALIDATORS_QUERY,
        fetchPolicy: "network-only",
      });
      expect(logger.info).toHaveBeenCalledWith("getExitingValidators succeeded, validatorCount=0");
      expect(logger.debug).toHaveBeenCalledWith("getExitingValidators resp", { resp: undefined });
    });

    it("correctly converts all field types including exitEpoch, exitDate, and slashed", async () => {
      const { client, apolloQueryMock } = createClient();
      const exitDateString1 = "2024-01-15T10:30:00Z";
      const exitDateString2 = "2024-02-20T15:45:30Z";
      const graphqlResponse = [
        {
          balance: "40",
          effectiveBalance: "32",
          publicKey: "validator-slashed",
          validatorIndex: "10",
          exitEpoch: "200",
          exitDate: exitDateString1,
          slashed: true,
        },
        {
          balance: "35",
          effectiveBalance: "32",
          publicKey: "validator-normal",
          validatorIndex: "20",
          exitEpoch: "150",
          exitDate: exitDateString2,
          slashed: false,
        },
      ];

      apolloQueryMock.mockResolvedValue({
        data: { allHeadValidators: { nodes: graphqlResponse } },
        error: undefined,
      });

      const result = await client.getExitingValidators();

      expect(result).toEqual([
        {
          balance: 40n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed",
          validatorIndex: 10n,
          exitEpoch: 200,
          exitDate: new Date(exitDateString1),
          slashed: true,
        },
        {
          balance: 35n,
          effectiveBalance: 32n,
          publicKey: "validator-normal",
          validatorIndex: 20n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString2),
          slashed: false,
        },
      ]);
    });

    it("handles exitEpoch as number (not string) correctly", async () => {
      const { client, apolloQueryMock } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const graphqlResponse = [
        {
          balance: "40",
          effectiveBalance: "32",
          publicKey: "validator-1",
          validatorIndex: "10",
          exitEpoch: 200, // Already a number, not a string
          exitDate: exitDateString,
          slashed: false,
        },
      ];

      apolloQueryMock.mockResolvedValue({
        data: { allHeadValidators: { nodes: graphqlResponse } },
        error: undefined,
      });

      const result = await client.getExitingValidators();

      expect(result).toEqual([
        {
          balance: 40n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 10n,
          exitEpoch: 200, // Should remain as number
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ]);
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
        { balance: 32n, effectiveBalance: 32n, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        activationEpoch: 0,
      };

      const validatorB: ValidatorBalance = {
        balance: 34n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-b",
        validatorIndex: 2n,
        activationEpoch: 0,
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
        activationEpoch: 0,
      };

      const validatorLow: ValidatorBalance = {
        balance: 40n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-low",
        validatorIndex: 11n,
        activationEpoch: 0,
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
        activationEpoch: 0,
      };
      const validatorEqualB: ValidatorBalance = {
        balance: 32n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-equal-b",
        validatorIndex: 21n,
        activationEpoch: 0,
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        activationEpoch: 0,
      };

      const validatorB: ValidatorBalance = {
        balance: 34n * ONE_GWEI,
        effectiveBalance: 32n * ONE_GWEI,
        publicKey: "validator-b",
        validatorIndex: 2n,
        activationEpoch: 0,
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
        activationEpoch: 0,
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
          activationEpoch: 0,
          pendingWithdrawalAmount: 3n,
          withdrawableAmount: 0n,
        },
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-2",
          validatorIndex: 2n,
          activationEpoch: 0,
          pendingWithdrawalAmount: 1n,
          withdrawableAmount: 0n,
        },
      ];

      const totalWei = client.getTotalPendingPartialWithdrawalsWei(validators);

      expect(totalWei).toBe(4n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalPendingPartialWithdrawalsWei totalWei=4000000000");
    });
  });

  describe("getTotalValidatorBalanceGwei", () => {
    it("returns undefined when input is undefined", () => {
      const { client, logger } = createClient();

      const result = client.getTotalValidatorBalanceGwei(undefined);

      expect(result).toBeUndefined();
      expect(logger.info).not.toHaveBeenCalled();
    });

    it("returns undefined when input is empty array", () => {
      const { client, logger } = createClient();

      const result = client.getTotalValidatorBalanceGwei([]);

      expect(result).toBeUndefined();
      expect(logger.info).not.toHaveBeenCalled();
    });

    it("correctly sums balances from multiple validators and logs it", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          activationEpoch: 0,
        },
        {
          balance: 40n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-2",
          validatorIndex: 2n,
          activationEpoch: 0,
        },
        {
          balance: 35n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-3",
          validatorIndex: 3n,
          activationEpoch: 0,
        },
      ];

      const totalGwei = client.getTotalValidatorBalanceGwei(validators);

      expect(totalGwei).toBe(107n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalValidatorBalanceGwei totalGwei=107000000000");
    });

    it("handles single validator case", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          activationEpoch: 0,
        },
      ];

      const totalGwei = client.getTotalValidatorBalanceGwei(validators);

      expect(totalGwei).toBe(32n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalValidatorBalanceGwei totalGwei=32000000000");
    });

    it("handles large bigint values correctly", () => {
      const { client, logger } = createClient();
      const largeBalance = 1000000n * ONE_GWEI;
      const validators: ValidatorBalance[] = [
        {
          balance: largeBalance,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          activationEpoch: 0,
        },
        {
          balance: largeBalance,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-2",
          validatorIndex: 2n,
          activationEpoch: 0,
        },
      ];

      const totalGwei = client.getTotalValidatorBalanceGwei(validators);

      expect(totalGwei).toBe(2000000n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalValidatorBalanceGwei totalGwei=2000000000000000");
    });

    it("handles validators with zero balance", () => {
      const { client, logger } = createClient();
      const validators: ValidatorBalance[] = [
        {
          balance: 0n,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          activationEpoch: 0,
        },
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-2",
          validatorIndex: 2n,
          activationEpoch: 0,
        },
      ];

      const totalGwei = client.getTotalValidatorBalanceGwei(validators);

      expect(totalGwei).toBe(32n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalValidatorBalanceGwei totalGwei=32000000000");
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-2", validatorIndex: 2n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "pubkey-abc", validatorIndex: 1n, activationEpoch: 0 },
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "pubkey-xyz", validatorIndex: 2n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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
          activationEpoch: 0,
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
        { balance: 32n * ONE_GWEI, effectiveBalance: 32n * ONE_GWEI, publicKey: "validator-1", validatorIndex: 1n, activationEpoch: 0 },
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

  describe("getSlashedValidators", () => {
    it("returns undefined and logs warning when input is undefined", () => {
      const { client, logger } = createClient();

      const result = client.getSlashedValidators(undefined);

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith("getSlashedValidators - invalid input: validators is undefined");
    });

    it("returns empty array when input is empty array", () => {
      const { client, logger } = createClient();

      const result = client.getSlashedValidators([]);

      expect(result).toEqual([]);
      expect(logger.info).toHaveBeenCalledWith("getSlashedValidators succeeded, slashedCount=0");
    });

    it("returns only slashed validators (slashed=true)", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 40n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 35n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed-2",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
      ];

      const result = client.getSlashedValidators(validators);

      expect(result).toEqual(validators);
      expect(logger.info).toHaveBeenCalledWith("getSlashedValidators succeeded, slashedCount=2");
    });

    it("filters out non-slashed validators", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 40n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 35n,
          effectiveBalance: 32n,
          publicKey: "validator-normal",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      const result = client.getSlashedValidators(validators);

      expect(result).toEqual([validators[0]]);
      expect(logger.info).toHaveBeenCalledWith("getSlashedValidators succeeded, slashedCount=1");
    });

    it("handles mixed slashed/non-slashed validators correctly", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 40n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 35n,
          effectiveBalance: 32n,
          publicKey: "validator-normal-1",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 38n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed-2",
          validatorIndex: 3n,
          exitEpoch: 200,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 33n,
          effectiveBalance: 32n,
          publicKey: "validator-normal-2",
          validatorIndex: 4n,
          exitEpoch: 250,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      const result = client.getSlashedValidators(validators);

      expect(result).toEqual([validators[0], validators[2]]);
      expect(logger.info).toHaveBeenCalledWith("getSlashedValidators succeeded, slashedCount=2");
    });
  });

  describe("getNonSlashedAndExitingValidators", () => {
    it("returns undefined and logs warning when input is undefined", () => {
      const { client, logger } = createClient();

      const result = client.getNonSlashedAndExitingValidators(undefined);

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith("getNonSlashedAndExitingValidators - invalid input: validators is undefined");
    });

    it("returns empty array when input is empty array", () => {
      const { client, logger } = createClient();

      const result = client.getNonSlashedAndExitingValidators([]);

      expect(result).toEqual([]);
      expect(logger.info).toHaveBeenCalledWith("getNonSlashedAndExitingValidators succeeded, nonSlashedCount=0");
    });

    it("returns only non-slashed validators (slashed=false)", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 35n,
          effectiveBalance: 32n,
          publicKey: "validator-normal-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 33n,
          effectiveBalance: 32n,
          publicKey: "validator-normal-2",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      const result = client.getNonSlashedAndExitingValidators(validators);

      expect(result).toEqual(validators);
      expect(logger.info).toHaveBeenCalledWith("getNonSlashedAndExitingValidators succeeded, nonSlashedCount=2");
    });

    it("filters out slashed validators", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 40n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 35n,
          effectiveBalance: 32n,
          publicKey: "validator-normal",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      const result = client.getNonSlashedAndExitingValidators(validators);

      expect(result).toEqual([validators[1]]);
      expect(logger.info).toHaveBeenCalledWith("getNonSlashedAndExitingValidators succeeded, nonSlashedCount=1");
    });

    it("handles mixed slashed/non-slashed validators correctly", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 40n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 35n,
          effectiveBalance: 32n,
          publicKey: "validator-normal-1",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 38n,
          effectiveBalance: 32n,
          publicKey: "validator-slashed-2",
          validatorIndex: 3n,
          exitEpoch: 200,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 33n,
          effectiveBalance: 32n,
          publicKey: "validator-normal-2",
          validatorIndex: 4n,
          exitEpoch: 250,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      const result = client.getNonSlashedAndExitingValidators(validators);

      expect(result).toEqual([validators[1], validators[3]]);
      expect(logger.info).toHaveBeenCalledWith("getNonSlashedAndExitingValidators succeeded, nonSlashedCount=2");
    });
  });

  describe("getTotalBalanceOfExitingValidators", () => {
    it("returns undefined and logs warning when input is undefined", () => {
      const { client, logger } = createClient();

      const result = client.getTotalBalanceOfExitingValidators(undefined);

      expect(result).toBeUndefined();
      expect(logger.warn).toHaveBeenCalledWith("getTotalBalanceOfExitingValidators - invalid input: validators is undefined", {
        validators: true,
      });
    });

    it("returns 0n and logs info when input is empty array", () => {
      const { client, logger } = createClient();

      const result = client.getTotalBalanceOfExitingValidators([]);

      expect(result).toBe(0n);
      expect(logger.info).toHaveBeenCalledWith("getTotalBalanceOfExitingValidators - empty array, returning 0");
    });

    it("correctly sums balances from multiple validators", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 40n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-2",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 35n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-3",
          validatorIndex: 3n,
          exitEpoch: 200,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
      ];

      const totalGwei = client.getTotalBalanceOfExitingValidators(validators);

      expect(totalGwei).toBe(107n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalBalanceOfExitingValidators totalGwei=107000000000");
    });

    it("handles single validator case", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      const totalGwei = client.getTotalBalanceOfExitingValidators(validators);

      expect(totalGwei).toBe(32n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalBalanceOfExitingValidators totalGwei=32000000000");
    });

    it("handles large bigint values correctly", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const largeBalance = 1000000n * ONE_GWEI;
      const validators: ExitingValidator[] = [
        {
          balance: largeBalance,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: largeBalance,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-2",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
      ];

      const totalGwei = client.getTotalBalanceOfExitingValidators(validators);

      expect(totalGwei).toBe(2000000n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalBalanceOfExitingValidators totalGwei=2000000000000000");
    });

    it("handles validators with zero balance", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 0n,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-2",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      const totalGwei = client.getTotalBalanceOfExitingValidators(validators);

      expect(totalGwei).toBe(32n * ONE_GWEI);
      expect(logger.info).toHaveBeenCalledWith("getTotalBalanceOfExitingValidators totalGwei=32000000000");
    });

    it("logs total balance on success", () => {
      const { client, logger } = createClient();
      const exitDateString = "2024-01-15T10:30:00Z";
      const validators: ExitingValidator[] = [
        {
          balance: 40n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      client.getTotalBalanceOfExitingValidators(validators);

      expect(logger.info).toHaveBeenCalledWith("getTotalBalanceOfExitingValidators totalGwei=40000000000");
    });
  });
});
