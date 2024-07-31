import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import {
  ContractTransactionResponse,
  JsonRpcProvider,
  Overrides,
  Signer,
  TransactionReceipt,
  TransactionResponse,
  Wallet,
} from "ethers";
import { TEST_CONTRACT_ADDRESS_2, TEST_L2_SIGNER_PRIVATE_KEY, testMessage } from "../../../utils/testing/constants";
import { LineaGasProvider } from "../../../clients/blockchain/gas/LineaGasProvider";
import { DEFAULT_MAX_CLAIM_GAS_LIMIT, DEFAULT_MAX_FEE_PER_GAS, DEFAULT_PROFIT_MARGIN } from "../../../core/constants";
import { IL2MessageServiceClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { LineaTransactionValidationService } from "../../LineaTransactionValidationService";
import { generateL2MessageServiceClient } from "../../../utils/testing/helpers";
import { L2ChainQuerier } from "../../../clients/blockchain/linea/L2ChainQuerier";

describe("LineaTransactionValidationService", () => {
  let lineaTransactionValidationService: LineaTransactionValidationService;
  let gasProvider: LineaGasProvider;
  let l2ChainQuerier: L2ChainQuerier;
  let l2ContractClient: IL2MessageServiceClient<
    Overrides,
    TransactionReceipt,
    TransactionResponse,
    ContractTransactionResponse,
    Signer
  >;

  beforeEach(() => {
    const clients = generateL2MessageServiceClient(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_2,
      "read-write",
      new Wallet(TEST_L2_SIGNER_PRIVATE_KEY),
      {
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        enforceMaxGasFee: false,
      },
    );

    l2ContractClient = clients.l2MessageServiceClient;
    gasProvider = clients.gasProvider;
    l2ChainQuerier = clients.l2ChainQuerier;

    lineaTransactionValidationService = new LineaTransactionValidationService(
      {
        profitMargin: DEFAULT_PROFIT_MARGIN,
        maxClaimGasLimit: DEFAULT_MAX_CLAIM_GAS_LIMIT,
      },
      l2ChainQuerier,
      l2ContractClient,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("evaluateTransaction", () => {
    it("Should throw an error when there is no extraData in the L2 block", async () => {
      jest.spyOn(l2ContractClient, "getSigner").mockReturnValueOnce(new Wallet(TEST_L2_SIGNER_PRIVATE_KEY));
      const estimatedGasLimit = 50_000n;
      jest.spyOn(gasProvider, "getGasFees").mockResolvedValueOnce({
        gasLimit: estimatedGasLimit,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
      });
      jest.spyOn(l2ChainQuerier, "getBlockExtraData").mockResolvedValueOnce(null);

      await expect(lineaTransactionValidationService.evaluateTransaction(testMessage)).rejects.toThrow("No extra data");
    });

    it("Should return transaction evaluation criteria with hasZeroFee = true", async () => {
      jest.spyOn(l2ContractClient, "getSigner").mockReturnValueOnce(new Wallet(TEST_L2_SIGNER_PRIVATE_KEY));
      const estimatedGasLimit = 50_000n;
      jest.spyOn(gasProvider, "getGasFees").mockResolvedValueOnce({
        gasLimit: estimatedGasLimit,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
      });
      jest.spyOn(l2ChainQuerier, "getBlockExtraData").mockResolvedValueOnce({
        version: 1,
        variableCost: 1_000_000,
        fixedCost: 1_000_000,
        ethGasPrice: 1_000_000,
      });

      testMessage.fee = 0n;
      const criteria = await lineaTransactionValidationService.evaluateTransaction(testMessage);

      expect(criteria).toStrictEqual({
        estimatedGasLimit: estimatedGasLimit,
        hasZeroFee: true,
        isRateLimitExceeded: false,
        isUnderPriced: true,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        threshold: 0,
      });
    });

    it("Should return transaction evaluation criteria with isUnderPriced = true", async () => {
      jest.spyOn(l2ContractClient, "getSigner").mockReturnValueOnce(new Wallet(TEST_L2_SIGNER_PRIVATE_KEY));
      const estimatedGasLimit = 50_000n;
      jest.spyOn(gasProvider, "getGasFees").mockResolvedValueOnce({
        gasLimit: estimatedGasLimit,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
      });
      jest.spyOn(l2ChainQuerier, "getBlockExtraData").mockResolvedValueOnce({
        version: 1,
        variableCost: 1_000_000,
        fixedCost: 1_000_000,
        ethGasPrice: 1_000_000,
      });

      testMessage.fee = 1n;
      const criteria = await lineaTransactionValidationService.evaluateTransaction(testMessage);

      expect(criteria).toStrictEqual({
        estimatedGasLimit: estimatedGasLimit,
        hasZeroFee: false,
        isRateLimitExceeded: false,
        isUnderPriced: true,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        threshold: 0,
      });
    });

    it("Should return transaction evaluation criteria with estimatedGasLimit = null", async () => {
      jest.spyOn(l2ContractClient, "getSigner").mockReturnValueOnce(new Wallet(TEST_L2_SIGNER_PRIVATE_KEY));
      jest.spyOn(gasProvider, "getGasFees").mockResolvedValueOnce({
        gasLimit: DEFAULT_MAX_CLAIM_GAS_LIMIT + 1n,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
      });
      jest.spyOn(l2ChainQuerier, "getBlockExtraData").mockResolvedValueOnce({
        version: 1,
        variableCost: 1_000_000,
        fixedCost: 1_000_000,
        ethGasPrice: 1_000_000,
      });

      const criteria = await lineaTransactionValidationService.evaluateTransaction(testMessage);

      expect(criteria).toStrictEqual({
        estimatedGasLimit: null,
        hasZeroFee: false,
        isRateLimitExceeded: false,
        isUnderPriced: true,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        threshold: 0,
      });
    });

    it("Should return transaction evaluation criteria for a valid message", async () => {
      jest.spyOn(l2ContractClient, "getSigner").mockReturnValueOnce(new Wallet(TEST_L2_SIGNER_PRIVATE_KEY));
      const estimatedGasLimit = 50_000n;
      jest.spyOn(gasProvider, "getGasFees").mockResolvedValueOnce({
        gasLimit: estimatedGasLimit,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
      });
      jest.spyOn(l2ChainQuerier, "getBlockExtraData").mockResolvedValueOnce({
        version: 1,
        variableCost: 1_000_000,
        fixedCost: 1_000_000,
        ethGasPrice: 1_000_000,
      });

      testMessage.fee = 100000000000000000000n;
      const criteria = await lineaTransactionValidationService.evaluateTransaction(testMessage);

      expect(criteria).toStrictEqual({
        estimatedGasLimit: estimatedGasLimit,
        hasZeroFee: false,
        isRateLimitExceeded: false,
        isUnderPriced: false,
        maxFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        maxPriorityFeePerGas: DEFAULT_MAX_FEE_PER_GAS,
        threshold: 2000000000000000,
      });
    });
  });
});
