import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L2MessageServiceContract } from "../../../contracts";
import { getTestL2Signer } from "../../../utils/testHelpers/contracts";
import { TEST_CONTRACT_ADDRESS_2, testL2NetworkConfig } from "../../../utils/testHelpers/constants";
import { L2ClaimStatusWatcher } from "../";

describe("L2ClaimStatusWatcher", () => {
  let l2ClaimStatusWatcher: L2ClaimStatusWatcher;
  let messageServiceContract: L2MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L2MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_2,
      "read-write",
      getTestL2Signer(),
    );

    l2ClaimStatusWatcher = new L2ClaimStatusWatcher(mock<DataSource>(), messageServiceContract, testL2NetworkConfig, {
      silent: true,
    });
  });

  describe("isRateLimitExceededError", () => {
    it("should return false as there is no rate limit mechanism from L1 to L2", async () => {
      expect(
        await l2ClaimStatusWatcher.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(false);
    });
  });
});
