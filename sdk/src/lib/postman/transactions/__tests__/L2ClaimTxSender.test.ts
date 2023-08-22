import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L2MessageServiceContract } from "../../../contracts";
import { getTestL2Signer } from "../../../utils/testHelpers/contracts";
import {
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  testL2NetworkConfig,
} from "../../../utils/testHelpers/constants";
import { L2ClaimTxSender } from "../";

describe("L2ClaimTxSender", () => {
  let l2ClaimTxSender: L2ClaimTxSender;
  let messageServiceContract: L2MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L2MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_2,
      "read-write",
      getTestL2Signer(),
    );

    l2ClaimTxSender = new L2ClaimTxSender(
      mock<DataSource>(),
      messageServiceContract,
      testL2NetworkConfig,
      TEST_CONTRACT_ADDRESS_1,
      {
        silent: true,
      },
    );
  });

  describe("isRateLimitExceeded", () => {
    it("should return false as there is no rate limit mechanism from L1 to L2", async () => {
      expect(await l2ClaimTxSender.isRateLimitExceeded("10000", "20000")).toStrictEqual(false);
    });
  });
});
