import {
  claimOnL1,
  getL2ToL1MessageStatus,
  getMessageProof,
  getTransactionReceiptByMessageHash,
} from "@consensys/linea-sdk-viem";
import { describe, it, expect, beforeEach, afterEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { getContractEvents, readContract } from "viem/actions";

import { IEthereumGasProvider } from "../../../../core/clients/blockchain/IGasProvider";
import { OnChainMessageStatus } from "../../../../core/enums";
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

const TEST_CONTRACT_ADDRESS = "0x1000000000000000000000000000000000000000";
const TEST_L2_CONTRACT_ADDRESS = "0x5000000000000000000000000000000000000000";
const TEST_TX_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020";
const TEST_MESSAGE_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010";
const TEST_ADDRESS_1 = "0x0000000000000000000000000000000000000001";
const TEST_ADDRESS_2 = "0x0000000000000000000000000000000000000002";
const TEST_MERKLE_ROOT = "0x3000000000000000000000000000000000000000000000000000000000000000";

const testMessageSentEvent = {
  messageHash: TEST_MESSAGE_HASH,
  messageSender: TEST_ADDRESS_1,
  destination: TEST_ADDRESS_2,
  fee: 0n,
  value: 0n,
  messageNonce: 1n,
  calldata: "0x",
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

  describe("getMessageByMessageHash", () => {
    it("returns the first matching MessageSent event mapped to postman format", async () => {
      (getContractEvents as jest.Mock).mockResolvedValue([
        {
          removed: false,
          blockNumber: 51n,
          transactionHash: TEST_TX_HASH,
          logIndex: 1,
          args: {
            _messageHash: TEST_MESSAGE_HASH,
            _from: TEST_ADDRESS_1,
            _to: TEST_ADDRESS_2,
            _fee: 0n,
            _value: 0n,
            _nonce: 1n,
            _calldata: "0x",
          },
        },
      ]);
      const result = await client.getMessageByMessageHash(TEST_MESSAGE_HASH);
      expect(result).toEqual(testMessageSentEvent);
    });

    it("returns null when no events found", async () => {
      (getContractEvents as jest.Mock).mockResolvedValue([]);
      const result = await client.getMessageByMessageHash(TEST_MESSAGE_HASH);
      expect(result).toBeNull();
    });
  });

  describe("getTransactionReceiptByMessageHash", () => {
    it("returns null when SDK throws (message not found)", async () => {
      (getTransactionReceiptByMessageHash as jest.Mock).mockRejectedValue(new Error("not found"));
      const result = await client.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);
      expect(result).toBeNull();
    });

    it("returns mapped receipt when SDK succeeds", async () => {
      (getTransactionReceiptByMessageHash as jest.Mock).mockResolvedValue({
        transactionHash: TEST_TX_HASH as `0x${string}`,
        blockNumber: 51n,
        status: "success",
        gasUsed: 21000n,
        effectiveGasPrice: 1000000000n,
        logs: [],
      });

      const result = await client.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);
      expect(result).not.toBeNull();
      expect(result?.hash).toBe(TEST_TX_HASH);
      expect(result?.status).toBe("success");
    });
  });

  describe("getMessageStatus", () => {
    it("delegates to getMessageStatusUsingMerkleTree via getL2ToL1MessageStatus", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMED);
      const result = await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.CLAIMED);
    });
  });

  describe("getMessageStatusUsingMerkleTree", () => {
    it("returns CLAIMED", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMED);
      const result = await client.getMessageStatusUsingMerkleTree({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.CLAIMED);
    });

    it("returns CLAIMABLE", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      const result = await client.getMessageStatusUsingMerkleTree({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.CLAIMABLE);
    });

    it("returns UNKNOWN", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.UNKNOWN);
      const result = await client.getMessageStatusUsingMerkleTree({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.UNKNOWN);
    });

    it("passes l2PublicClient and contract addresses to SDK", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.UNKNOWN);
      await client.getMessageStatusUsingMerkleTree({ messageHash: TEST_MESSAGE_HASH });
      expect(getL2ToL1MessageStatus).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({
          l2Client: l2PublicClient,
          messageHash: TEST_MESSAGE_HASH,
          lineaRollupAddress: TEST_CONTRACT_ADDRESS,
          l2MessageServiceAddress: TEST_L2_CONTRACT_ADDRESS,
        }),
      );
    });

    it("narrows block range when messageBlockNumber is provided", async () => {
      (getL2ToL1MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      await client.getMessageStatusUsingMerkleTree({ messageHash: TEST_MESSAGE_HASH, messageBlockNumber: 51 });
      expect(getL2ToL1MessageStatus).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({ l2LogsBlockRange: { fromBlock: 51n, toBlock: 51n } }),
      );
    });
  });

  describe("getMessageStatusUsingMessageHash", () => {
    it("returns CLAIMED when status is 2", async () => {
      (readContract as jest.Mock).mockResolvedValue(2n);

      const result = await client.getMessageStatusUsingMessageHash(TEST_MESSAGE_HASH, {});
      expect(result).toBe(OnChainMessageStatus.CLAIMED);
    });

    it("returns CLAIMED when status 0 but MessageClaimed event exists", async () => {
      (readContract as jest.Mock).mockResolvedValue(0n);
      (getContractEvents as jest.Mock).mockResolvedValue([
        {
          removed: false,
          blockNumber: 100n,
          logIndex: 0,
          transactionHash: TEST_TX_HASH,
          args: { _messageHash: TEST_MESSAGE_HASH },
        },
      ]);

      const result = await client.getMessageStatusUsingMessageHash(TEST_MESSAGE_HASH, {});
      expect(result).toBe(OnChainMessageStatus.CLAIMED);
    });

    it("returns UNKNOWN when status 0 and no MessageClaimed event", async () => {
      (readContract as jest.Mock).mockResolvedValue(0n);
      (getContractEvents as jest.Mock).mockResolvedValue([]);

      const result = await client.getMessageStatusUsingMessageHash(TEST_MESSAGE_HASH, {});
      expect(result).toBe(OnChainMessageStatus.UNKNOWN);
    });
  });

  describe("getMessageSiblings", () => {
    it("returns padded siblings for a treeDepth-sized batch", () => {
      const hashes = [TEST_MESSAGE_HASH, "0x" + "ff".repeat(32), "0x" + "aa".repeat(32)];
      const treeDepth = 2; // 2^2 = 4 messages per tree
      const siblings = client.getMessageSiblings(TEST_MESSAGE_HASH, hashes, treeDepth);
      expect(siblings.length).toBe(4); // padded to 4
      expect(siblings[0]).toBe(TEST_MESSAGE_HASH);
    });

    it("throws when message hash not found", () => {
      expect(() => client.getMessageSiblings("0x" + "00".repeat(32), [TEST_MESSAGE_HASH], 2)).toThrow(
        "Message hash not found in messages",
      );
    });
  });

  describe("getFinalizationMessagingInfo", () => {
    const L2_MERKLE_TREE_ADDED_SIG = "0x300e6f978eee6a4b0bba78dd8400dc64fd5652dbfc868a2258e16d0977be222b";
    const L2_MESSAGING_BLOCK_ANCHORED_SIG = "0x3c116827db9db3a30c1a25db8b0ee4bab9d2b223560209cfd839601b621c726d";

    it("throws when receipt has no logs", async () => {
      publicClient.getTransactionReceipt.mockResolvedValue({ logs: [] } as unknown as Awaited<
        ReturnType<PublicClient["getTransactionReceipt"]>
      >);
      await expect(client.getFinalizationMessagingInfo(TEST_TX_HASH)).rejects.toThrow(
        "Transaction does not exist or no logs found",
      );
    });

    it("throws when no L2MerkleRootAdded events found", async () => {
      publicClient.getTransactionReceipt.mockResolvedValue({
        logs: [
          {
            address: TEST_CONTRACT_ADDRESS,
            topics: [
              L2_MESSAGING_BLOCK_ANCHORED_SIG,
              "0x0000000000000000000000000000000000000000000000000000000000000033",
            ],
            data: "0x",
            blockNumber: 100n,
            transactionHash: TEST_TX_HASH,
            logIndex: 0,
          },
        ],
      } as unknown as Awaited<ReturnType<PublicClient["getTransactionReceipt"]>>);

      await expect(client.getFinalizationMessagingInfo(TEST_TX_HASH)).rejects.toThrow(
        "No L2MerkleRootAdded events found",
      );
    });

    it("parses finalization info correctly", async () => {
      const treeDepthHex = "0x0000000000000000000000000000000000000000000000000000000000000005";
      publicClient.getTransactionReceipt.mockResolvedValue({
        logs: [
          {
            address: TEST_CONTRACT_ADDRESS,
            topics: [L2_MERKLE_TREE_ADDED_SIG, TEST_MERKLE_ROOT, treeDepthHex],
            data: "0x",
            blockNumber: 100n,
            transactionHash: TEST_TX_HASH,
            logIndex: 0,
          },
          {
            address: TEST_CONTRACT_ADDRESS,
            topics: [
              L2_MESSAGING_BLOCK_ANCHORED_SIG,
              "0x0000000000000000000000000000000000000000000000000000000000000033",
            ],
            data: "0x",
            blockNumber: 100n,
            transactionHash: TEST_TX_HASH,
            logIndex: 1,
          },
          {
            address: TEST_CONTRACT_ADDRESS,
            topics: [
              L2_MESSAGING_BLOCK_ANCHORED_SIG,
              "0x0000000000000000000000000000000000000000000000000000000000000037",
            ],
            data: "0x",
            blockNumber: 100n,
            transactionHash: TEST_TX_HASH,
            logIndex: 2,
          },
        ],
      } as unknown as Awaited<ReturnType<PublicClient["getTransactionReceipt"]>>);

      const info = await client.getFinalizationMessagingInfo(TEST_TX_HASH);
      expect(info.treeDepth).toBe(5);
      expect(info.l2MerkleRoots).toEqual([TEST_MERKLE_ROOT]);
      expect(info.l2MessagingBlocksRange.startingBlock).toBe(0x33); // 51
      expect(info.l2MessagingBlocksRange.endBlock).toBe(0x37); // 55
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
