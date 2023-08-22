import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L2MessageServiceContract } from "../../../contracts";
import { getTestL2Signer } from "../../../utils/testHelpers/contracts";
import { TEST_CONTRACT_ADDRESS_2, testL2NetworkConfig } from "../../../utils/testHelpers/constants";
import { L2SentEventListener } from "../";

describe("L2SentEventListener", () => {
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
    it("Should instanciate the L2SentEventListener", () => {
      const l2SentEventListener = new L2SentEventListener(
        mock<DataSource>(),
        messageServiceContract,
        testL2NetworkConfig,
        {
          silent: true,
        },
      );

      expect(l2SentEventListener).toBeDefined();
    });
  });
});
