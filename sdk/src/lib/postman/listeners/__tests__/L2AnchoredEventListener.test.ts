import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L2AnchoredEventListener } from "../";
import { L2MessageServiceContract } from "../../../contracts";
import { getTestL2Signer } from "../../../utils/testHelpers/contracts";
import {
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  testL2NetworkConfig,
} from "../../../utils/testHelpers/constants";

describe("L2AnchoredEventListener", () => {
  let messageServiceContract: L2MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L2MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_2,
      "read-write",
      getTestL2Signer(),
    );
  });

  describe("constructor", () => {
    it("Should instanciate the L2AnchoredEventListener", () => {
      const l2AnchoredEventListener = new L2AnchoredEventListener(
        mock<DataSource>(),
        messageServiceContract,
        testL2NetworkConfig,
        TEST_CONTRACT_ADDRESS_1,
        {
          silent: true,
        },
      );

      expect(l2AnchoredEventListener).toBeDefined();
    });
  });
});
