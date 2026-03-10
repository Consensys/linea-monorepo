import { claimOnL1, getL2ToL1MessageStatus, getMessageProof } from "@consensys/linea-sdk-viem";
import { describe, it, expect, beforeEach, afterEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { readContract } from "viem/actions";

import { IEthereumGasProvider } from "../../../../../core/clients/blockchain/IGasProvider";
import { OnChainMessageStatus } from "../../../../../core/enums";
import { ViemLineaRollupClient } from "../ViemLineaRollupClient";

import type { PublicClient, WalletClient } from "viem";

jest.mock("@consensys/linea-sdk-viem", () => ({
  claimOnL1: jest.fn(),
  getL2ToL1MessageStatus: jest.fn(),
  getMessageProof: jest.fn(),
  getMessagesByTransactionHash: jest.fn(),
  getTransactionReceiptByMessageHash: jest.fn(),
}));
jest.mock("viem/actions", () => ({
  estimateContractGas: jest.fn(),
  getContractEvents: jest.fn(),
  readContract: jest.fn(),
}));

const TEST_CONTRACT_ADDRESS = "0x1000000000000000000000000000000000000000" as `0x${string}`;
const TEST_L2_CONTRACT_ADDRESS = "0x5000000000000000000000000000000000000000" as `0x${string}`;
const TEST_TX_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020" as `0x${string}`;
const TEST_MESSAGE_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010" as `0x${string}`;
const TEST_ADDRESS_1 = "0x0000000000000000000000000000000000000001" as `0x${string}`;
const TEST_ADDRESS_2 = "0x0000000000000000000000000000000000000002" as `0x${string}`;
const TEST_MERKLE_ROOT = "0x3000000000000000000000000000000000000000000000000000000000000000" as `0x${string}`;

const testMessageSentEvent = {
  messageHash: TEST_MESSAGE_HASH,
  messageSender: TEST_ADDRESS_1,
  destination: TEST_ADDRESS_2,
  fee: 0n,
  value: 0n,
  messageNonce: 1n,
  calldata: "0x" as `0x${string}`,
  contractAddress: TEST_CONTRACT_ADDRESS,
  blockNumber: 51,
  transactionHash: TEST_TX_HASH,
  logIndex: 1,
};

describe("ViemLineaRollupClient", () => {
  let publicClient: ReturnType<typeof mock<PublicClient>>;
  let walletClient: ReturnType<typeof mock<WalletClient>>;
  let l2PublicClient: ReturnType<typeof mock<PublicClient>>;
  let gasProvider: ReturnType<typeof mock<IEthereumGasProvider>>;
  let client: ViemLineaRollupClient;

  beforeEach(() => {
    publicClient = mock<PublicClient>();
    walletClient = mock<WalletClient>();
    l2PublicClient = mock<PublicClient>();
    gasProvider = mock<IEthereumGasProvider>();

    client = new ViemLineaRollupClient(
      publicClient,
      walletClient,
      TEST_CONTRACT_ADDRESS,
      l2PublicClient,
      TEST_L2_CONTRACT_ADDRESS,
      gasProvider,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("getMessageStatus", () => {
    it("returns CLAIMED", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMED);
      const result = await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.CLAIMED);
    });

    it("returns CLAIMABLE", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      const result = await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.CLAIMABLE);
    });

    it("returns UNKNOWN", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.UNKNOWN);
      const result = await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.UNKNOWN);
    });

    it("narrows block range when messageBlockNumber is provided", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH, messageBlockNumber: 51 });
      expect(getL2ToL1MessageStatus).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({ l2LogsBlockRange: { fromBlock: 51n, toBlock: 51n } }),
      );
    });
  });

  describe("isRateLimitExceeded", () => {
    it("returns false when total is within limit", async () => {
      (readContract as jest.Mock)
        .mockResolvedValueOnce(1000n) // limitInWei
        .mockResolvedValueOnce(100n); // currentPeriodAmountInWei

      const result = await client.isRateLimitExceeded(50n, 50n);
      // 100 + 50 + 50 = 200; 1000 * 0.95 = 950; 200 <= 950
      expect(result).toBe(false);
    });

    it("returns true when total exceeds limit * margin", async () => {
      (readContract as jest.Mock)
        .mockResolvedValueOnce(1000n) // limitInWei
        .mockResolvedValueOnce(950n); // currentPeriodAmountInWei

      const result = await client.isRateLimitExceeded(100n, 0n);
      // 950 + 100 + 0 = 1050; 1000 * 0.95 = 950; 1050 > 950
      expect(result).toBe(true);
    });
  });

  describe("isRateLimitExceededError", () => {
    it("returns false for string errors", async () => {
      publicClient.getTransaction.mockResolvedValue(
        null as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      const result = await client.isRateLimitExceededError(TEST_TX_HASH);
      expect(result).toBe(false);
    });
  });

  describe("getMessageProof", () => {
    it("delegates to SDK getMessageProof with correct params", async () => {
      const mockProof = {
        proof: [TEST_MERKLE_ROOT as `0x${string}`],
        root: TEST_MERKLE_ROOT as `0x${string}`,
        leafIndex: 0,
      };
      (getMessageProof as jest.Mock).mockResolvedValue(mockProof);

      const result = await client.getMessageProof(TEST_MESSAGE_HASH);
      expect(result).toEqual(mockProof);
      expect(getMessageProof).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({
          l2Client: l2PublicClient,
          messageHash: TEST_MESSAGE_HASH,
          lineaRollupAddress: TEST_CONTRACT_ADDRESS,
          l2MessageServiceAddress: TEST_L2_CONTRACT_ADDRESS,
        }),
      );
    });
  });

  describe("claim", () => {
    it("uses claimOnL1 from sdk-viem and returns TransactionSubmission", async () => {
      (getMessageProof as jest.Mock).mockResolvedValue({
        proof: [TEST_MERKLE_ROOT as `0x${string}`],
        root: TEST_MERKLE_ROOT as `0x${string}`,
        leafIndex: 0,
      });
      gasProvider.getGasFees.mockResolvedValue({ maxFeePerGas: 1000n, maxPriorityFeePerGas: 100n });
      (claimOnL1 as jest.Mock).mockResolvedValue("0xclaimhash" as `0x${string}`);

      const result = await client.claim(testMessageSentEvent, {
        overrides: { nonce: 2, gasLimit: 150000n, maxFeePerGas: 800n, maxPriorityFeePerGas: 80n },
      });

      expect(result.hash).toBe("0xclaimhash");
      expect(result.nonce).toBe(2);
      expect(result.gasLimit).toBe(150000n);
      expect(claimOnL1).toHaveBeenCalledWith(
        walletClient,
        expect.objectContaining({
          messageNonce: testMessageSentEvent.messageNonce,
          lineaRollupAddress: TEST_CONTRACT_ADDRESS,
        }),
      );
    });
  });

  describe("retryTransactionWithHigherFee", () => {
    it("throws when priceBumpPercent is not an integer", async () => {
      await expect(client.retryTransactionWithHigherFee(TEST_TX_HASH, 1.5)).rejects.toThrow(
        "'priceBumpPercent' must be an integer",
      );
    });

    it("throws when transaction not found", async () => {
      publicClient.getTransaction.mockResolvedValue(
        null as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      await expect(client.retryTransactionWithHigherFee(TEST_TX_HASH)).rejects.toThrow("not found");
    });

    it("bumps fees and sends transaction", async () => {
      publicClient.getTransaction.mockResolvedValue({
        hash: TEST_TX_HASH as `0x${string}`,
        nonce: 5,
        gas: 100000n,
        maxFeePerGas: 1000n,
        maxPriorityFeePerGas: 100n,
        to: TEST_ADDRESS_2 as `0x${string}`,
        value: 0n,
        input: "0x" as `0x${string}`,
      } as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>);

      gasProvider.getMaxFeePerGas.mockReturnValue(100_000_000n);
      walletClient.sendTransaction.mockResolvedValue("0xnewhash" as `0x${string}`);

      const result = await client.retryTransactionWithHigherFee(TEST_TX_HASH, 10);

      expect(result.hash).toBe("0xnewhash");
      expect(result.nonce).toBe(5);
      // 1000 * 110 / 100 = 1100
      expect(result.maxFeePerGas).toBe(1100n);
      expect(result.maxPriorityFeePerGas).toBe(110n);
    });

    it("caps fees at gasProvider.getMaxFeePerGas()", async () => {
      publicClient.getTransaction.mockResolvedValue({
        hash: TEST_TX_HASH as `0x${string}`,
        nonce: 5,
        gas: 100000n,
        maxFeePerGas: 1000n,
        maxPriorityFeePerGas: 1000n,
        to: TEST_ADDRESS_2 as `0x${string}`,
        value: 0n,
        input: "0x" as `0x${string}`,
      } as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>);

      gasProvider.getMaxFeePerGas.mockReturnValue(500n); // cap at 500
      walletClient.sendTransaction.mockResolvedValue("0xnewhash" as `0x${string}`);

      const result = await client.retryTransactionWithHigherFee(TEST_TX_HASH, 10);
      expect(result.maxFeePerGas).toBe(500n);
      expect(result.maxPriorityFeePerGas).toBe(500n);
    });
  });
});
