import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { BigNumber } from "ethers";
import { L1MessageServiceContract } from "../../../contracts";
import { getTestL1Signer } from "../../../utils/testHelpers/contracts";
import {
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  testL1NetworkConfig,
} from "../../../utils/testHelpers/constants";
import { L1ClaimTxSender } from "../";
import { ZkEvmV2 } from "../../../../typechain";
import { mockProperty, undoMockProperty } from "../../../utils/testHelpers/helpers";

describe("L1ClaimTxSender", () => {
  let l1ClaimTxSender: L1ClaimTxSender;
  let messageServiceContract: L1MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L1MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_1,
      "read-write",
      getTestL1Signer(),
    );

    l1ClaimTxSender = new L1ClaimTxSender(
      mock<DataSource>(),
      messageServiceContract,
      testL1NetworkConfig,
      TEST_CONTRACT_ADDRESS_2,
      {
        silent: true,
      },
    );
  });

  describe("isRateLimitExceeded", () => {
    it("should return true if withdrawal rate limit is exceeded", async () => {
      const messageFee = "1000000000";
      const messageValue = "1000000000000";
      mockProperty(messageServiceContract, "contract", {
        ...messageServiceContract.contract,
        limitInWei: jest.fn().mockResolvedValueOnce(BigNumber.from(1000000000000)),
        currentPeriodAmountInWei: jest.fn().mockResolvedValueOnce(BigNumber.from(1000000000000)),
      } as unknown as ZkEvmV2);

      expect(await l1ClaimTxSender.isRateLimitExceeded(messageFee, messageValue)).toStrictEqual(true);

      undoMockProperty(messageServiceContract, "contract");
    });

    it("should return false if withdrawal rate limit is not exceeded", async () => {
      const messageFee = "1000000000";
      const messageValue = "1000000";
      mockProperty(messageServiceContract, "contract", {
        ...messageServiceContract.contract,
        limitInWei: jest.fn().mockResolvedValueOnce(BigNumber.from(1000000000000)),
        currentPeriodAmountInWei: jest.fn().mockResolvedValueOnce(BigNumber.from(1000000)),
      } as unknown as ZkEvmV2);

      expect(await l1ClaimTxSender.isRateLimitExceeded(messageFee, messageValue)).toStrictEqual(false);

      undoMockProperty(messageServiceContract, "contract");
    });
  });
});
