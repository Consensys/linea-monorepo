import { claimOnL2, getL1ToL2MessageStatus, getTransactionReceiptByMessageHash } from "@consensys/linea-sdk-viem";
import { describe, it, expect, beforeEach, afterEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { getContractEvents } from "viem/actions";

import { ILineaGasProvider } from "../../../../core/clients/blockchain/IGasProvider";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../../../core/enums";
import { ViemL2MessageServiceClient } from "../ViemL2MessageServiceClient";

import type { PublicClient, WalletClient } from "viem";

jest.mock("@consensys/linea-sdk-viem", () => ({
  claimOnL2: jest.fn(),
  getL1ToL2MessageStatus: jest.fn(),
  getMessagesByTransactionHash: jest.fn(),
  getTransactionReceiptByMessageHash: jest.fn(),
}));
jest.mock("viem/actions", () => ({
  getContractEvents: jest.fn(),
}));

const TEST_CONTRACT_ADDRESS = "0x2000000000000000000000000000000000000000";
const TEST_TX_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020";
const TEST_MESSAGE_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010";
const TEST_ADDRESS_1 = "0x0000000000000000000000000000000000000001";
const TEST_ADDRESS_2 = "0x0000000000000000000000000000000000000002";
const TEST_SIGNER_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";

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
  logIndex: 0,
};

