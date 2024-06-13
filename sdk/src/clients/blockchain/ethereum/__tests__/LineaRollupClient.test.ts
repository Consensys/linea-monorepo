import { describe, afterEach, it, expect, beforeEach } from "@jest/globals";
import { MockProxy, mock } from "jest-mock-extended";
import { ContractTransactionResponse, JsonRpcProvider, Wallet } from "ethers";
import {
  testMessageSentEvent,
  TEST_MESSAGE_HASH,
  TEST_CONTRACT_ADDRESS_1,
  TEST_TRANSACTION_HASH,
  TEST_ADDRESS_2,
  testMessageClaimedEvent,
  testL2MessagingBlockAnchoredEvent,
  TEST_MERKLE_ROOT,
  TEST_MESSAGE_HASH_2,
  TEST_MERKLE_ROOT_2,
} from "../../../../utils/testing/constants";
import { LineaRollup, LineaRollup__factory } from "../../typechain";
import {
  generateL2MerkleTreeAddedLog,
  generateL2MessagingBlockAnchoredLog,
  generateMessage,
  generateTransactionReceipt,
  generateTransactionReceiptWithLogs,
  generateTransactionResponse,
  mockProperty,
} from "../../../../utils/testing/helpers";
import { IL2MessageServiceLogClient } from "../../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";
import { LineaRollupClient } from "../LineaRollupClient";
import { ZERO_ADDRESS } from "../../../../core/constants";
import { OnChainMessageStatus } from "../../../../core/enums/MessageEnums";
import { GasEstimationError } from "../../../../core/errors/GasFeeErrors";
import { BaseError } from "../../../../core/errors/Base";
import { ILineaRollupLogClient } from "../../../../core/clients/blockchain/ethereum/ILineaRollupLogClient";

