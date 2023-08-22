import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L1AnchoredEventListener } from "../";
import { L1MessageServiceContract } from "../../../contracts";
import { getTestL1Signer } from "../../../utils/testHelpers/contracts";
import {
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  testL1NetworkConfig,
} from "../../../utils/testHelpers/constants";

describe("L1AnchoredEventListener", () => {
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
    it("Should instanciate the L1AnchoredEventListener", () => {
      const l1AnchoredEventListener = new L1AnchoredEventListener(
        mock<DataSource>(),
        messageServiceContract,
        testL1NetworkConfig,
        TEST_CONTRACT_ADDRESS_2,
        {
          silent: true,
        },
      );

      expect(l1AnchoredEventListener).toBeDefined();
    });
  });
});