describe("ViemL2MessageServiceClient", () => {
  let publicClient: ReturnType<typeof mock<PublicClient>>;
  let walletClient: ReturnType<typeof mock<WalletClient>>;
  let gasProvider: ReturnType<typeof mock<ILineaGasProvider>>;
  let client: ViemL2MessageServiceClient;

  beforeEach(() => {
    publicClient = mock<PublicClient>();
    walletClient = mock<WalletClient>();
    gasProvider = mock<ILineaGasProvider>();

    client = new ViemL2MessageServiceClient(
      publicClient,
      walletClient,
      TEST_CONTRACT_ADDRESS,
      gasProvider,
      TEST_SIGNER_ADDRESS,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("getContractAddress", () => {
    it("returns the contract address", () => {
      expect(client.getContractAddress()).toBe(TEST_CONTRACT_ADDRESS);
    });
  });

  describe("getMessageByMessageHash", () => {
    it("returns the first matching event mapped to postman format", async () => {
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
      expect(result).toMatchObject({
        messageHash: TEST_MESSAGE_HASH,
        messageSender: TEST_ADDRESS_1,
        destination: TEST_ADDRESS_2,
        fee: 0n,
        value: 0n,
        messageNonce: 1n,
        calldata: "0x",
        blockNumber: 51,
        transactionHash: TEST_TX_HASH,
        logIndex: 1,
      });
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
    });
  });

  describe("getMessageStatus", () => {
    it("returns UNKNOWN", async () => {
      (getL1ToL2MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.UNKNOWN);
      const result = await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.UNKNOWN);
    });

    it("returns CLAIMABLE", async () => {
      (getL1ToL2MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      const result = await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.CLAIMABLE);
    });

    it("returns CLAIMED", async () => {
      (getL1ToL2MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMED);
      const result = await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(result).toBe(OnChainMessageStatus.CLAIMED);
    });

    it("calls SDK with correct params", async () => {
      (getL1ToL2MessageStatus as jest.Mock).mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      await client.getMessageStatus({ messageHash: TEST_MESSAGE_HASH });
      expect(getL1ToL2MessageStatus).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({ messageHash: TEST_MESSAGE_HASH, l2MessageServiceAddress: TEST_CONTRACT_ADDRESS }),
      );
    });
  });

  describe("isRateLimitExceeded", () => {
    it("always returns false", async () => {
      const result = await client.isRateLimitExceeded(1000000n, 1000000n);
      expect(result).toBe(false);
    });
  });

  describe("isRateLimitExceededError", () => {
    it("returns false for null transaction", async () => {
      publicClient.getTransaction.mockResolvedValue(
        null as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      const result = await client.isRateLimitExceededError(TEST_TX_HASH);
      expect(result).toBe(false);
    });
  });

  describe("encodeClaimMessageTransactionData", () => {
    it("returns encoded function data starting with 0x", () => {
      const message = {
        messageSender: TEST_ADDRESS_1,
        destination: TEST_ADDRESS_2,
        fee: 0n,
        value: 0n,
        messageNonce: 1n,
        calldata: "0x",
        messageHash: TEST_MESSAGE_HASH,
        contractAddress: TEST_CONTRACT_ADDRESS,
        sentBlockNumber: 51,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
      };

      const encoded = client.encodeClaimMessageTransactionData(
        message as Parameters<typeof client.encodeClaimMessageTransactionData>[0],
      );
      expect(encoded).toMatch(/^0x/);
    });
  });

  describe("estimateClaimGasFees", () => {
    it("calls gas provider with encoded calldata", async () => {
      const mockFees = { maxFeePerGas: 100n, maxPriorityFeePerGas: 10n, gasLimit: 50000n };
      gasProvider.getGasFees.mockResolvedValue(mockFees);

      const result = await client.estimateClaimGasFees(testMessageSentEvent);

      expect(gasProvider.getGasFees).toHaveBeenCalledWith(
        expect.objectContaining({
          from: TEST_SIGNER_ADDRESS,
          to: TEST_CONTRACT_ADDRESS,
          value: 0n,
        }),
      );
      expect(result).toEqual(mockFees);
    });
  });

  describe("claim", () => {
    it("uses claimOnL2 from sdk-viem and returns TransactionSubmission", async () => {
      (claimOnL2 as jest.Mock).mockResolvedValue("0xclaimhash" as `0x${string}`);

      const result = await client.claim(testMessageSentEvent, {
        overrides: { nonce: 3, gasLimit: 200000n, maxFeePerGas: 500n, maxPriorityFeePerGas: 50n },
      });

      expect(result.hash).toBe("0xclaimhash");
      expect(result.nonce).toBe(3);
      expect(result.gasLimit).toBe(200000n);
      expect(claimOnL2).toHaveBeenCalledWith(
        walletClient,
        expect.objectContaining({
          messageNonce: testMessageSentEvent.messageNonce,
          l2MessageServiceAddress: TEST_CONTRACT_ADDRESS,
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
      walletClient.sendTransaction.mockResolvedValue("0xretryhash" as `0x${string}`);

      const result = await client.retryTransactionWithHigherFee(TEST_TX_HASH, 10);

      expect(result.hash).toBe("0xretryhash");
      expect(result.nonce).toBe(5);
      expect(result.maxFeePerGas).toBe(1100n); // 1000 * 110 / 100
      expect(result.maxPriorityFeePerGas).toBe(110n); // 100 * 110 / 100
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

      gasProvider.getMaxFeePerGas.mockReturnValue(500n);
      walletClient.sendTransaction.mockResolvedValue("0xretryhash" as `0x${string}`);

      const result = await client.retryTransactionWithHigherFee(TEST_TX_HASH, 10);
      expect(result.maxFeePerGas).toBe(500n);
      expect(result.maxPriorityFeePerGas).toBe(500n);
    });

    it("fetches current fees when tx lacks maxFeePerGas", async () => {
      publicClient.getTransaction.mockResolvedValue({
        hash: TEST_TX_HASH as `0x${string}`,
        nonce: 5,
        gas: 100000n,
        maxFeePerGas: null,
        maxPriorityFeePerGas: null,
        to: TEST_ADDRESS_2 as `0x${string}`,
        value: 0n,
        input: "0x" as `0x${string}`,
      } as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>);

      publicClient.estimateFeesPerGas.mockResolvedValue({
        maxFeePerGas: 2000n,
        maxPriorityFeePerGas: 200n,
      } as Awaited<ReturnType<PublicClient["estimateFeesPerGas"]>>);
      walletClient.sendTransaction.mockResolvedValue("0xretryhash" as `0x${string}`);

      const result = await client.retryTransactionWithHigherFee(TEST_TX_HASH, 10);
      expect(result.maxFeePerGas).toBe(2000n);
      expect(result.maxPriorityFeePerGas).toBe(200n);
    });
  });
});
