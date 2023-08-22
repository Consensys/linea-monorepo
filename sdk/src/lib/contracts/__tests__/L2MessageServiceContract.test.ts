import { JsonRpcProvider } from "@ethersproject/providers";
import { describe, afterEach, it, expect, beforeEach } from "@jest/globals";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { BigNumber, ContractFunction, Signer, ethers } from "ethers";
import { L2MessageServiceContract } from "../L2MessageServiceContract";
import { getTestL2Signer } from "../../utils/testHelpers/contracts";
import { L2MessageService, L2MessageService__factory } from "../../../typechain";
import {
  TEST_ADDRESS_1,
  TEST_ADDRESS_2,
  TEST_CONTRACT_ADDRESS_1,
  TEST_MESSAGE_HASH,
  TEST_TRANSACTION_HASH,
  testMessageSentEvent,
} from "../../utils/testHelpers/constants";
import { mapMessageSentEventOrLogToMessage } from "../../utils/mappers";
import { EventParser } from "../EventParser";
import {
  generateMessage,
  generateTransactionReceipt,
  generateTransactionResponse,
  mockProperty,
  undoMockProperty,
} from "../../utils/testHelpers/helpers";
import { GasEstimationError } from "../../utils/errors";
import { OnChainMessageStatus } from "../../utils/enum";

