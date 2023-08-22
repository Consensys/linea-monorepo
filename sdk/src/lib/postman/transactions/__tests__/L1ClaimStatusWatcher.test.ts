import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L1MessageServiceContract } from "../../../contracts";
import { getTestL1Signer } from "../../../utils/testHelpers/contracts";
import { TEST_CONTRACT_ADDRESS_1, testL1NetworkConfig } from "../../../utils/testHelpers/constants";
import { L1ClaimStatusWatcher } from "../";
import { generateTransactionResponse } from "../../../utils/testHelpers/helpers";

describe("L1ClaimStatusWatcher", () => {
  let l1ClaimStatusWatcher: L1ClaimStatusWatcher;
  let messageServiceContract: L1MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L1MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_1,
      "read-write",
      getTestL1Signer(),
    );

    l1ClaimStatusWatcher = new L1ClaimStatusWatcher(mock<DataSource>(), messageServiceContract, testL1NetworkConfig, {
      silent: true,
    });
  });

  describe("isRateLimitExceededError", () => {
    it("should return false when something went wrong (http error etc)", async () => {
      jest.spyOn(messageServiceContract.provider, "getTransaction").mockRejectedValueOnce({});

      expect(
        await l1ClaimStatusWatcher.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(false);
    });

    it("should return false when transaction revert reason is not RateLimitExceeded", async () => {
      jest
        .spyOn(messageServiceContract.provider, "getTransaction")
        .mockResolvedValueOnce(generateTransactionResponse());
      jest.spyOn(messageServiceContract.provider, "call").mockResolvedValueOnce("0xa74c1c6d");

      expect(
        await l1ClaimStatusWatcher.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(false);
    });

    it("should return true when transaction revert reason is RateLimitExceeded", async () => {
      jest
        .spyOn(messageServiceContract.provider, "getTransaction")
        .mockResolvedValueOnce(generateTransactionResponse());
      jest.spyOn(messageServiceContract.provider, "call").mockResolvedValueOnce("0xa74c1c5f");

      expect(
        await l1ClaimStatusWatcher.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(true);
    });
  });
});
