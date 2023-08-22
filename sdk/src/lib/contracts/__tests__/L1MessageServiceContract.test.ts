import { JsonRpcProvider } from "@ethersproject/providers";
import { describe, afterEach, it, expect, beforeEach } from "@jest/globals";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { BigNumber, ContractFunction, Signer, ethers } from "ethers";
import { L1MessageServiceContract } from "../L1MessageServiceContract";
import { getTestL1Signer } from "../../utils/testHelpers/contracts";
import { ZkEvmV2, ZkEvmV2__factory } from "../../../typechain";
import {
  TEST_ADDRESS_1,
  TEST_ADDRESS_2,
  TEST_CONTRACT_ADDRESS_1,
  TEST_MESSAGE_HASH,
  TEST_TRANSACTION_HASH,
  testMessageClaimedEvent,
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

describe("L1MessageServiceContract", () => {
  let providerMock: MockProxy<JsonRpcProvider>;
  let l1MessageServiceContract: L1MessageServiceContract;

  beforeEach(() => {
    providerMock = mock<JsonRpcProvider>();
    l1MessageServiceContract = new L1MessageServiceContract(
      providerMock,
      TEST_CONTRACT_ADDRESS_1,
      "read-write",
      getTestL1Signer(),
    );
  });

  afterEach(() => {
    mockClear(providerMock);
  });

  describe("getContractAbi", () => {
    it("should return ZkEvmV2 contract abi", async () => {
      expect(l1MessageServiceContract.getContractAbi()).toStrictEqual(ZkEvmV2__factory.abi);
    });
  });

  describe("getMessageByMessageHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([]);
      expect(await l1MessageServiceContract.getMessageByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(null);
    });

    it("should return message when message hash exists", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      expect(await l1MessageServiceContract.getMessageByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(
        mapMessageSentEventOrLogToMessage(testMessageSentEvent),
      );
    });
  });

  describe("getMessagesByTransactionHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest
        .spyOn(l1MessageServiceContract.provider, "getTransactionReceipt")
        .mockImplementationOnce(jest.fn().mockResolvedValueOnce(null));
      expect(await l1MessageServiceContract.getMessagesByTransactionHash(TEST_TRANSACTION_HASH)).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(l1MessageServiceContract.provider, "getTransactionReceipt").mockResolvedValueOnce(transactionReceipt);
      expect(await l1MessageServiceContract.getMessagesByTransactionHash(TEST_MESSAGE_HASH)).toStrictEqual([
        mapMessageSentEventOrLogToMessage(ZkEvmV2__factory.createInterface().parseLog(transactionReceipt.logs[0])),
      ]);
    });
  });

  describe("getTransactionReceiptByMessageHash", () => {
    it("should return null when message hash does not exist", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([]);
      expect(await l1MessageServiceContract.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(null);
    });

    it("should return null when transaction receipt does not exist", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      jest
        .spyOn(l1MessageServiceContract.provider, "getTransactionReceipt")
        .mockImplementationOnce(jest.fn().mockResolvedValueOnce(null));
      expect(await l1MessageServiceContract.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(null);
    });

    it("should return an array of messages when transaction hash exists and contains MessageSent events", async () => {
      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      jest.spyOn(l1MessageServiceContract.provider, "getTransactionReceipt").mockResolvedValueOnce(transactionReceipt);
      expect(await l1MessageServiceContract.getTransactionReceiptByMessageHash(TEST_MESSAGE_HASH)).toStrictEqual(
        transactionReceipt,
      );
    });
  });

  describe("getContract", () => {
    it("should throw an error when mode = 'read-write' and signer params is undefined", () => {
      try {
        l1MessageServiceContract.getContract(TEST_CONTRACT_ADDRESS_1);
      } catch (e) {
        expect(e).toEqual(new Error("Please provide a signer."));
      }
    });

    it("should return ZkEvmV2 contract instance with signer when mode = 'read-write'", () => {
      const signer = getTestL1Signer();
      const zkevmV2Instance = l1MessageServiceContract.getContract(TEST_CONTRACT_ADDRESS_1, signer);
      expect(JSON.stringify(zkevmV2Instance)).toStrictEqual(
        JSON.stringify(ZkEvmV2__factory.connect(TEST_CONTRACT_ADDRESS_1, signer)),
      );
    });

    it("should return ZkEvmV2 contract instance with provider when mode = 'read-only'", () => {
      const messageServiceContract = new L1MessageServiceContract(
        providerMock,
        TEST_CONTRACT_ADDRESS_1,
        "read-only",
        getTestL1Signer(),
      );
      const zkevmV2Instance = messageServiceContract.getContract(TEST_CONTRACT_ADDRESS_1);
      expect(JSON.stringify(zkevmV2Instance)).toStrictEqual(
        JSON.stringify(ZkEvmV2__factory.connect(TEST_CONTRACT_ADDRESS_1, messageServiceContract.provider)),
      );
    });
  });

  describe("getCurrentNonce", () => {
    it("should throw an error when mode = 'read-write' and this.signer is undefined", async () => {
      jest
        .spyOn(L1MessageServiceContract.prototype, "getContract")
        .mockReturnValueOnce(ZkEvmV2__factory.connect(TEST_CONTRACT_ADDRESS_1, getTestL1Signer()));

      const messageServiceContract = new L1MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-write");

      await expect(messageServiceContract.getCurrentNonce()).rejects.toThrow(new Error("Please provide a signer."));
    });

    it("should throw an error when mode = 'read-only' and accountAddress param is undefined", async () => {
      const messageServiceContract = new L1MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      await expect(messageServiceContract.getCurrentNonce()).rejects.toThrow(
        new Error("Please provider an account address."),
      );
    });

    it("should return account nonce for the account address passed in params", async () => {
      const messageServiceContract = new L1MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      jest.spyOn(messageServiceContract.provider, "getTransactionCount").mockResolvedValueOnce(10);
      expect(await messageServiceContract.getCurrentNonce(TEST_ADDRESS_1)).toStrictEqual(10);
    });

    it("should return account nonce for this.signer address when mode = 'read-write'", async () => {
      const getTransactionCountSpy = jest
        .spyOn(l1MessageServiceContract.provider, "getTransactionCount")
        .mockResolvedValueOnce(10);
      expect(await l1MessageServiceContract.getCurrentNonce()).toStrictEqual(10);

      expect(getTransactionCountSpy).toHaveBeenCalledTimes(1);
      expect(getTransactionCountSpy).toHaveBeenCalledWith(await getTestL1Signer().getAddress());
    });
  });

  describe("getCurrentBlockNumber", () => {
    it("should throw an error when mode = 'read-write' and this.signer is undefined", async () => {
      const currentBlockNumber = 50_000;
      jest.spyOn(l1MessageServiceContract.provider, "getBlockNumber").mockResolvedValueOnce(currentBlockNumber);
      expect(await l1MessageServiceContract.getCurrentBlockNumber()).toStrictEqual(currentBlockNumber);
    });
  });

  describe("getEvents", () => {
    it("should return empty array when there are no events", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([]);
      const messageSentEventFilter = l1MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 0;
      const toBlock = "latest";
      expect(await l1MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock)).toStrictEqual([]);
    });

    it("should return empty array when there are events but event.blockNumber === fromBlock && event.logIndex < fromBlockLogIndex", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      const messageSentEventFilter = l1MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 51;
      const toBlock = "latest";
      const fromBlockLogIndex = 2;
      expect(
        await l1MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock, fromBlockLogIndex),
      ).toStrictEqual([]);
    });

    it("should return filtered and parsed events when there are events and fromBlockLogIndex is undefined", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      const messageSentEventFilter = l1MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 0;
      const toBlock = "latest";
      expect(await l1MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock)).toStrictEqual(
        EventParser.filterAndParseEvents([testMessageSentEvent]),
      );
    });

    it("should return filtered and parsed events when fromBlockLogIndex is defined", async () => {
      jest.spyOn(l1MessageServiceContract.contract, "queryFilter").mockResolvedValueOnce([testMessageSentEvent]);
      const messageSentEventFilter = l1MessageServiceContract.contract.filters.MessageSent();
      const fromBlock = 0;
      const toBlock = "latest";
      const fromBlockLogIndex = 1;
      expect(
        await l1MessageServiceContract.getEvents(messageSentEventFilter, fromBlock, toBlock, fromBlockLogIndex),
      ).toStrictEqual(EventParser.filterAndParseEvents([testMessageSentEvent]));
    });
  });

  describe("getMessageStatus", () => {
    it("should return UNKNOWN when on chain message status === 0 and there is no MessageClaimed event", async () => {
      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        inboxL2L1MessageStatus: jest.fn().mockResolvedValueOnce(BigNumber.from(0)),
        queryFilter: jest.fn().mockResolvedValueOnce([]),
      } as unknown as ZkEvmV2);

      expect(await l1MessageServiceContract.getMessageStatus(TEST_MESSAGE_HASH)).toStrictEqual(
        OnChainMessageStatus.UNKNOWN,
      );

      undoMockProperty(l1MessageServiceContract, "contract");
    });
    it("should return CLAIMABLE when on chain message status === 1", async () => {
      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        inboxL2L1MessageStatus: jest.fn().mockResolvedValueOnce(BigNumber.from(1)),
      } as unknown as ZkEvmV2);

      expect(await l1MessageServiceContract.getMessageStatus(TEST_MESSAGE_HASH)).toStrictEqual(
        OnChainMessageStatus.CLAIMABLE,
      );
      undoMockProperty(l1MessageServiceContract, "contract");
    });

    it("should return CLAIMED when on chain message status === 0 and there is a MessageClaimed event", async () => {
      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        inboxL2L1MessageStatus: jest.fn().mockResolvedValueOnce(BigNumber.from(0)),
        queryFilter: jest.fn().mockResolvedValueOnce([testMessageClaimedEvent]),
      } as unknown as ZkEvmV2);

      expect(await l1MessageServiceContract.getMessageStatus(TEST_MESSAGE_HASH)).toStrictEqual(
        OnChainMessageStatus.CLAIMED,
      );
      undoMockProperty(l1MessageServiceContract, "contract");
    });
  });

  describe("estimateClaimGas", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      const messageServiceContract = new L1MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      const message = generateMessage();
      await expect(messageServiceContract.estimateClaimGas(message)).rejects.toThrow(
        new Error("'EstimateClaimGas' function not callable using readOnly mode."),
      );
    });

    it("should throw a GasEstimationError when the gas estimation failed", async () => {
      const message = generateMessage();
      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        estimateGas: {
          ...l1MessageServiceContract.contract.estimateGas,
          claimMessage: jest.fn().mockRejectedValueOnce(new Error("Gas estimation failed")),
        } as { [name: string]: ContractFunction<BigNumber> },
      } as ZkEvmV2);
      jest
        .spyOn(l1MessageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });

      await expect(l1MessageServiceContract.estimateClaimGas(message)).rejects.toThrow(
        new GasEstimationError(new Error("Gas estimation failed"), message),
      );

      undoMockProperty(l1MessageServiceContract, "contract");
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { feeRecipient, ...message } = generateMessage();
      const estimatedGasLimit = BigNumber.from(50_000);

      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        estimateGas: {
          ...l1MessageServiceContract.contract.estimateGas,
          claimMessage: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        } as { [name: string]: ContractFunction<BigNumber> },
      } as ZkEvmV2);

      const claimMessageSpy = jest.spyOn(l1MessageServiceContract.contract.estimateGas, "claimMessage");
      jest
        .spyOn(l1MessageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });

      await l1MessageServiceContract.estimateClaimGas(message);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        ethers.constants.AddressZero,
        message.calldata,
        message.messageNonce,
        { maxFeePerGas: BigNumber.from(100_000_000) },
      );
      undoMockProperty(l1MessageServiceContract, "contract");
    });

    it("should return estimated gas limit for the claim message transaction", async () => {
      const message = generateMessage();
      const estimatedGasLimit = BigNumber.from(50_000);

      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        estimateGas: {
          ...l1MessageServiceContract.contract.estimateGas,
          claimMessage: jest.fn().mockResolvedValueOnce(estimatedGasLimit),
        } as { [name: string]: ContractFunction<BigNumber> },
      } as ZkEvmV2);

      const claimMessageSpy = jest.spyOn(l1MessageServiceContract.contract.estimateGas, "claimMessage");
      const get1559FeesSpy = jest
        .spyOn(l1MessageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });

      expect(await l1MessageServiceContract.estimateClaimGas(message)).toStrictEqual(estimatedGasLimit);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        TEST_ADDRESS_2,
        message.calldata,
        message.messageNonce,
        { maxFeePerGas: BigNumber.from(100_000_000) },
      );
      expect(get1559FeesSpy).toHaveBeenCalledTimes(1);

      undoMockProperty(l1MessageServiceContract, "contract");
    });
  });

  describe("claim", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      const messageServiceContract = new L1MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      const message = generateMessage();
      await expect(messageServiceContract.claim(message)).rejects.toThrow(
        new Error("'claim' function not callable using readOnly mode."),
      );
    });

    it("should set feeRecipient === ZeroAddress when feeRecipient param is undefined", async () => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { feeRecipient, ...message } = generateMessage();

      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        claimMessage: jest.fn().mockResolvedValueOnce(generateTransactionResponse()),
      } as unknown as ZkEvmV2);

      const claimMessageSpy = jest.spyOn(l1MessageServiceContract.contract, "claimMessage");
      const get1559FeesSpy = jest
        .spyOn(l1MessageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });

      await l1MessageServiceContract.claim(message);

      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(claimMessageSpy).toHaveBeenCalledWith(
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        ethers.constants.AddressZero,
        message.calldata,
        message.messageNonce,
        { maxFeePerGas: BigNumber.from(100_000_000) },
      );
      expect(get1559FeesSpy).toHaveBeenCalledTimes(1);

      undoMockProperty(l1MessageServiceContract, "contract");
    });

    it("should return execute claim message transaction", async () => {
      const message = generateMessage();
      const transactionResponse = generateTransactionResponse();
      mockProperty(l1MessageServiceContract, "contract", {
        ...l1MessageServiceContract.contract,
        claimMessage: jest.fn().mockResolvedValueOnce(transactionResponse),
      } as unknown as ZkEvmV2);

      const claimMessageSpy = jest.spyOn(l1MessageServiceContract.contract, "claimMessage");
      const get1559FeesSpy = jest
        .spyOn(l1MessageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });

      expect(await l1MessageServiceContract.claim(message)).toStrictEqual(transactionResponse);
      expect(claimMessageSpy).toHaveBeenCalledTimes(1);
      expect(get1559FeesSpy).toHaveBeenCalledTimes(1);

      undoMockProperty(l1MessageServiceContract, "contract");
    });
  });

  describe("retryTransactionWithHigherFee", () => {
    it("should throw an error when mode = 'read-only'", async () => {
      const messageServiceContract = new L1MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-only");
      await expect(messageServiceContract.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new Error("'retryTransactionWithHigherFee' function not callable using readOnly mode."),
      );
    });

    it("should throw an error when mode = 'read-write' and this.signer is undefined", async () => {
      jest
        .spyOn(L1MessageServiceContract.prototype, "getContract")
        .mockReturnValueOnce(ZkEvmV2__factory.connect(TEST_CONTRACT_ADDRESS_1, getTestL1Signer()));

      const messageServiceContract = new L1MessageServiceContract(providerMock, TEST_CONTRACT_ADDRESS_1, "read-write");
      await expect(messageServiceContract.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).rejects.toThrow(
        new Error("Please provide a signer."),
      );
    });

    it("should return null when transactionHash param is undefined", async () => {
      expect(await l1MessageServiceContract.retryTransactionWithHigherFee()).toStrictEqual(null);
    });

    it("should retry the transaction with higher fees", async () => {
      const transactionResponse = generateTransactionResponse();
      const getTransactionSpy = jest
        .spyOn(l1MessageServiceContract.provider, "getTransaction")
        .mockResolvedValueOnce(transactionResponse);
      const get1559FeesSpy = jest
        .spyOn(l1MessageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });
      const sendTransactionSpy = jest
        .spyOn(l1MessageServiceContract.provider, "sendTransaction")
        .mockResolvedValueOnce(transactionResponse);
      mockProperty(l1MessageServiceContract, "signer", {
        ...l1MessageServiceContract.signer,
        signTransaction: jest.fn().mockResolvedValueOnce(""),
      } as unknown as Signer);

      expect(await l1MessageServiceContract.retryTransactionWithHigherFee(TEST_TRANSACTION_HASH)).toStrictEqual(
        transactionResponse,
      );
      expect(getTransactionSpy).toHaveBeenCalledTimes(1);
      expect(get1559FeesSpy).toHaveBeenCalledTimes(1);
      expect(sendTransactionSpy).toHaveBeenCalledTimes(1);

      undoMockProperty(l1MessageServiceContract, "signer");
    });
  });
});
