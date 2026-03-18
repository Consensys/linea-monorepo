import { claimOnL2, getL1ToL2MessageStatus } from "@consensys/linea-sdk-viem";
import { describe, it, expect, beforeEach, afterEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { decodeErrorResult, type PublicClient, type WalletClient } from "viem";

import { ILineaGasProvider } from "../../../../../core/clients/blockchain/IGasProvider";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../../../../core/enums";
import { ViemL2MessageServiceClient } from "../ViemL2MessageServiceClient";

jest.mock("@consensys/linea-sdk-viem", () => ({
  claimOnL2: jest.fn(),
  getL1ToL2MessageStatus: jest.fn(),
  getMessagesByTransactionHash: jest.fn(),
  getTransactionReceiptByMessageHash: jest.fn(),
}));
jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return { ...actual, decodeErrorResult: jest.fn() };
});
jest.mock("viem/actions", () => ({
  getContractEvents: jest.fn(),
}));

const TEST_CONTRACT_ADDRESS = "0x2000000000000000000000000000000000000000" as `0x${string}`;
const TEST_TX_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020" as `0x${string}`;
const TEST_MESSAGE_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010" as `0x${string}`;
const TEST_ADDRESS_1 = "0x0000000000000000000000000000000000000001" as `0x${string}`;
const TEST_ADDRESS_2 = "0x0000000000000000000000000000000000000002" as `0x${string}`;
const TEST_SIGNER_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" as `0x${string}`;

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

    it("returns true when parsed error name is RateLimitExceeded", async () => {
      const mockTx = {
        to: TEST_ADDRESS_2,
        from: TEST_ADDRESS_1,
        nonce: 1,
        gas: 21000n,
        input: "0x" as `0x${string}`,
        value: 0n,
        maxFeePerGas: 1000n,
        maxPriorityFeePerGas: 100n,
      };
      publicClient.getTransaction.mockResolvedValue(
        mockTx as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      publicClient.call.mockResolvedValue({ data: "0xdeadbeef" } as never);
      (decodeErrorResult as jest.Mock).mockReturnValue({ errorName: "RateLimitExceeded", args: [] });

      const result = await client.isRateLimitExceededError(TEST_TX_HASH);
      expect(result).toBe(true);
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
        claimCycleCount: 0,
      };

      const encoded = client.encodeClaimMessageTransactionData(
        message as Parameters<typeof client.encodeClaimMessageTransactionData>[0],
      );
      expect(encoded).toMatch(/^0x/);
    });

    it("uses provided feeRecipient instead of ZERO_ADDRESS", () => {
      const feeRecipient = "0x7777777777777777777777777777777777777777" as `0x${string}`;
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
        claimCycleCount: 0,
        feeRecipient,
      };

      const encoded = client.encodeClaimMessageTransactionData(
        message as Parameters<typeof client.encodeClaimMessageTransactionData>[0],
      );
      expect(encoded).toMatch(/^0x/);
      expect(encoded.toLowerCase()).toContain(feeRecipient.slice(2).toLowerCase());
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

    it("uses claimViaAddress when provided in opts", async () => {
      const claimViaAddress = "0x9000000000000000000000000000000000000000" as `0x${string}`;
      const mockFees = { maxFeePerGas: 100n, maxPriorityFeePerGas: 10n, gasLimit: 50000n };
      gasProvider.getGasFees.mockResolvedValue(mockFees);

      await client.estimateClaimGasFees(testMessageSentEvent, { claimViaAddress });

      expect(gasProvider.getGasFees).toHaveBeenCalledWith(
        expect.objectContaining({
          to: claimViaAddress,
        }),
      );
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

    it("falls back to zero nonce/gasLimit and undefined fees when no overrides are provided", async () => {
      (claimOnL2 as jest.Mock).mockResolvedValue("0xclaimhash" as `0x${string}`);

      const result = await client.claim(testMessageSentEvent);

      expect(result.hash).toBe("0xclaimhash");
      expect(result.nonce).toBe(0);
      expect(result.gasLimit).toBe(0n);
      expect(result.maxFeePerGas).toBeUndefined();
      expect(result.maxPriorityFeePerGas).toBeUndefined();
      expect(claimOnL2).toHaveBeenCalledWith(
        walletClient,
        expect.objectContaining({
          l2MessageServiceAddress: TEST_CONTRACT_ADDRESS,
          nonce: undefined,
          gas: undefined,
          maxFeePerGas: undefined,
          maxPriorityFeePerGas: undefined,
        }),
      );
    });

    it("uses claimViaAddress when provided", async () => {
      const claimViaAddress = "0x9000000000000000000000000000000000000000" as `0x${string}`;
      (claimOnL2 as jest.Mock).mockResolvedValue("0xclaimhash" as `0x${string}`);

      await client.claim(testMessageSentEvent, { claimViaAddress });

      expect(claimOnL2).toHaveBeenCalledWith(
        walletClient,
        expect.objectContaining({
          l2MessageServiceAddress: claimViaAddress,
        }),
      );
    });

    it("uses feeRecipient from message when provided", async () => {
      const feeRecipient = "0x7777777777777777777777777777777777777777" as `0x${string}`;
      (claimOnL2 as jest.Mock).mockResolvedValue("0xclaimhash" as `0x${string}`);

      await client.claim({ ...testMessageSentEvent, feeRecipient });

      expect(claimOnL2).toHaveBeenCalledWith(
        walletClient,
        expect.objectContaining({
          feeRecipient,
        }),
      );
    });
  });

  describe("parseTransactionError", () => {
    const mockTx = {
      to: TEST_ADDRESS_2,
      from: TEST_ADDRESS_1,
      nonce: 1,
      gas: 21000n,
      input: "0x" as `0x${string}`,
      value: 0n,
      maxFeePerGas: 1000n,
      maxPriorityFeePerGas: 100n,
    };

    it("returns decoded error when call returns revert data", async () => {
      publicClient.getTransaction.mockResolvedValue(
        mockTx as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      publicClient.call.mockResolvedValue({ data: "0xdeadbeef" } as never);
      (decodeErrorResult as jest.Mock).mockReturnValue({ errorName: "SomeError", args: [42n] });

      const result = await client.parseTransactionError(TEST_TX_HASH);
      expect(result).toEqual({ name: "SomeError", args: [42n] });
    });

    it("extracts error data from call exception", async () => {
      publicClient.getTransaction.mockResolvedValue(
        mockTx as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      publicClient.call.mockRejectedValue({ data: "0xcafebabe" as `0x${string}` });
      (decodeErrorResult as jest.Mock).mockReturnValue({ errorName: "CallError", args: [] });

      const result = await client.parseTransactionError(TEST_TX_HASH);
      expect(result).toEqual({ name: "CallError", args: [] });
    });

    it("returns '0x' when call returns no data", async () => {
      publicClient.getTransaction.mockResolvedValue(
        mockTx as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      publicClient.call.mockResolvedValue({ data: undefined } as never);

      const result = await client.parseTransactionError(TEST_TX_HASH);
      expect(result).toBe("0x");
    });

    it("returns raw encoded data when decode throws", async () => {
      publicClient.getTransaction.mockResolvedValue(
        mockTx as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      publicClient.call.mockResolvedValue({ data: "0xdeadbeef" } as never);
      (decodeErrorResult as jest.Mock).mockImplementation(() => {
        throw new Error("decode failed");
      });

      const result = await client.parseTransactionError(TEST_TX_HASH);
      expect(result).toBe("0xdeadbeef");
    });

    it("passes undefined for null tx.to, tx.maxFeePerGas, and tx.maxPriorityFeePerGas", async () => {
      const nullFieldsTx = {
        to: null,
        from: TEST_ADDRESS_1,
        nonce: 1,
        gas: 21000n,
        input: "0x" as `0x${string}`,
        value: 0n,
        maxFeePerGas: null,
        maxPriorityFeePerGas: null,
      };
      publicClient.getTransaction.mockResolvedValue(
        nullFieldsTx as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      publicClient.call.mockResolvedValue({ data: "0xdeadbeef" } as never);
      (decodeErrorResult as jest.Mock).mockReturnValue({ errorName: "SomeError", args: [1n] });

      const result = await client.parseTransactionError(TEST_TX_HASH);

      expect(publicClient.call).toHaveBeenCalledWith(
        expect.objectContaining({
          to: undefined,
          maxFeePerGas: undefined,
          maxPriorityFeePerGas: undefined,
        }),
      );
      expect(result).toEqual({ name: "SomeError", args: [1n] });
    });

    it("returns empty args array when decoded.args is undefined", async () => {
      publicClient.getTransaction.mockResolvedValue(
        mockTx as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );
      publicClient.call.mockResolvedValue({ data: "0xdeadbeef" } as never);
      (decodeErrorResult as jest.Mock).mockReturnValue({ errorName: "SomeError", args: undefined });

      const result = await client.parseTransactionError(TEST_TX_HASH);
      expect(result).toEqual({ name: "SomeError", args: [] });
    });

    it("returns '0x' when getTransaction returns null", async () => {
      publicClient.getTransaction.mockResolvedValue(
        null as unknown as Awaited<ReturnType<PublicClient["getTransaction"]>>,
      );

      const result = await client.parseTransactionError(TEST_TX_HASH);
      expect(result).toBe("0x");
    });
  });
});
