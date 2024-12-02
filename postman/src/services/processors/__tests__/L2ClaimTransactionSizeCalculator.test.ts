import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import {
  ContractTransactionResponse,
  Overrides,
  Signer,
  TransactionReceipt,
  TransactionResponse,
  Wallet,
} from "ethers";
import { BaseError, LineaProvider } from "@consensys/linea-sdk";
import { TEST_CONTRACT_ADDRESS_2, TEST_L2_SIGNER_PRIVATE_KEY, testMessage } from "../../../utils/testing/constants";
import { EthereumMessageDBService } from "../../persistence/EthereumMessageDBService";
import { L2ClaimTransactionSizeCalculator } from "../../L2ClaimTransactionSizeCalculator";
import { DEFAULT_MAX_FEE_PER_GAS } from "../../../core/constants";
import { IL2MessageServiceClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { serialize } from "../../../core/utils/serialize";
import { generateL2MessageServiceClient } from "../../../utils/testing/helpers";

describe("L2ClaimTransactionSizeCalculator", () => {
  let transactionSizeCalculator: L2ClaimTransactionSizeCalculator;

  const databaseService = mock<EthereumMessageDBService>();
  let l2ContractClient: IL2MessageServiceClient<
    Overrides,
    TransactionReceipt,
    TransactionResponse,
    ContractTransactionResponse,
    Signer
  >;

  beforeEach(() => {
    const clients = generateL2MessageServiceClient(
      mock<LineaProvider>(),
      TEST_CONTRACT_ADDRESS_2,
      "read-only",
      undefined,
      {
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        enforceMaxGasFee: false,
      },
    );

    l2ContractClient = clients.l2MessageServiceClient;
    transactionSizeCalculator = new L2ClaimTransactionSizeCalculator(l2ContractClient);
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("process", () => {
    it("Should throw an error if  signer is undefined", async () => {
      jest.spyOn(databaseService, "getNFirstMessagesByStatus").mockResolvedValue([]);

      await expect(
        transactionSizeCalculator.calculateTransactionSize(testMessage, {
          gasLimit: 50_000n,
          maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
          maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        }),
      ).rejects.toThrow(
        new BaseError(`Transaction size calculation error: ${serialize(new BaseError("Signer is undefined."))}`),
      );
    });

    it("Should return transaction size", async () => {
      jest.spyOn(l2ContractClient, "getSigner").mockReturnValueOnce(new Wallet(TEST_L2_SIGNER_PRIVATE_KEY));

      const transactionSize = await transactionSizeCalculator.calculateTransactionSize(testMessage, {
        gasLimit: 50_000n,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
      });

      expect(transactionSize).toStrictEqual(77);
    });
  });
});
