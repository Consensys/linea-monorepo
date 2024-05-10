import { describe, afterEach, it, expect, beforeEach } from "@jest/globals";
import { MockProxy, mock } from "jest-mock-extended";
import { ContractTransactionResponse, JsonRpcProvider, Wallet } from "ethers";
import {
  testMessageSentEvent,
  TEST_MESSAGE_HASH,
  TEST_CONTRACT_ADDRESS_1,
  TEST_TRANSACTION_HASH,
  TEST_ADDRESS_2,
} from "../../../../utils/testing/constants";
import { L2MessageService, L2MessageService__factory } from "../../typechain";
import {
  generateMessage,
  generateTransactionReceipt,
  generateTransactionResponse,
  mockProperty,
} from "../../../../utils/testing/helpers";
import { IL2MessageServiceLogClient } from "../../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";
import { L2MessageServiceClient } from "../L2MessageServiceClient";
import { ZERO_ADDRESS } from "../../../../core/constants";
import { OnChainMessageStatus } from "../../../../core/enums/MessageEnums";
import { GasEstimationError } from "../../../../core/errors/GasFeeErrors";
import { BaseError } from "../../../../core/errors/Base";

describe("TestL2MessageServiceClient", () => {
  let providerMock: MockProxy<JsonRpcProvider>;
  let walletMock: MockProxy<Wallet>;
  let l2MessageServiceMock: MockProxy<L2MessageService>;
  let l2MessageServiceLogClientMock: MockProxy<IL2MessageServiceLogClient>;
  let l2MessageServiceClient: L2MessageServiceClient;

  beforeEach(() => {
    providerMock = mock<JsonRpcProvider>();
    walletMock = mock<Wallet>();
    l2MessageServiceLogClientMock = mock<IL2MessageServiceLogClient>();
    l2MessageServiceMock = mock<L2MessageService>();
    l2MessageServiceClient = new L2MessageServiceClient(
      providerMock,
      TEST_CONTRACT_ADDRESS_1,
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
          new L2MessageServiceClient(
            providerMock,
            TEST_CONTRACT_ADDRESS_1,
            l2MessageServiceLogClientMock,
            "read-write",
          ),
      ).toThrowError(new BaseError("Please provide a signer."));
    });
  });

  describe("getMessageByMessageHash", () => {
    it("should return a MessageSent", async () => {
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);

      const messageSentEvent = await l2MessageServiceClient.getMessageByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentEvent).toStrictEqual(testMessageSentEvent);
    });

    it("should return null if empty events returned", async () => {
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEvents").mockResolvedValue([]);

      const messageSentEvent = await l2MessageServiceClient.getMessageByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentEvent).toStrictEqual(null);
    });
  });

  describe("getMessagesByTransactionHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(null);

      const messageSentEvents = await l2MessageServiceClient.getMessagesByTransactionHash(TEST_TRANSACTION_HASH);

      expect(messageSentEvents).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);

      const messageSentEvents = await l2MessageServiceClient.getMessagesByTransactionHash(TEST_MESSAGE_HASH);

      expect(messageSentEvents).toStrictEqual([testMessageSentEvent]);
    });
  });

  describe("getTransactionReceiptByMessageHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEvents").mockResolvedValue([]);

      const messageSentTxReceipt = await l2MessageServiceClient.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentTxReceipt).toStrictEqual(null);
    });

    it("should return null when transaction receipt does not exist", async () => {
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(null);

      const messageSentTxReceipt = await l2MessageServiceClient.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentTxReceipt).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(l2MessageServiceLogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValue(transactionReceipt);

      const messageSentTxReceipt = await l2MessageServiceClient.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH);

      expect(messageSentTxReceipt).toStrictEqual(transactionReceipt);
    });
  });

  describe("getMessageStatus", () => {
    it("should return UNKNOWN when on chain message status === 0", async () => {
      jest.spyOn(l2MessageServiceMock, "inboxL1L2MessageStatus").mockResolvedValue(0n);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await l2MessageServiceClient.getMessageStatus(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.UNKNOWN);
    });

    it("should return CLAIMABLE when on chain message status === 1", async () => {
      jest.spyOn(l2MessageServiceMock, "inboxL1L2MessageStatus").mockResolvedValue(1n);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await l2MessageServiceClient.getMessageStatus(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMABLE);
    });

    it("should return CLAIMED when on chain message status === 2", async () => {
      jest.spyOn(l2MessageServiceMock, "inboxL1L2MessageStatus").mockResolvedValue(2n);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const messageStatus = await l2MessageServiceClient.getMessageStatus(TEST_MESSAGE_HASH);

      expect(messageStatus).toStrictEqual(OnChainMessageStatus.CLAIMED);
    });
  });

  describe("estimateClaimGas", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(l2MessageServiceClient.estimateClaimGas(message)).rejects.toThrow(
        new Error("'EstimateClaimGas' function not callable using readOnly mode."),
      );
    });

    it("should throw a GasEstimationError when the gas estimation failed", async () => {
      const message = generateMessage();
      mockProperty(l2MessageServiceMock, "claimMessage", {
        estimateGas: jest.fn().mockRejectedValueOnce(new Error("Gas estimation failed").message),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      await expect(l2MessageServiceClient.estimateClaimGas(message)).rejects.toThrow(
        new GasEstimationError("Gas estimation failed", message),
      );
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      const message = generateMessage();
      const estimatedGasLimit = 50_000n;
      mockProperty(l2MessageServiceMock, "claimMessage", {
        estimateGas: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );
      const claimMessageSpy = jest.spyOn(l2MessageServiceMock.claimMessage, "estimateGas");

      const estimateClaimGasReturned = await l2MessageServiceClient.estimateClaimGas(message);

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
        {},
      );
    });

    it("should return estimated gas limit for the claim message transaction", async () => {
      const message = generateMessage();
      const estimatedGasLimit = 50_000n;
      mockProperty(l2MessageServiceMock, "claimMessage", {
        estimateGas: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );
      const claimMessageSpy = jest.spyOn(l2MessageServiceMock.claimMessage, "estimateGas");

      const estimateClaimGasReturned = await l2MessageServiceClient.estimateClaimGas({
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
        {},
      );
    });
  });

  describe("claim", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );
      const message = generateMessage();
      await expect(l2MessageServiceClient.claim(message)).rejects.toThrow(
        new Error("'claim' function not callable using readOnly mode."),
      );
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      const message = generateMessage();
      const txResponse = generateTransactionResponse();
      jest.spyOn(l2MessageServiceMock, "claimMessage").mockResolvedValue(txResponse as ContractTransactionResponse);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );
      const claimMessageSpy = jest.spyOn(l2MessageServiceMock, "claimMessage");

      await l2MessageServiceClient.claim(message);

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
        {},
      );
    });

    it("should return executed claim message transaction", async () => {
      const message = generateMessage();
      const txResponse = generateTransactionResponse();
      jest.spyOn(l2MessageServiceMock, "claimMessage").mockResolvedValue(txResponse as ContractTransactionResponse);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );
      const claimMessageSpy = jest.spyOn(l2MessageServiceMock, "claimMessage");

      await l2MessageServiceClient.claim({
        ...message,
        feeRecipient: TEST_ADDRESS_2,
      });

      expect(txResponse).toStrictEqual(txResponse);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        TEST_ADDRESS_2,
        message.calldata,
        message.messageNonce,
        {},
      );
    });
  });

  describe("retryTransactionWithHigherFee", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-only",
        walletMock,
      );

      await expect(l2MessageServiceClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new BaseError("'retryTransactionWithHigherFee' function not callable using readOnly mode."),
      );
    });

    it("should throw an error when priceBumpPercent is not an integer", async () => {
      await expect(l2MessageServiceClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH, 1.1)).rejects.toThrow(
        new BaseError("'priceBumpPercent' must be an integer"),
      );
    });

    it("should throw an error when getTransaction return null", async () => {
      jest.spyOn(providerMock, "getTransaction").mockResolvedValue(null);

      await expect(l2MessageServiceClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new BaseError(`Transaction with hash ${TEST_TRANSACTION_HASH} not found.`),
      );
    });

    it("should retry the transaction with higher fees", async () => {
      const transactionResponse = generateTransactionResponse();
      const getTransactionSpy = jest.spyOn(providerMock, "getTransaction").mockResolvedValue(transactionResponse);
      const signTransactionSpy = jest.spyOn(walletMock, "signTransaction").mockResolvedValue("");
      const sendTransactionSpy = jest.spyOn(providerMock, "broadcastTransaction");

      await l2MessageServiceClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH);

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
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        500000000n,
      );

      await l2MessageServiceClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH, 1000);

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
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
        500000000n,
        undefined,
        true,
      );

      await l2MessageServiceClient.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH, 1000);

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
    it("should return true if exceeded rate limit", async () => {
      jest.spyOn(l2MessageServiceMock, "limitInWei").mockResolvedValue(2000000000n);
      jest.spyOn(l2MessageServiceMock, "currentPeriodAmountInWei").mockResolvedValue(1000000000n);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const isRateLimitExceeded = await l2MessageServiceClient.isRateLimitExceeded(1000000000n, 1000000000n);

      expect(isRateLimitExceeded).toBeTruthy;
    });

    it("should return false if not exceeded rate limit", async () => {
      jest.spyOn(l2MessageServiceMock, "limitInWei").mockResolvedValue(2000000000n);
      jest.spyOn(l2MessageServiceMock, "currentPeriodAmountInWei").mockResolvedValue(1000000000n);
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      const isRateLimitExceeded = await l2MessageServiceClient.isRateLimitExceeded(100000000n, 100000000n);

      expect(isRateLimitExceeded).toBeFalsy;
    });
  });

  describe("isRateLimitExceededError", () => {
    it("should return false when something went wrong (http error etc)", async () => {
      jest.spyOn(providerMock, "getTransaction").mockRejectedValueOnce({});
      expect(
        await l2MessageServiceClient.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(false);
    });

    it("should return false when transaction revert reason is not RateLimitExceeded", async () => {
      jest.spyOn(providerMock, "getTransaction").mockResolvedValueOnce(generateTransactionResponse());
      jest.spyOn(providerMock, "call").mockResolvedValueOnce("0xa74c1c6d");

      expect(
        await l2MessageServiceClient.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(false);
    });

    it("should return true when transaction revert reason is RateLimitExceeded", async () => {
      mockProperty(l2MessageServiceMock, "interface", {
        ...l2MessageServiceMock.interface,
        parseError: jest.fn().mockReturnValueOnce({ name: "RateLimitExceeded" }),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any);
      jest.spyOn(providerMock, "getTransaction").mockResolvedValueOnce(generateTransactionResponse());
      jest.spyOn(providerMock, "call").mockResolvedValueOnce("0xa74c1c5f");
      jest.spyOn(L2MessageService__factory, "connect").mockReturnValueOnce(l2MessageServiceMock);
      l2MessageServiceClient = new L2MessageServiceClient(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        l2MessageServiceLogClientMock,
        "read-write",
        walletMock,
      );

      expect(
        await l2MessageServiceClient.isRateLimitExceededError(
          "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
        ),
      ).toStrictEqual(true);
    });
  });
});
