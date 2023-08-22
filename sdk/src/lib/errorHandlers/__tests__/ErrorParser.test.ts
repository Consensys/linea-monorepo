import { describe, it, expect } from "@jest/globals";
import { BigNumber } from "ethers";
import { EthersError } from "@enzoferey/ethers-error-parser";
import { ErrorParser } from "../ErrorParser";

describe("ErrorParser", () => {
  describe("parseErrorWithMitigation", () => {
    it("should return null when error is null", () => {
      expect(ErrorParser.parseErrorWithMitigation(null as unknown as EthersError)).toStrictEqual(null);
    });

    it("should return UNKNOWN_ERROR and shouldRetry = true when error code = CALL_EXCEPTION ", () => {
      const errorMessage = {
        message: "",
        code: "CALL_EXCEPTION",
        reason: "any reason",
      };

      expect(ErrorParser.parseErrorWithMitigation(errorMessage)).toStrictEqual({
        context: "any reason",
        errorCode: "UNKNOWN_ERROR",
        mitigation: {
          retryPeriodInMs: 5000,
          retryWithBlocking: true,
          shouldRetry: true,
        },
      });
    });

    it("should return UNKNOWN_ERROR and shouldRetry = true when error code = INSUFFICIENT_FUNDS_FOR_GAS and nested error code = TRANSACTION_UNDERPRICED", () => {
      const errorMessage = {
        code: "INSUFFICIENT_FUNDS_FOR_GAS",
        message: "",
        error: {
          error: {
            error: {
              code: "TRANSACTION_UNDERPRICED",
            },
            body: '{"error":{"message":"gas required exceeds allowance (0)"}}',
          },
        },
      };

      expect(ErrorParser.parseErrorWithMitigation(errorMessage)).toStrictEqual({
        context: "INSUFFICIENT_FUNDS_FOR_GAS",
        errorCode: "UNKNOWN_ERROR",
        mitigation: {
          retryPeriodInMs: 5000,
          retryWithBlocking: true,
          shouldRetry: true,
        },
      });
    });

    it("should return UNKNOWN_ERROR and shouldRetry = true when error code = NONCE_EXPIRED ", () => {
      const errorMessage = {
        code: "NONCE_EXPIRED",
        message: "nonce has already been used",
        transaction: {
          gasLimit: BigNumber.from(50_000),
          nonce: 1,
        },
      };

      expect(ErrorParser.parseErrorWithMitigation(errorMessage as unknown as EthersError)).toStrictEqual({
        context: "1",
        errorCode: "UNKNOWN_ERROR",
        mitigation: {
          retryWithBlocking: true,
          shouldRetry: true,
        },
      });
    });

    it("should return UNPREDICTABLE_GAS_LIMIT and shouldRetry = false when error code = UNPREDICTABLE_GAS_LIMIT ", () => {
      const errorMessage = {
        code: "UNPREDICTABLE_GAS_LIMIT",
        message: "",
        error: {
          code: "REQUIRE_TRANSACTION",
          data: {
            message: `execution reverted: some reason`,
          },
        },
      };

      expect(ErrorParser.parseErrorWithMitigation(errorMessage as unknown as EthersError)).toStrictEqual({
        context: "UNPREDICTABLE_GAS_LIMIT",
        errorCode: "UNPREDICTABLE_GAS_LIMIT",
        mitigation: {
          shouldRetry: false,
        },
      });
    });

    it("should return NETWORK_ERROR and shouldRetry = true when error code = NETWORK_ERROR ", () => {
      const errorMessage = {
        code: "NETWORK_ERROR",
        message: "",
        error: {
          code: "NETWORK_ERROR",
          data: {
            message: `execution reverted: some reason`,
          },
        },
      };

      expect(ErrorParser.parseErrorWithMitigation(errorMessage as unknown as EthersError)).toStrictEqual({
        context: "NETWORK_ERROR",
        errorCode: "NETWORK_ERROR",
        mitigation: {
          retryPeriodInMs: 5000,
          retryWithBlocking: true,
          shouldRetry: true,
        },
        reason: undefined,
      });
    });

    it("should return INSUFFICIENT_FUNDS and shouldRetry = true when error code = INSUFFICIENT_FUNDS", () => {
      const errorMessage = {
        code: "INSUFFICIENT_FUNDS",
        message: "",
        error: {
          code: "INSUFFICIENT_FUNDS",
          data: {
            message: `execution reverted: some reason`,
          },
        },
      };

      expect(ErrorParser.parseErrorWithMitigation(errorMessage as unknown as EthersError)).toStrictEqual({
        context: "INSUFFICIENT_FUNDS",
        errorCode: "INSUFFICIENT_FUNDS",
        mitigation: {
          retryPeriodInMs: 5000,
          retryWithBlocking: true,
          shouldRetry: true,
        },
        reason: undefined,
      });
    });

    it("should return UNKNOWN_ERROR and shouldRetry = true when error code = TRANSACTION_RAN_OUT_OF_GAS", () => {
      const errorMessage = {
        message: "",
        transaction: {
          gasLimit: BigNumber.from(50_000),
          nonce: 0,
        },
        receipt: {
          gasUsed: BigNumber.from(50_000),
        },
      };

      expect(ErrorParser.parseErrorWithMitigation(errorMessage as unknown as EthersError)).toStrictEqual({
        context: "50000",
        errorCode: "UNKNOWN_ERROR",
        mitigation: {
          shouldRetry: false,
        },
      });
    });
  });
});