describe("L2MessageServiceContract", () => {
  let providerMock: MockProxy<JsonRpcProvider>;
  let l2MessageServiceContract: L2MessageServiceContract;

  beforeEach(() => {
    providerMock = mock<JsonRpcProvider>();
    l2MessageServiceContract = new L2MessageServiceContract(
      providerMock,
      TEST_CONTRACT_ADDRESS_1,
      "read-write",
      getTestL2Signer(),
    );
  });

  afterEach(() => {
    mockClear(providerMock);
  });

  describe("getContractAbi", () => {
    it("should return L2MessageService contract abi", async () => {
      expect(l2MessageServiceContract.getContractAbi()).toStrictEqual(L2MessageService__factory.abi);
    });
  });

  describe("getMessageByMessageHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([]);
      expect(await l2MessageServiceContract.getMessageByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(null);
    });

    it("should return message when message hash exists", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      expect(await l2MessageServiceContract.getMessageByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(
        mapMessageSentEventOrLogToMessage(testMessageSentEvent),
      );
    });
  });

  describe("getMessagesByTransactionHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest
        .spyOn(l2MessageServiceContract.provider, "getTransactionReceipt")
        .mockImplementationOnce(jest.fn().mockResolvedValueOnce(null));
      expect(await l2MessageServiceContract.getMessagesByTransactionHash(TEST_TRANSACTION_HASH)).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(l2MessageServiceContract.provider, "getTransactionReceipt").mockResolvedValueOnce(transactionReceipt);
      expect(await l2MessageServiceContract.getMessagesByTransactionHash(TEST_MESSAGE_HASH)).toStrictEqual([
        mapMessageSentEventOrLogToMessage(
          L2MessageService__factory.createInterface().parseLog(transactionReceipt.logs[0]),
        ),
      ]);
    });
  });

  describe("getTransactionReceiptByMessageHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([]);
      expect(await l2MessageServiceContract.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(null);
    });

    it("should return null when transaction receipt does not exist", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      jest
        .spyOn(l2MessageServiceContract.provider, "getTransactionReceipt")
        .mockImplementationOnce(jest.fn().mockResolvedValueOnce(null));
      expect(await l2MessageServiceContract.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      jest.spyOn(l2MessageServiceContract.provider, "getTransactionReceipt").mockResolvedValueOnce(transactionReceipt);
      expect(await l2MessageServiceContract.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(
        transactionReceipt,
      );
    });
  });

  describe("getContract", () => {
    it("should throw an error when mode = 'read-write' and signer params is undefined", () => {
      try {
        l2MessageServiceContract.getContract(TEST_CONTRACT_ADDRESS_1);
      } catch (e) {
        expect(e).toEqual(new Error("Please provide a signer."));
      }
    });

    it("should return L2MessageService contract instance with signer when mode = 'read-write'", () => {
      const signer = getTestL2Signer();
      const l2MessageServiceInstance = l2MessageServiceContract.getContract(TEST_CONTRACT_ADDRESS_1, signer);
      expect(JSON.stringify(l2MessageServiceInstance)).toStrictEqual(
        JSON.stringify(L2MessageService__factory.connect(TEST_CONTRACT_ADDRESS_1, signer)),
      );
    });

    it("should return L2MessageService contract instance with provider when mode = 'read-only'", () => {
      const messageServiceContract = new L2MessageServiceContract(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        "read-only",
        getTestL2Signer(),
      );
      const l2MessageServiceInstance = messageServiceContract.getContract(TEST_CONTRACT_ADDRESS_1);
      expect(JSON.stringify(l2MessageServiceInstance)).toStrictEqual(
        JSON.stringify(L2MessageService__factory.connect(TEST_CONTRACT_ADDRESS_1, messageServiceContract.provider)),
      );
    });
  });

  describe("getCurrentNonce", () => {
    it("should throw an error when mode = 'read-write' and this.signer is undefined", async () => {
      jest
        .spyOn(L2MessageServiceContract.prototype, "getContract")
        .mockReturnValueOnce(L2MessageService__factory.connect(TEST_CONTRACT_ADDRESS_1, getTestL2Signer()));

      const messageServiceContract = new L2MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-write");

      await expect(messageServiceContract.getCurrentNonce()).rejects.toThrow(new Error("Please provide a signer."));
    });

    it("should throw an error when mode = 'read-only' and accountAddress param is undefined", async () => {
      const messageServiceContract = new L2MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      await expect(messageServiceContract.getCurrentNonce()).rejects.toThrow(
        new Error("Please provider an account address."),
      );
    });

    it("should return account nonce for the account address passed in params", async () => {
      const messageServiceContract = new L2MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      jest.spyOn(messageServiceContract.provider, "getTransactionCount").mockResolvedValueOnce(10);
      expect(await messageServiceContract.getCurrentNonce(TEST_ADDRESS_1)).toStrictEqual(10);
    });

    it("should return account nonce for this.signer address when mode = 'read-write'", async () => {
      const getTransactionCountSpy = jest
        .spyOn(l2MessageServiceContract.provider, "getTransactionCount")
        .mockResolvedValueOnce(10);
      expect(await l2MessageServiceContract.getCurrentNonce()).toStrictEqual(10);

      expect(getTransactionCountSpy).toHaveBeenCalledTimes(1);
      expect(getTransactionCountSpy).toHaveBeenCalledWith(await getTestL2Signer().getAddress());
    });
  });

  describe("getCurrentBlockNumber", () => {
    it("should throw an error when mode = 'read-write' and this.signer is undefined", async () => {
      const currentBlockNumber = 50_000;
      jest.spyOn(l2MessageServiceContract.provider, "getBlockNumber").mockResolvedValueOnce(currentBlockNumber);
      expect(await l2MessageServiceContract.getCurrentBlockNumber()).toStrictEqual(currentBlockNumber);
    });
  });

  describe("getEvents", () => {
    it("should return empty array when there are no events", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([]);
      const messageSentEventFilter = l2MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 0;
      const toBlock = "latest";
      expect(await l2MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock)).toStrictEqual([]);
    });

    it("should return empty array when there are events but event.blockNumber === fromBlock && event.logIndex < fromBlockLogIndex", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      const messageSentEventFilter = l2MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 51;
      const toBlock = "latest";
      const fromBlockLogIndex = 2;
      expect(
        await l2MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock, fromBlockLogIndex),
      ).toStrictEqual([]);
    });

    it("should return filtered and parsed events when there are events and fromBlockLogIndex is undefined", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      const messageSentEventFilter = l2MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 0;
      const toBlock = "latest";
      expect(await l2MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock)).toStrictEqual(
        EventParser.filterAndParseEvents([testMessageSentEvent]),
      );
    });

    it("should return filtered and parsed events when fromBlockLogIndex is defined", async () => {
      jest.spyOn(l2MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      const messageSentEventFilter = l2MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 0;
      const toBlock = "latest";
      const fromBlockLogIndex = 1;
      expect(
        await l2MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock, fromBlockLogIndex),
      ).toStrictEqual(EventParser.filterAndParseEvents([testMessageSentEvent]));
    });
  });

  describe("getMessageStatus", () => {
    it("should return UNKNOWN when on chain message status === 0", async () => {
      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        inboxL1L2MessageStatus: jest.fn().mockResolvedValueOnce(BigNumber.from(0)),
      } as unknown as L2MessageService);

      expect(await l2MessageServiceContract.getMessageStatus(TEST_MESSAGE_HASH)).toStrictEqual(
        OnChainMessageStatus.UNKNOWN,
      );

      undoMockProperty(l2MessageServiceContract, "contract");
    });
    it("should return CLAIMABLE when on chain message status === 1", async () => {
      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        inboxL1L2MessageStatus: jest.fn().mockResolvedValueOnce(BigNumber.from(1)),
      } as unknown as L2MessageService);

      expect(await l2MessageServiceContract.getMessageStatus(TEST_MESSAGE_HASH)).toStrictEqual(
        OnChainMessageStatus.CLAIMABLE,
      );
      undoMockProperty(l2MessageServiceContract, "contract");
    });

    it("should return CLAIMED when on chain message status === 2", async () => {
      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        inboxL1L2MessageStatus: jest.fn().mockResolvedValueOnce(BigNumber.from(2)),
      } as unknown as L2MessageService);

      expect(await l2MessageServiceContract.getMessageStatus(TEST_MESSAGE_HASH)).toStrictEqual(
        OnChainMessageStatus.CLAIMED,
      );
      undoMockProperty(l2MessageServiceContract, "contract");
    });
  });

  describe("estimateClaimGas", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      const messageServiceContract = new L2MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      const message = generateMessage();
      await expect(messageServiceContract.estimateClaimGas(message)).rejects.toThrow(
        new Error("'EstimateClaimGas' function not callable using readOnly mode."),
      );
    });

    it("should throw a GasEstimationError when the gas estimation failed", async () => {
      const message = generateMessage();
      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        estimateGas: {
          ...l2MessageServiceContract.contract.estimateGas,
          claimMessage: jest.fn().mockRejectedValueOnce(new Error("Gas estimation failed")),
        } as { [name: string]: ContractFunction<BigNumber> },
      } as L2MessageService);

      await expect(l2MessageServiceContract.estimateClaimGas(message)).rejects.toThrow(
        new GasEstimationError(new Error("Gas estimation failed"), message),
      );

      undoMockProperty(l2MessageServiceContract, "contract");
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { feeRecipient, ...message } = generateMessage();
      const estimatedGasLimit = BigNumber.from(50_000);

      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        estimateGas: {
          ...l2MessageServiceContract.contract.estimateGas,
          claimMessage: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        } as { [name: string]: ContractFunction<BigNumber> },
      } as L2MessageService);

      const claimMessageSpy = jest.spyOn(l2MessageServiceContract.contract.estimateGas, "claimMessage");

      await l2MessageServiceContract.estimateClaimGas(message);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        ethers.constants.AddressZero,
        message.calldata,
        message.messageNonce,
        {},
      );
      undoMockProperty(l2MessageServiceContract, "contract");
    });

    it("should return estimated gas limit for the claim message transaction", async () => {
      const message = generateMessage();
      const estimatedGasLimit = BigNumber.from(50_000);

      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        estimateGas: {
          ...l2MessageServiceContract.contract.estimateGas,
          claimMessage: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        } as { [name: string]: ContractFunction<BigNumber> },
      } as L2MessageService);

      const claimMessageSpy = jest.spyOn(l2MessageServiceContract.contract.estimateGas, "claimMessage");

      expect(await l2MessageServiceContract.estimateClaimGas(message)).toStrictEqual(estimatedGasLimit);
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

      undoMockProperty(l2MessageServiceContract, "contract");
    });
  });

  describe("claim", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      const messageServiceContract = new L2MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      const message = generateMessage();
      await expect(messageServiceContract.claim(message)).rejects.toThrow(
        new Error("'claim' function not callable using readOnly mode."),
      );
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { feeRecipient, ...message } = generateMessage();

      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        claimMessage: jest.fn().mockResolvedValueOnce(generateTransactionResponse()),
      } as unknown as L2MessageService);

      const claimMessageSpy = jest.spyOn(l2MessageServiceContract.contract, "claimMessage");

      await l2MessageServiceContract.claim(message);

      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        ethers.constants.AddressZero,
        message.calldata,
        message.messageNonce,
        {},
      );

      undoMockProperty(l2MessageServiceContract, "contract");
    });

    it("should return execute claim message transaction", async () => {
      const message = generateMessage();
      const transactionResponse = generateTransactionResponse();
      mockProperty(l2MessageServiceContract, "contract", {
        ...l2MessageServiceContract.contract,
        claimMessage: jest.fn().mockResolvedValueOnce(transactionResponse),
      } as unknown as L2MessageService);

      const claimMessageSpy = jest.spyOn(l2MessageServiceContract.contract, "claimMessage");

      expect(await l2MessageServiceContract.claim(message)).toStrictEqual(transactionResponse);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);

      undoMockProperty(l2MessageServiceContract, "contract");
    });
  });

  describe("retryTransactionWithHigherFee", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      const messageServiceContract = new L2MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      await expect(messageServiceContract.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new Error("'retryTransactionWithHigherFee' function not callable using readOnly mode."),
      );
    });

    it("should throw an error when mode = 'read-write' and this.signer is undefined", async () => {
      jest
        .spyOn(L2MessageServiceContract.prototype, "getContract")
        .mockReturnValueOnce(L2MessageService__factory.connect(TEST_CONTRACT_ADDRESS_1, getTestL2Signer()));

      const messageServiceContract = new L2MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-write");
      await expect(messageServiceContract.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new Error("Please provide a signer."),
      );
    });

    it("should return null when transactionHash param is undefined", async () => {
      expect(await l2MessageServiceContract.retryTransactionWithHigherFee()).toStrictEqual(null);
    });

    it("should retry the transaction with higher fees", async () => {
      const transactionResponse = generateTransactionResponse();
      const getTransactionSpy = jest
        .spyOn(l2MessageServiceContract.provider, "getTransaction")
        .mockResolvedValueOnce(transactionResponse);
      const sendTransactionSpy = jest
        .spyOn(l2MessageServiceContract.provider, "sendTransaction")
        .mockResolvedValueOnce(transactionResponse);
      mockProperty(l2MessageServiceContract, "signer", {
        ...l2MessageServiceContract.signer,
        signTransaction: jest.fn().mockResolvedValueOnce(""),
      } as unknown as Signer);

      expect(await l2MessageServiceContract.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).toStrictEqual(
        transactionResponse,
      );
      expect(getTransactionSpy).toHaveBeenCalledTimes(1);
      expect(sendTransactionSpy).toHaveBeenCalledTimes(1);

      undoMockProperty(l2MessageServiceContract, "signer");
    });
  });
});
