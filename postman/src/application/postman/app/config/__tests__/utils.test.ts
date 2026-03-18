import { describe, it, expect } from "@jest/globals";

import {
  DEFAULT_CALLDATA_ENABLED,
  DEFAULT_DB_CLEANER_ENABLED,
  DEFAULT_DB_CLEANING_INTERVAL,
  DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE,
  DEFAULT_ENABLE_POSTMAN_SPONSORING,
  DEFAULT_ENFORCE_MAX_GAS_FEE,
  DEFAULT_EOA_ENABLED,
  DEFAULT_GAS_ESTIMATION_PERCENTILE,
  DEFAULT_INITIAL_FROM_BLOCK,
  DEFAULT_L2_MESSAGE_TREE_DEPTH,
  DEFAULT_LISTENER_BLOCK_CONFIRMATIONS,
  DEFAULT_LISTENER_INTERVAL,
  DEFAULT_MAX_BLOCKS_TO_FETCH_LOGS,
  DEFAULT_MAX_CLAIM_GAS_LIMIT,
  DEFAULT_MAX_FEE_PER_GAS_CAP,
  DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
  DEFAULT_MAX_NONCE_DIFF,
  DEFAULT_MAX_NUMBER_OF_RETRIES,
  DEFAULT_MAX_BUMPS_PER_CYCLE,
  DEFAULT_MAX_POSTMAN_SPONSOR_GAS_LIMIT,
  DEFAULT_MAX_RETRY_CYCLES,
  DEFAULT_MESSAGE_SUBMISSION_TIMEOUT,
  DEFAULT_PROFIT_MARGIN,
  DEFAULT_RETRY_DELAY_IN_SECONDS,
} from "../../../../../core/constants";
import {
  TEST_ADDRESS_1,
  TEST_ADDRESS_2,
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  TEST_L1_SIGNER_PRIVATE_KEY,
  TEST_L2_SIGNER_PRIVATE_KEY,
  TEST_RPC_URL,
} from "../../../../../utils/testing/constants";
import { postmanOptionsSchema } from "../schema";
import { getConfig, isFunctionInterfaceValid, isValidFiltrexExpression, validateEventsFiltersConfig } from "../utils";

