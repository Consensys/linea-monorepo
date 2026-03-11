import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

jest.mock("@consensys/linea-native-libs", () => ({
  GoNativeCompressor: class {
    constructor(_dataLimit: number) {}
    getCompressedTxSize(_data: Uint8Array): number {
      return 100;
    }
  },
}));

import { IL2MessageServiceClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { ITransactionSigner } from "../../../core/services/ITransactionSigner";
import { DEFAULT_MAX_FEE_PER_GAS, testMessage } from "../../../utils/testing/constants";
import { L2ClaimTransactionSizeCalculator } from "../../L2ClaimTransactionSizeCalculator";

describe("L2ClaimTransactionSizeCalculator", () => {
  let transactionSizeCalculator: L2ClaimTransactionSizeCalculator;

  const l2ContractClient = mock<IL2MessageServiceClient>();
  const transactionSigner = mock<ITransactionSigner>();

  beforeEach(() => {
    transactionSizeCalculator = new L2ClaimTransactionSizeCalculator(l2ContractClient, transactionSigner);
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("process", () => {
    it("Should throw an error if encodeClaimMessageTransactionData throws", async () => {
      jest.spyOn(l2ContractClient, "encodeClaimMessageTransactionData").mockImplementation(() => {
        throw new Error("encode error");
      });

      await expect(
        transactionSizeCalculator.calculateTransactionSize(testMessage, {
          gasLimit: 50_000n,
          maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
          maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        }),
      ).rejects.toThrow("encode error");
    });

    it("Should return transaction size", async () => {
      jest.spyOn(l2ContractClient, "encodeClaimMessageTransactionData").mockReturnValue("0x1234");
      jest.spyOn(l2ContractClient, "getContractAddress").mockReturnValue("0x0000000000000000000000000000000000000001");
      jest.spyOn(transactionSigner, "signAndSerialize").mockResolvedValue(new Uint8Array(50));

      const transactionSize = await transactionSizeCalculator.calculateTransactionSize(testMessage, {
        gasLimit: 50_000n,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
      });

      expect(transactionSize).toBeGreaterThan(0);
    });
  });
});