describe("TestLineaRollupClient", () => {
  let providerMock: MockProxy<JsonRpcProvider>;
  let walletMock: MockProxy<Wallet>;
  let lineaRollupMock: MockProxy<LineaRollup>;
  let lineaRollupLogClientMock: MockProxy<ILineaRollupLogClient>;
  let l2MessageServiceLogClientMock: MockProxy<IL2MessageServiceLogClient>;
  let lineaRolliupClient: LineaRollupClient;

  beforeEach(() => {
    providerMock = mock<JsonRpcProvider>();
    walletMock = mock<Wallet>();
    l2MessageServiceLogClientMock = mock<IL2MessageServiceLogClient>();
    lineaRollupLogClientMock = mock<ILineaRollupLogClient>();
    lineaRollupMock = mock<LineaRollup>();
    lineaRolliupClient = new LineaRollupClient(
      providerMock,
      TEST_CONTRACT_ADDRESS_1,
      lineaRollupLogClientMock,
      l2MessageServiceLogClientMock,
      "read-write",
      walletMock,
    );
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe("constructor", () => {
    it("should throw an error when mode = 'read-write' and this.signer is undefined", async () => {
      expect(
        () =>
          new LineaRollupClient(
            providerMock,
            TEST_CONTRACT_ADDRESS_1,
            lineaRollupLogClientMock,
            l2MessageServiceLogClientMock,
            "read-write",
          ),
      ).toThrowError(new BaseError("Please provide a signer."));
    });
  });

  describe("getMessageByMessageHash", () => {
    it("should return a MessageSent", async () => {
      jest.spyOn(lineaRollupLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);

      const messageSentEvent = await lineaRolliupClient.getMessageByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentEvent).toStrictEqual(testMessageSentEvent);
    });

    it("should return null if empty events returned", async () => {
      jest.spyOn(lineaRollupLogClientMock, "getMessageSentEvents").mockResolvedValue([]);

      const messageSentEvent = await lineaRolliupClient.getMessageByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentEvent).toStrictEqual(null);
    });
  });

  describe("getMessagesByTransactionHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(null);

      const messageSentEvents = await lineaRolliupClient.getMessagesByTransactionHash(TEST_TRANSACTION_HASH);

      expect(messageSentEvents).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);
      jest.spyOn(lineaRollupLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);

      const messageSentEvents = await lineaRolliupClient.getMessagesByTransactionHash(TEST_MESSAGE_HASH);

      expect(messageSentEvents).toStrictEqual([testMessageSentEvent]);
    });
  });

  describe("getTransactionReceiptByMessageHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(lineaRollupLogClientMock, "getMessageSentEvents").mockResolvedValue([]);

      const messageSentTxReceipt = await lineaRolliupClient.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentTxReceipt).toStrictEqual(null);
    });

    it("should return null when transaction receipt does not exist", async () => {
      jest.spyOn(lineaRollupLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(null);

      const messageSentTxReceipt = await lineaRolliupClient.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentTxReceipt).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(lineaRollupLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);

      const messageSentTxReceipt = await lineaRolliupClient.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentTxReceipt).toStrictEqual(transactionReceipt);
    });
  });

  describe("getMessageStatusUsingMessageHash", () => {
    it("should return UNKNOWN when on chain message status === 0 and no claimed event was found", async () => {
      jest.spyOn(lineaRollupMock, "inboxL2L1MessageStatus").mockResolvedValue(0n);
      jest.spyOn(lineaRollupLogClientMock, "getMessageClaimedEvents").mockResolvedValue([]);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatusUsingMessageHash(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.UNKNOWN);
    });

    it("should return CLAIMABLE when on chain message status === 1", async () => {
      jest.spyOn(lineaRollupMock, "inboxL2L1MessageStatus").mockResolvedValue(1n);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatusUsingMessageHash(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMABLE);
    });

    it("should return CLAIMED when on chain message status === 0 and the claimed event was found", async () => {
      jest.spyOn(lineaRollupMock, "inboxL2L1MessageStatus").mockResolvedValue(0n);
      jest.spyOn(lineaRollupLogClientMock, "getMessageClaimedEvents").mockResolvedValue([testMessageClaimedEvent]);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatusUsingMessageHash(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMED);
    });
  });

  describe("getMessageStatus", () => {
    it("should return UNKNOWN when l2MessagingBlockAnchoredEvent is absent and isMeessageClaimed return false", async () => {
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents").mockResolvedValue([]);
      jest.spyOn(lineaRollupMock, "isMessageClaimed").mockResolvedValue(false);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatus(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.UNKNOWN);
    });

    it("should return CLAIMABLE when l2MessagingBlockAnchoredEvent is present and isMessageClaimed return false", async () => {
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(lineaRollupMock, "isMessageClaimed").mockResolvedValue(false);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatus(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMABLE);
    });

    it("should return CLAIMED when isMessageClaimed return true", async () => {
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents").mockResolvedValue([]);
      jest.spyOn(lineaRollupMock, "isMessageClaimed").mockResolvedValue(true);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatus(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMED);
    });
  });

  describe("getMessageStatusUsingMerkleTree", () => {
    it("should throw error when the corresponding message sent event was not found on L2", async () => {
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash").mockResolvedValue([]);

      await expect(lineaRolliupClient.getMessageStatusUsingMerkleTree(TEST_MESSAGE_HASH)).rejects.toThrow(
        new BaseError(`Message hash does not exist on L2. Message hash: ${TEST_MESSAGE_HASH}`),
      );
    });

    it("should return UNKNOWN when l2MessagingBlockAnchoredEvent is absent and isMeessageClaimed return false", async () => {
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents").mockResolvedValue([]);
      jest.spyOn(lineaRollupMock, "isMessageClaimed").mockResolvedValue(false);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatusUsingMerkleTree(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.UNKNOWN);
    });

    it("should return CLAIMABLE when l2MessagingBlockAnchoredEvent is present and isMessageClaimed return false", async () => {
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(lineaRollupMock, "isMessageClaimed").mockResolvedValue(false);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatusUsingMerkleTree(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMABLE);
    });

    it("should return CLAIMED when isMessageClaimed return true", async () => {
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents").mockResolvedValue([]);
      jest.spyOn(lineaRollupMock, "isMessageClaimed").mockResolvedValue(true);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await lineaRolliupClient.getMessageStatusUsingMerkleTree(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMED);
    });
  });

  describe("estimateClaimWithoutProofGas", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(lineaRolliupClient.estimateClaimWithoutProofGas(message)).rejects.toThrow(
        new BaseError("'EstimateClaimGas' function not callable using readOnly mode."),
      );
    });

    it("should throw a GasEstimationError when the gas estimation failed", async () => {
      const message = generateMessage();
      mockProperty(lineaRollupMock, "claimMessage", {
        estimateGas: jest.fn().mockRejectedValueOnce(new Error("Gas estimation failed").message),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        undefined,
        undefined,
        true,
      );

      await expect(lineaRolliupClient.estimateClaimWithoutProofGas(message)).rejects.toThrow(
        new GasEstimationError("Gas estimation failed", message),
      );
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      const message = generateMessage();
      const estimatedGasLimit = 50_000n;
      mockProperty(lineaRollupMock, "claimMessage", {
        estimateGas: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        1000000000n,
        undefined,
        true,
      );
      const claimMessageSpy = jest.spyOn(lineaRollupMock.claimMessage, "estimateGas");

      const estimateClaimGasReturned = await lineaRolliupClient.estimateClaimWithoutProofGas(message);

      expect(estimateClaimGasReturned).toStrictEqual(estimatedGasLimit);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        ZERO_ADDRESS,
        message.calldata,
        message.messageNonce,
        {
          maxFeePerGas: 1000000000n,
          maxPriorityFeePerGas: 1000000000n,
        },
      );
    });

    it("should return estimated gas limit for the claim message transaction", async () => {
      const message = generateMessage();
      const estimatedGasLimit = 50_000n;
      mockProperty(lineaRollupMock, "claimMessage", {
        estimateGas: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        1000000000n,
        undefined,
        true,
      );
      const claimMessageSpy = jest.spyOn(lineaRollupMock.claimMessage, "estimateGas");

      const estimateClaimGasReturned = await lineaRolliupClient.estimateClaimWithoutProofGas({
        ...message,
        feeRecipient: TEST_ADDRESS_2,
      });

      expect(estimateClaimGasReturned).toStrictEqual(estimatedGasLimit);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        TEST_ADDRESS_2,
        message.calldata,
        message.messageNonce,
        {
          maxFeePerGas: 1000000000n,
          maxPriorityFeePerGas: 1000000000n,
        },
      );
    });
  });

  describe("estimateClaimWithProofGas", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(
        lineaRolliupClient.estimateClaimWithProofGas({
          ...message,
          leafIndex: 0,
          merkleRoot: TEST_MERKLE_ROOT,
          proof: [
            "0x0000000000000000000000000000000000000000000000000000000000000000",
            "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
            "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
            "0x21ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85",
            "0xe58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a19344",
          ],
        }),
      ).rejects.toThrow("'EstimateClaimWithProofGas' function not callable using readOnly mode.");
    });

    it("should throw GasEstimationError if estimateGas throws error", async () => {
      const message = generateMessage();
      mockProperty(lineaRollupMock, "claimMessageWithProof", {
        estimateGas: jest.fn().mockRejectedValueOnce(new Error("Failed to estimate gas")),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        1000000000n,
        undefined,
        true,
      );
      await expect(
        lineaRolliupClient.estimateClaimWithProofGas({
          ...message,
          leafIndex: 0,
          merkleRoot: TEST_MERKLE_ROOT,
          proof: [
            "0x0000000000000000000000000000000000000000000000000000000000000000",
            "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
            "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
            "0x21ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85",
            "0xe58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a19344",
          ],
        }),
      ).rejects.toThrow("Failed to estimate gas");
    });
  });

  describe("estimateClaimGas", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(lineaRolliupClient.estimateClaimGas(message)).rejects.toThrow(
        "'EstimateClaimGas' function not callable using readOnly mode.",
      );
    });

    it("should throw a GasEstimationError when the message hash does not exist on L2", async () => {
      const message = generateMessage();
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash").mockResolvedValue([]);

      await expect(lineaRolliupClient.estimateClaimGas(message)).rejects.toThrow(
        `Message hash does not exist on L2. Message hash: ${TEST_MESSAGE_HASH}`,
      );
    });

    it("should throw a GasEstimationError when the L2 block number has not been finalized on L1", async () => {
      const message = generateMessage();
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents").mockResolvedValue([]);

      await expect(lineaRolliupClient.estimateClaimGas(message)).rejects.toThrow(
        "L2 block number 51 has not been finalized on L1",
      );
    });

    it("should throw a GasEstimationError when finalization transaction does not exist", async () => {
      const message = generateMessage();
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(null);

      await expect(lineaRolliupClient.estimateClaimGas(message)).rejects.toThrow(
        `Transaction does not exist or no logs found in this transaction: ${TEST_TRANSACTION_HASH}.`,
      );
    });

    it("should throw a GasEstimationError when no related event logs were found", async () => {
      const message = generateMessage();
      const transactionReceipt = generateTransactionReceipt();
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);

      await expect(lineaRolliupClient.estimateClaimGas(message)).rejects.toThrow(
        "No L2MerkleRootAdded events found in this transaction.",
      );
    });

    it("should throw a GasEstimationError when no L2MessagingBlocksAnchored event logs were found", async () => {
      const message = generateMessage();
      const transactionReceipt = generateTransactionReceiptWithLogs(undefined, [
        generateL2MerkleTreeAddedLog(TEST_TRANSACTION_HASH, 5),
      ]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);

      await expect(lineaRolliupClient.estimateClaimGas(message)).rejects.toThrow(
        "No L2MessagingBlocksAnchored events found in this transaction.",
      );
    });

    it("should throw a GasEstimationError when no MessageSent events found in the given block range on L2", async () => {
      const message = generateMessage();
      const transactionReceipt = generateTransactionReceiptWithLogs(undefined, [
        generateL2MerkleTreeAddedLog(TEST_TRANSACTION_HASH, 5),
        generateL2MessagingBlockAnchoredLog(10n),
      ]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByBlockRange").mockResolvedValue([]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);

      await expect(lineaRolliupClient.estimateClaimGas(message)).rejects.toThrowError();
    });

    it("should return estimated gas limit if all the relevant event logs were found", async () => {
      const message = generateMessage();
      const transactionReceipt = generateTransactionReceiptWithLogs(undefined, [
        generateL2MerkleTreeAddedLog(TEST_MERKLE_ROOT, 5),
        generateL2MessagingBlockAnchoredLog(10n),
      ]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByBlockRange")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);

      const estimatedGasLimit = 50_000n;
      mockProperty(lineaRollupMock, "claimMessageWithProof", {
        estimateGas: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      mockProperty(lineaRollupMock, "interface", {
        parseLog: jest
          .fn()
          .mockReturnValueOnce({
            args: { treeDepth: 5, l2MerkleRoot: TEST_MERKLE_ROOT },
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any)
          .mockReturnValueOnce({
            args: { l2Block: 10n },
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        1000000000n,
        undefined,
        true,
      );
      const claimMessageWithProofSpy = jest.spyOn(lineaRollupMock.claimMessageWithProof, "estimateGas");

      const estimatedClaimGas = await lineaRolliupClient.estimateClaimGas(message);

      expect(estimatedClaimGas).toStrictEqual(estimatedGasLimit);
      expect(claimMessageWithProofSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageWithProofSpy).toHaveBeenCalledWith(
        {
          from: message.messageSender,
          to: message.destination,
          fee: message.fee,
          value: message.value,
          feeRecipient: ZERO_ADDRESS,
          data: message.calldata,
          messageNumber: message.messageNonce,
          leafIndex: 0,
          merkleRoot: TEST_MERKLE_ROOT,
          proof: [
            "0x0000000000000000000000000000000000000000000000000000000000000000",
            "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
            "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
            "0x21ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85",
            "0xe58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a19344",
          ],
        },
        {
          maxFeePerGas: 1000000000n,
          maxPriorityFeePerGas: 1000000000n,
        },
      );
    });
  });

  describe("claimWithoutProof", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(lineaRolliupClient.claimWithoutProof(message)).rejects.toThrow(
        new Error("'claim' function not callable using readOnly mode."),
      );
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      const message = generateMessage();
      const txResponse = generateTransactionResponse();
      jest.spyOn(lineaRollupMock, "claimMessage").mockResolvedValue(txResponse as ContractTransactionResponse);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        500000000n,
        undefined,
        true,
      );
      const claimMessageSpy = jest.spyOn(lineaRollupMock, "claimMessage");

      await lineaRolliupClient.claimWithoutProof(message);

      expect(txResponse).toStrictEqual(txResponse);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        ZERO_ADDRESS,
        message.calldata,
        message.messageNonce,
        {
          maxPriorityFeePerGas: 500000000n,
          maxFeePerGas: 500000000n,
        },
      );
    });

    it("should return executed claim message transaction", async () => {
      const message = generateMessage();
      const txResponse = generateTransactionResponse();
      jest.spyOn(lineaRollupMock, "claimMessage").mockResolvedValue(txResponse as ContractTransactionResponse);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        500000000n,
        undefined,
        true,
      );
      const claimMessageSpy = jest.spyOn(lineaRollupMock, "claimMessage");

      const txResponseReturned = await lineaRolliupClient.claimWithoutProof({
        ...message,
        feeRecipient: TEST_ADDRESS_2,
      });

      expect(txResponseReturned).toStrictEqual(txResponse);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        TEST_ADDRESS_2,
        message.calldata,
        message.messageNonce,
        {
          maxPriorityFeePerGas: 500000000n,
          maxFeePerGas: 500000000n,
        },
      );
    });
  });

  describe("claimWithProof", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(
        lineaRolliupClient.claimWithProof({
          ...message,
          leafIndex: 0,
          merkleRoot: TEST_MERKLE_ROOT,
          proof: [
            "0x0000000000000000000000000000000000000000000000000000000000000000",
            "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
            "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
            "0x21ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85",
            "0xe58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a19344",
          ],
        }),
      ).rejects.toThrow(new Error("'claimWithProof' function not callable using readOnly mode."));
    });
  });

  describe("claim", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(lineaRolliupClient.claim(message)).rejects.toThrow(
        new Error("'claim' function not callable using readOnly mode."),
      );
    });

    it("should return executed claim message transaction", async () => {
      const message = generateMessage();
      const txResponse = generateTransactionResponse();
      const transactionReceipt = generateTransactionReceiptWithLogs(undefined, [
        generateL2MerkleTreeAddedLog(TEST_MERKLE_ROOT, 5),
        generateL2MessagingBlockAnchoredLog(10n),
      ]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByBlockRange")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);

      mockProperty(lineaRollupMock, "interface", {
        parseLog: jest
          .fn()
          .mockReturnValueOnce({
            args: { treeDepth: 5, l2MerkleRoot: TEST_MERKLE_ROOT },
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any)
          .mockReturnValueOnce({
            args: { l2Block: 10n },
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(lineaRollupMock, "claimMessageWithProof").mockResolvedValue(txResponse as ContractTransactionResponse);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        1000000000n,
        undefined,
        true,
      );
      const claimMessageWithProofSpy = jest.spyOn(lineaRollupMock, "claimMessageWithProof");

      const txResponseReturned = await lineaRolliupClient.claim(message);

      expect(txResponseReturned).toStrictEqual(txResponse);
      expect(claimMessageWithProofSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageWithProofSpy).toHaveBeenCalledWith(
        {
          from: message.messageSender,
          to: message.destination,
          fee: message.fee,
          value: message.value,
          feeRecipient: ZERO_ADDRESS,
          data: message.calldata,
          messageNumber: message.messageNonce,
          leafIndex: 0,
          merkleRoot: TEST_MERKLE_ROOT,
          proof: [
            "0x0000000000000000000000000000000000000000000000000000000000000000",
            "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
            "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
            "0x21ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85",
            "0xe58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a19344",
          ],
        },
        {
          maxFeePerGas: 1000000000n,
          maxPriorityFeePerGas: 1000000000n,
        },
      );
    });
  });

  describe("getMessageSiblings", () => {
    it("should throw a BaseError when message hash not found in messages", () => {
      const messageHash = TEST_MESSAGE_HASH;
      const messageHashes = [TEST_MESSAGE_HASH_2];

      expect(() => lineaRolliupClient.getMessageSiblings(messageHash, messageHashes, 5)).toThrow(
        "Message hash not found in messages.",
      );
    });
  });

  describe("getMessageProof", () => {
    it("should throw a BaseError if merkle tree build failed", async () => {
      const messageHash = TEST_MESSAGE_HASH;
      const transactionReceipt = generateTransactionReceiptWithLogs(undefined, [
        generateL2MerkleTreeAddedLog(TEST_MERKLE_ROOT, 5),
        generateL2MessagingBlockAnchoredLog(10n),
      ]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByMessageHash")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(l2MessageServiceLogClientMock, "getMessageSentEventsByBlockRange")
        .mockResolvedValue([testMessageSentEvent]);
      jest
        .spyOn(lineaRollupLogClientMock, "getL2MessagingBlockAnchoredEvents")
        .mockResolvedValue([testL2MessagingBlockAnchoredEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);
      mockProperty(lineaRollupMock, "interface", {
        parseLog: jest
          .fn()
          .mockReturnValueOnce({
            args: { treeDepth: 5, l2MerkleRoot: TEST_MERKLE_ROOT_2 },
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any)
          .mockReturnValueOnce({
            args: { l2Block: 10n },
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
          } as any),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        1000000000n,
        undefined,
        true,
      );

      await expect(lineaRolliupClient.getMessageProof(messageHash)).rejects.toThrow("Merkle tree build failed.");
    });
  });

  describe("retryTransactionWithHigherFee", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );

      await expect(lineaRolliupClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new BaseError("'retryTransactionWithHigherFee' function not callable using readOnly mode."),
      );
    });

    it("should throw an error when priceBumpPercent is not an integer", async () => {
      await expect(lineaRolliupClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH, 1.1)).rejects.toThrow(
        new BaseError("'priceBumpPercent' must be an integer"),
      );
    });

    it("should throw an error when getTransaction return null", async () => {
      jest.spyOn(providerMock, "getTransaction").mockResolvedValue(null);

      await expect(lineaRolliupClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new BaseError(`Transaction with hash ${TEST_TRANSACTION_HASH} not found.`),
      );
    });

    it("should retry the transaction with higher fees", async () => {
      const transactionResponse = generateTransactionResponse();
      const getTransactionSpy = jest.spyOn(providerMock, "getTransaction").mockResolvedValue(transactionResponse);
      const signTransactionSpy = jest.spyOn(walletMock, "signTransaction").mockResolvedValue("");
      const sendTransactionSpy = jest.spyOn(providerMock, "broadcastTransaction");

      await lineaRolliupClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH);

      expect(getTransactionSpy).toHaveBeenCalledTimes(1);
      expect(signTransactionSpy).toHaveBeenCalledTimes(1);
      expect(signTransactionSpy).toHaveBeenCalledWith({
        to: transactionResponse.to,
        value: transactionResponse.value,
        data: transactionResponse.data,
        nonce: transactionResponse.nonce,
        gasLimit: transactionResponse.gasLimit,
        chainId: transactionResponse.chainId,
        type: 2,
        maxPriorityFeePerGas: 55000000n,
        maxFeePerGas: 110000000n,
      });
      expect(sendTransactionSpy).toHaveBeenCalledTimes(1);
    });

    it("should retry the transaction with higher fees and capped by the predefined maxFeePerGas", async () => {
      const transactionResponse = generateTransactionResponse();
      const getTransactionSpy = jest.spyOn(providerMock, "getTransaction").mockResolvedValue(transactionResponse);
      const signTransactionSpy = jest.spyOn(walletMock, "signTransaction").mockResolvedValue("");
      const sendTransactionSpy = jest.spyOn(providerMock, "broadcastTransaction");
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        500000000n,
      );

      await lineaRolliupClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH, 1000);

      expect(getTransactionSpy).toHaveBeenCalledTimes(1);
      expect(signTransactionSpy).toHaveBeenCalledTimes(1);
      expect(signTransactionSpy).toHaveBeenCalledWith({
        to: transactionResponse.to,
        value: transactionResponse.value,
        data: transactionResponse.data,
        nonce: transactionResponse.nonce,
        gasLimit: transactionResponse.gasLimit,
        chainId: transactionResponse.chainId,
        type: 2,
        maxPriorityFeePerGas: 500000000n,
        maxFeePerGas: 500000000n,
      });
      expect(sendTransactionSpy).toHaveBeenCalledTimes(1);
    });

    it("should retry the transaction with the predefined maxFeePerGas if enforceMaxGasFee is true", async () => {
      const transactionResponse = generateTransactionResponse({
        maxPriorityFeePerGas: undefined,
        maxFeePerGas: undefined,
      });
      const getTransactionSpy = jest.spyOn(providerMock, "getTransaction").mockResolvedValue(transactionResponse);
      const signTransactionSpy = jest.spyOn(walletMock, "signTransaction").mockResolvedValue("");
      const sendTransactionSpy = jest.spyOn(providerMock, "broadcastTransaction");
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        500000000n,
        undefined,
        true,
      );

      await lineaRolliupClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH, 1000);

      expect(getTransactionSpy).toHaveBeenCalledTimes(1);
      expect(signTransactionSpy).toHaveBeenCalledTimes(1);
      expect(signTransactionSpy).toHaveBeenCalledWith({
        to: transactionResponse.to,
        value: transactionResponse.value,
        data: transactionResponse.data,
        nonce: transactionResponse.nonce,
        gasLimit: transactionResponse.gasLimit,
        chainId: transactionResponse.chainId,
        type: 2,
        maxPriorityFeePerGas: 500000000n,
        maxFeePerGas: 500000000n,
      });
      expect(sendTransactionSpy).toHaveBeenCalledTimes(1);
    });
  });

  describe("isRateLimitExceeded", () => {
    it("should always return false", async () => {
      const isRateLimitExceeded = await lineaRolliupClient.isRateLimitExceeded(1000000000n, 1000000000n);

      expect(isRateLimitExceeded).toBeFalsy;
    });
  });

  describe("isRateLimitExceededError", () => {
    it("should return false when something went wrong (http error etc)", async () => {
      jest.spyOn(providerMock, "getTransaction").mockRejectedValueOnce({});
      expect(
        await lineaRolliupClient.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(false);
    });

    it("should return false when transaction revert reason is not RateLimitExceeded", async () => {
      jest.spyOn(providerMock, "getTransaction").mockResolvedValueOnce(generateTransactionResponse());
      jest.spyOn(providerMock, "call").mockResolvedValueOnce("0xa74c1c6d");

      expect(
        await lineaRolliupClient.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(false);
    });

    it("should return true when transaction revert reason is RateLimitExceeded", async () => {
      mockProperty(lineaRollupMock, "interface", {
        ...lineaRollupMock.interface,
        parseError: jest.fn().mockReturnValueOnce({ name: "RateLimitExceeded" }),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(providerMock, "getTransaction").mockResolvedValueOnce(generateTransactionResponse());
      jest.spyOn(providerMock, "call").mockResolvedValueOnce("0xa74c1c5f");
      jest.spyOn(LineaRollup__factory, "connect").mockReturnValueOnce(lineaRollupMock);
      lineaRolliupClient = new LineaRollupClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        lineaRollupLogClientMock,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      expect(
        await lineaRolliupClient.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(true);
    });
  });
});