describe("Config utils", () => {
  describe("getConfig", () => {
    it("should return the default config when no optional parameters are passed.", () => {
      const config = getConfig({
        l1Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
          listener: {},
          claiming: {
            signer: { type: "private-key" as const, privateKey: TEST_L1_SIGNER_PRIVATE_KEY },
          },
        },
        l2Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
          listener: {},
          claiming: {
            signer: { type: "private-key" as const, privateKey: TEST_L2_SIGNER_PRIVATE_KEY },
          },
        },
        l1L2AutoClaimEnabled: false,
        l2L1AutoClaimEnabled: false,
        databaseOptions: {
          type: "postgres",
        },
      });
      expect(config).toStrictEqual({
        databaseCleanerConfig: {
          cleaningInterval: DEFAULT_DB_CLEANING_INTERVAL,
          daysBeforeNowToDelete: DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE,
          enabled: DEFAULT_DB_CLEANER_ENABLED,
        },
        databaseOptions: {
          type: "postgres",
        },
        l1Config: {
          claiming: {
            claimViaAddress: undefined,
            feeRecipientAddress: undefined,
            gasEstimationPercentile: DEFAULT_GAS_ESTIMATION_PERCENTILE,
            isMaxGasFeeEnforced: DEFAULT_ENFORCE_MAX_GAS_FEE,
            maxClaimGasLimit: DEFAULT_MAX_CLAIM_GAS_LIMIT,
            maxFeePerGasCap: DEFAULT_MAX_FEE_PER_GAS_CAP,
            maxNonceDiff: DEFAULT_MAX_NONCE_DIFF,
            maxNumberOfRetries: DEFAULT_MAX_NUMBER_OF_RETRIES,
            maxBumpsPerCycle: DEFAULT_MAX_BUMPS_PER_CYCLE,
            maxRetryCycles: DEFAULT_MAX_RETRY_CYCLES,
            messageSubmissionTimeout: DEFAULT_MESSAGE_SUBMISSION_TIMEOUT,
            profitMargin: DEFAULT_PROFIT_MARGIN,
            retryDelayInSeconds: DEFAULT_RETRY_DELAY_IN_SECONDS,
            signer: { type: "private-key" as const, privateKey: TEST_L1_SIGNER_PRIVATE_KEY },
            isPostmanSponsorshipEnabled: DEFAULT_ENABLE_POSTMAN_SPONSORING,
            maxPostmanSponsorGasLimit: DEFAULT_MAX_POSTMAN_SPONSOR_GAS_LIMIT,
          },
          isCalldataEnabled: DEFAULT_CALLDATA_ENABLED,
          isEOAEnabled: DEFAULT_EOA_ENABLED,
          listener: {
            blockConfirmation: DEFAULT_LISTENER_BLOCK_CONFIRMATIONS,
            initialFromBlock: DEFAULT_INITIAL_FROM_BLOCK,
            maxBlocksToFetchLogs: DEFAULT_MAX_BLOCKS_TO_FETCH_LOGS,
            maxFetchMessagesFromDb: DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
            pollingInterval: DEFAULT_LISTENER_INTERVAL,
            receiptPollingInterval: DEFAULT_LISTENER_INTERVAL,
          },
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
          rpcUrl: TEST_RPC_URL,
        },
        l1L2AutoClaimEnabled: false,
        l2Config: {
          claiming: {
            claimViaAddress: undefined,
            feeRecipientAddress: undefined,
            gasEstimationPercentile: DEFAULT_GAS_ESTIMATION_PERCENTILE,
            isMaxGasFeeEnforced: DEFAULT_ENFORCE_MAX_GAS_FEE,
            maxClaimGasLimit: DEFAULT_MAX_CLAIM_GAS_LIMIT,
            maxFeePerGasCap: DEFAULT_MAX_FEE_PER_GAS_CAP,
            maxNonceDiff: DEFAULT_MAX_NONCE_DIFF,
            maxNumberOfRetries: DEFAULT_MAX_NUMBER_OF_RETRIES,
            maxBumpsPerCycle: DEFAULT_MAX_BUMPS_PER_CYCLE,
            maxRetryCycles: DEFAULT_MAX_RETRY_CYCLES,
            messageSubmissionTimeout: DEFAULT_MESSAGE_SUBMISSION_TIMEOUT,
            profitMargin: DEFAULT_PROFIT_MARGIN,
            retryDelayInSeconds: DEFAULT_RETRY_DELAY_IN_SECONDS,
            signer: { type: "private-key" as const, privateKey: TEST_L2_SIGNER_PRIVATE_KEY },
            isPostmanSponsorshipEnabled: DEFAULT_ENABLE_POSTMAN_SPONSORING,
            maxPostmanSponsorGasLimit: DEFAULT_MAX_POSTMAN_SPONSOR_GAS_LIMIT,
          },
          enableLineaEstimateGas: false,
          isCalldataEnabled: DEFAULT_CALLDATA_ENABLED,
          isEOAEnabled: DEFAULT_EOA_ENABLED,
          l2MessageTreeDepth: DEFAULT_L2_MESSAGE_TREE_DEPTH,
          listener: {
            blockConfirmation: DEFAULT_LISTENER_BLOCK_CONFIRMATIONS,
            initialFromBlock: DEFAULT_INITIAL_FROM_BLOCK,
            maxBlocksToFetchLogs: DEFAULT_MAX_BLOCKS_TO_FETCH_LOGS,
            maxFetchMessagesFromDb: DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
            pollingInterval: DEFAULT_LISTENER_INTERVAL,
            receiptPollingInterval: DEFAULT_LISTENER_INTERVAL,
          },
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
          rpcUrl: TEST_RPC_URL,
        },
        l2L1AutoClaimEnabled: false,
        loggerOptions: undefined,
        apiConfig: {
          port: 3000,
        },
      });
    });

    it("should return the config when some optional parameters are passed.", () => {
      const config = getConfig({
        l1Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
          listener: {
            pollingInterval: DEFAULT_LISTENER_INTERVAL + 1000,
          },
          claiming: {
            signer: { type: "private-key" as const, privateKey: TEST_L1_SIGNER_PRIVATE_KEY },
            claimViaAddress: TEST_CONTRACT_ADDRESS_1,
            feeRecipientAddress: TEST_ADDRESS_1,
          },
        },
        l2Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
          enableLineaEstimateGas: true,
          listener: {
            pollingInterval: DEFAULT_LISTENER_INTERVAL + 1000,
          },
          claiming: {
            signer: { type: "private-key" as const, privateKey: TEST_L2_SIGNER_PRIVATE_KEY },
            claimViaAddress: TEST_CONTRACT_ADDRESS_2,
            feeRecipientAddress: TEST_ADDRESS_2,
          },
        },
        l1L2AutoClaimEnabled: true,
        l2L1AutoClaimEnabled: true,
        databaseOptions: {
          type: "postgres",
        },
        databaseCleanerOptions: {
          enabled: true,
        },
        apiOptions: {
          port: 9090,
        },
      });
      expect(config).toStrictEqual({
        databaseCleanerConfig: {
          cleaningInterval: DEFAULT_DB_CLEANING_INTERVAL,
          daysBeforeNowToDelete: DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE,
          enabled: true,
        },
        databaseOptions: {
          type: "postgres",
        },
        l1Config: {
          claiming: {
            claimViaAddress: TEST_CONTRACT_ADDRESS_1,
            feeRecipientAddress: TEST_ADDRESS_1,
            gasEstimationPercentile: DEFAULT_GAS_ESTIMATION_PERCENTILE,
            isMaxGasFeeEnforced: DEFAULT_ENFORCE_MAX_GAS_FEE,
            maxClaimGasLimit: DEFAULT_MAX_CLAIM_GAS_LIMIT,
            maxFeePerGasCap: DEFAULT_MAX_FEE_PER_GAS_CAP,
            maxNonceDiff: DEFAULT_MAX_NONCE_DIFF,
            maxNumberOfRetries: DEFAULT_MAX_NUMBER_OF_RETRIES,
            maxBumpsPerCycle: DEFAULT_MAX_BUMPS_PER_CYCLE,
            maxRetryCycles: DEFAULT_MAX_RETRY_CYCLES,
            messageSubmissionTimeout: DEFAULT_MESSAGE_SUBMISSION_TIMEOUT,
            profitMargin: DEFAULT_PROFIT_MARGIN,
            retryDelayInSeconds: DEFAULT_RETRY_DELAY_IN_SECONDS,
            signer: { type: "private-key" as const, privateKey: TEST_L1_SIGNER_PRIVATE_KEY },
            isPostmanSponsorshipEnabled: DEFAULT_ENABLE_POSTMAN_SPONSORING,
            maxPostmanSponsorGasLimit: DEFAULT_MAX_POSTMAN_SPONSOR_GAS_LIMIT,
          },
          isCalldataEnabled: DEFAULT_CALLDATA_ENABLED,
          isEOAEnabled: DEFAULT_EOA_ENABLED,
          listener: {
            blockConfirmation: DEFAULT_LISTENER_BLOCK_CONFIRMATIONS,
            initialFromBlock: DEFAULT_INITIAL_FROM_BLOCK,
            maxBlocksToFetchLogs: DEFAULT_MAX_BLOCKS_TO_FETCH_LOGS,
            maxFetchMessagesFromDb: DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
            pollingInterval: DEFAULT_LISTENER_INTERVAL + 1000,
            receiptPollingInterval: DEFAULT_LISTENER_INTERVAL,
          },
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
          rpcUrl: TEST_RPC_URL,
        },
        l1L2AutoClaimEnabled: true,
        l2Config: {
          claiming: {
            claimViaAddress: TEST_CONTRACT_ADDRESS_2,
            feeRecipientAddress: TEST_ADDRESS_2,
            gasEstimationPercentile: DEFAULT_GAS_ESTIMATION_PERCENTILE,
            isMaxGasFeeEnforced: DEFAULT_ENFORCE_MAX_GAS_FEE,
            maxClaimGasLimit: DEFAULT_MAX_CLAIM_GAS_LIMIT,
            maxFeePerGasCap: DEFAULT_MAX_FEE_PER_GAS_CAP,
            maxNonceDiff: DEFAULT_MAX_NONCE_DIFF,
            maxNumberOfRetries: DEFAULT_MAX_NUMBER_OF_RETRIES,
            maxBumpsPerCycle: DEFAULT_MAX_BUMPS_PER_CYCLE,
            maxRetryCycles: DEFAULT_MAX_RETRY_CYCLES,
            messageSubmissionTimeout: DEFAULT_MESSAGE_SUBMISSION_TIMEOUT,
            profitMargin: DEFAULT_PROFIT_MARGIN,
            retryDelayInSeconds: DEFAULT_RETRY_DELAY_IN_SECONDS,
            signer: { type: "private-key" as const, privateKey: TEST_L2_SIGNER_PRIVATE_KEY },
            isPostmanSponsorshipEnabled: DEFAULT_ENABLE_POSTMAN_SPONSORING,
            maxPostmanSponsorGasLimit: DEFAULT_MAX_POSTMAN_SPONSOR_GAS_LIMIT,
          },
          enableLineaEstimateGas: true,
          isCalldataEnabled: DEFAULT_CALLDATA_ENABLED,
          isEOAEnabled: DEFAULT_EOA_ENABLED,
          l2MessageTreeDepth: DEFAULT_L2_MESSAGE_TREE_DEPTH,
          listener: {
            blockConfirmation: DEFAULT_LISTENER_BLOCK_CONFIRMATIONS,
            initialFromBlock: DEFAULT_INITIAL_FROM_BLOCK,
            maxBlocksToFetchLogs: DEFAULT_MAX_BLOCKS_TO_FETCH_LOGS,
            maxFetchMessagesFromDb: DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
            pollingInterval: DEFAULT_LISTENER_INTERVAL + 1000,
            receiptPollingInterval: DEFAULT_LISTENER_INTERVAL,
          },
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
          rpcUrl: TEST_RPC_URL,
        },
        l2L1AutoClaimEnabled: true,
        loggerOptions: undefined,
        apiConfig: {
          port: 9090,
        },
      });
    });
  });

  describe("validateEventsFiltersConfig", () => {
    it("should throw an error when the from address event filter is not valid", () => {
      expect(() =>
        validateEventsFiltersConfig({
          fromAddressFilter: "0x123",
        }),
      ).toThrow("Invalid fromAddressFilter: 0x123");
    });

    it("should throw an error when the to address event filter is not valid", () => {
      expect(() =>
        validateEventsFiltersConfig({
          toAddressFilter: "0x123",
        }),
      ).toThrow("Invalid toAddressFilter: 0x123");
    });

    it("should not throw an error when filters are valid", () => {
      expect(() =>
        validateEventsFiltersConfig({
          fromAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
          toAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
          calldataFilter: {
            criteriaExpression: `calldata.funcSignature == "0x26dfbc20" and calldata.amount > 0`,
            calldataFunctionInterface: "function receiveFromOtherLayer(address recipient, uint256 amount)",
          },
        }),
      ).not.toThrow();
    });

    it("should throw an error when calldataFilter filter expression is invalid", () => {
      expect(() =>
        validateEventsFiltersConfig({
          fromAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
          toAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
          calldataFilter: {
            criteriaExpression: `calldata.funcSignature == "0x26dfbc20" and calldata.amount = 0`,
            calldataFunctionInterface: "function receiveFromOtherLayer(address recipient, uint256 amount)",
          },
        }),
      ).toThrow('Invalid calldataFilter expression: calldata.funcSignature == "0x26dfbc20" and calldata.amount = 0');
    });

    it("should throw an error when calldataFunctionInterface is invalid", () => {
      expect(() =>
        validateEventsFiltersConfig({
          fromAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
          toAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
          calldataFilter: {
            criteriaExpression: `calldata.funcSignature == "0x26dfbc20" and calldata.amount > 0`,
            calldataFunctionInterface: "function receiveFromOtherLayer(address recipient uint256 amount)",
          },
        }),
      ).toThrow("Invalid calldataFunctionInterface: function receiveFromOtherLayer(address recipient uint256 amount)");
    });

    it("should not throw when no event filters are provided", () => {
      expect(() => validateEventsFiltersConfig(undefined)).not.toThrow();
    });
  });

  describe("getConfig — ZodError handling", () => {
    it("should throw a descriptive error when zod validation fails", () => {
      expect(() =>
        getConfig({
          l1Options: {
            rpcUrl: "",
            messageServiceContractAddress: "not-an-address" as `0x${string}`,
            listener: {},
            claiming: {
              signer: { type: "private-key" as const, privateKey: "not-a-key" as `0x${string}` },
            },
          },
          l2Options: {
            rpcUrl: "",
            messageServiceContractAddress: "not-an-address" as `0x${string}`,
            listener: {},
            claiming: {
              signer: { type: "private-key" as const, privateKey: "not-a-key" as `0x${string}` },
            },
          },
          l1L2AutoClaimEnabled: false,
          l2L1AutoClaimEnabled: false,
          databaseOptions: { type: "postgres" as const },
        }),
      ).toThrow("Invalid postman configuration:");
    });

    it("should include field paths in the validation error message", () => {
      try {
        getConfig({
          l1Options: {
            rpcUrl: "",
            messageServiceContractAddress: "bad" as `0x${string}`,
            listener: {},
            claiming: {
              signer: { type: "private-key" as const, privateKey: "bad" as `0x${string}` },
            },
          },
          l2Options: {
            rpcUrl: "",
            messageServiceContractAddress: "bad" as `0x${string}`,
            listener: {},
            claiming: {
              signer: { type: "private-key" as const, privateKey: "bad" as `0x${string}` },
            },
          },
          l1L2AutoClaimEnabled: false,
          l2L1AutoClaimEnabled: false,
          databaseOptions: { type: "postgres" as const },
        });
        fail("Expected error to be thrown");
      } catch (e) {
        expect(e).toBeInstanceOf(Error);
        expect((e as Error).message).toContain("Invalid postman configuration:");
        expect((e as Error).message).toContain("  - ");
      }
    });
  });

  describe("getConfig — non-ZodError rethrow", () => {
    it("should rethrow non-ZodError errors from schema.parse", () => {
      const parseSpy = jest.spyOn(postmanOptionsSchema, "parse").mockImplementation(() => {
        throw new TypeError("unexpected failure");
      });

      expect(() =>
        getConfig({
          l1Options: {
            rpcUrl: TEST_RPC_URL,
            messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
            listener: {},
            claiming: { signer: { type: "private-key" as const, privateKey: TEST_L1_SIGNER_PRIVATE_KEY } },
          },
          l2Options: {
            rpcUrl: TEST_RPC_URL,
            messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
            listener: {},
            claiming: { signer: { type: "private-key" as const, privateKey: TEST_L2_SIGNER_PRIVATE_KEY } },
          },
          l1L2AutoClaimEnabled: false,
          l2L1AutoClaimEnabled: false,
          databaseOptions: { type: "postgres" as const },
        }),
      ).toThrow(TypeError);

      parseSpy.mockRestore();
    });
  });

  describe("getConfig — L2 event filters", () => {
    it("should include event filters for L2 when provided", () => {
      const config = getConfig({
        l1Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
          listener: {},
          claiming: { signer: { type: "private-key" as const, privateKey: TEST_L1_SIGNER_PRIVATE_KEY } },
        },
        l2Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
          listener: {
            eventFilters: {
              fromAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
            },
          },
          claiming: { signer: { type: "private-key" as const, privateKey: TEST_L2_SIGNER_PRIVATE_KEY } },
        },
        l1L2AutoClaimEnabled: false,
        l2L1AutoClaimEnabled: false,
        databaseOptions: { type: "postgres" as const },
      });

      expect(config.l2Config.listener.eventFilters).toEqual({
        fromAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
      });
    });

    it("should include event filters for L1 when provided", () => {
      const config = getConfig({
        l1Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
          listener: {
            eventFilters: {
              toAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
            },
          },
          claiming: { signer: { type: "private-key" as const, privateKey: TEST_L1_SIGNER_PRIVATE_KEY } },
        },
        l2Options: {
          rpcUrl: TEST_RPC_URL,
          messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
          listener: {},
          claiming: { signer: { type: "private-key" as const, privateKey: TEST_L2_SIGNER_PRIVATE_KEY } },
        },
        l1L2AutoClaimEnabled: false,
        l2L1AutoClaimEnabled: false,
        databaseOptions: { type: "postgres" as const },
      });

      expect(config.l1Config.listener.eventFilters).toEqual({
        toAddressFilter: "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
      });
    });
  });

  describe("isFunctionInterfaceValid", () => {
    it("should return true for a valid function interface", () => {
      expect(isFunctionInterfaceValid("function transfer(address to, uint256 amount)")).toBe(true);
    });

    it("should return false for an invalid function interface", () => {
      expect(isFunctionInterfaceValid("not a valid function")).toBe(false);
    });
  });

  describe("isValidFiltrexExpression", () => {
    it("should return true for a valid expression", () => {
      expect(isValidFiltrexExpression("calldata.amount > 0")).toBe(true);
    });

    it("should return false for an invalid expression", () => {
      expect(isValidFiltrexExpression("amount = = 0")).toBe(false);
    });
  });
});
