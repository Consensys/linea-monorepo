import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L1MessageServiceContract } from "../../../contracts";
import { getTestL1Signer } from "../../../utils/testHelpers/contracts";
import { TEST_CONTRACT_ADDRESS_1, testL1NetworkConfig } from "../../../utils/testHelpers/constants";
import { L1SentEventListener } from "../";

describe("L1SentEventListener", () => {
  let messageServiceContract: L1MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L1MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_1,
      "read-write",
      getTestL1Signer(),
    );
  });

  describe("constructor", () => {
    it("Should instanciate the L1SentEventListener", () => {
      const l1SentEventListener = new L1SentEventListener(
        mock<DataSource>(),
        messageServiceContract,
        testL1NetworkConfig,
        {
          silent: true,
        },
      );

      expect(l1SentEventListener).toBeDefined();
    });
  });
});
