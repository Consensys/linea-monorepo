import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { Result } from "ethers/lib/utils";
import { LoggerOptions } from "winston";
import { L1MessageServiceContract } from "../../../contracts";
import { getTestL1Signer } from "../../../utils/testHelpers/contracts";
import {
  TEST_ADDRESS_1,
  TEST_ADDRESS_2,
  TEST_CONTRACT_ADDRESS_1,
  TEST_MESSAGE_HASH,
  testL1NetworkConfig,
} from "../../../utils/testHelpers/constants";
import { generateMessageFromDb } from "../../../utils/testHelpers/helpers";
import { L1SentEventListener } from "../";
import { DatabaseErrorType, DatabaseRepoName, Direction } from "../../utils/enums";
import { DatabaseAccessError } from "../../utils/errors";
import { SentEventListener } from "../SentEventListener";
import { LineaLogger, getLogger } from "../../../logger";
import { L1NetworkConfig } from "../../utils/types";
import { DEFAULT_LISTENER_INTERVAL } from "../../../utils/constants";

class TestSentEventListener extends SentEventListener<L1MessageServiceContract> {
  public logger: LineaLogger;
  public override onlyEOA: boolean;

  constructor(
    dataSource: DataSource,
    messageServiceContract: L1MessageServiceContract,
    config: L1NetworkConfig,
    direction: Direction,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, messageServiceContract, config, direction);
    this.logger = getLogger(L1SentEventListener.name, loggerOptions);
    this.onlyEOA = super.onlyEOA;
  }

  public override shouldExcludeMessage(messageCalldata: string, messageHash: string): boolean {
    return super.shouldExcludeMessage(messageCalldata, messageHash);
  }
}

describe("SentEventListener", () => {
  let sentEventListener: TestSentEventListener;
  let messageServiceContract: L1MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L1MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_1,
      "read-write",
      getTestL1Signer(),
    );

    sentEventListener = new TestSentEventListener(
      mock<DataSource>(),
      messageServiceContract,
      testL1NetworkConfig,
      Direction.L2_TO_L1,
      {
        silent: true,
      },
    );
  });

  describe("start", () => {
    it("should start the SentEventListener service", async () => {
      const getBlockNumberSpy = jest
        .spyOn(messageServiceContract.contract.provider, "getBlockNumber")
        .mockResolvedValue(100_000);
      const getLatestMessageSentBlockNumberSpy = jest
        .spyOn(sentEventListener, "getLatestMessageSentBlockNumber")
        .mockResolvedValue(100_000);
      const listenForMessageSentEventsSpy = jest
        .spyOn(sentEventListener, "listenForMessageSentEvents")
        .mockResolvedValue();

      await sentEventListener.start();

      expect(getBlockNumberSpy).toHaveBeenCalledTimes(1);
      expect(getLatestMessageSentBlockNumberSpy).toHaveBeenCalledTimes(1);
      expect(listenForMessageSentEventsSpy).toHaveBeenCalledTimes(1);
      expect(listenForMessageSentEventsSpy).toHaveBeenCalledWith(DEFAULT_LISTENER_INTERVAL, 100_000, 0);

      sentEventListener.stop();
    });

    it("should start the SentEventListener service from intialBlock from config if it is defined", async () => {
      const initialFromBlock = 10;
      const sentEventListenerWithInitialBlock = new TestSentEventListener(
        mock<DataSource>(),
        messageServiceContract,
        {
          ...testL1NetworkConfig,
          listener: {
            ...testL1NetworkConfig.listener,
            initialFromBlock,
          },
        },
        Direction.L2_TO_L1,
        {
          silent: true,
        },
      );

      const getBlockNumberSpy = jest
        .spyOn(messageServiceContract.contract.provider, "getBlockNumber")
        .mockResolvedValue(100_000);
      const getLatestMessageSentBlockNumberSpy = jest
        .spyOn(sentEventListenerWithInitialBlock, "getLatestMessageSentBlockNumber")
        .mockResolvedValue(100_000);
      const listenForMessageSentEventsSpy = jest
        .spyOn(sentEventListenerWithInitialBlock, "listenForMessageSentEvents")
        .mockResolvedValue();

      await sentEventListenerWithInitialBlock.start();

      expect(getBlockNumberSpy).toHaveBeenCalledTimes(1);
      expect(getLatestMessageSentBlockNumberSpy).toHaveBeenCalledTimes(1);
      expect(listenForMessageSentEventsSpy).toHaveBeenCalledTimes(1);
      expect(listenForMessageSentEventsSpy).toHaveBeenCalledWith(DEFAULT_LISTENER_INTERVAL, initialFromBlock, 0);

      sentEventListenerWithInitialBlock.stop();
    });
  });

  describe("stop", () => {
    it("should start the SentEventListener service", async () => {
      const loggerSpy = jest.spyOn(sentEventListener.logger, "info");
      sentEventListener.stop();

      expect(loggerSpy).toHaveBeenCalledTimes(2);
      expect(loggerSpy).toHaveBeenNthCalledWith(1, "Stopping SentEventListener...");
      expect(loggerSpy).toHaveBeenNthCalledWith(2, "SentEventListener stopped.");
    });
  });

  describe("calculateFromBlockNumber", () => {
    it("should return 'toBlockNumber' if 'fromBlockNumber' > 'toBlockNumber'", async () => {
      const fromBlockNumber = 10;
      const toBlockNumber = 8;

      expect(sentEventListener.calculateFromBlockNumber(fromBlockNumber, toBlockNumber)).toStrictEqual(toBlockNumber);
    });

    it("should return 0 if 'fromBlockNumber' <= 'toBlockNumber' and fromBlockNumber < 0", async () => {
      const fromBlockNumber = -1;
      const toBlockNumber = 12;

      expect(sentEventListener.calculateFromBlockNumber(fromBlockNumber, toBlockNumber)).toStrictEqual(0);
    });

    it("should return 'fromBlockNumber' if 'fromBlockNumber' <= 'toBlockNumber'", async () => {
      const fromBlockNumber = 10;
      const toBlockNumber = 12;

      expect(sentEventListener.calculateFromBlockNumber(fromBlockNumber, toBlockNumber)).toStrictEqual(fromBlockNumber);
    });
  });

  describe("getLatestMessageSentBlockNumber", () => {
    it("should return null when there is no message in the DB", async () => {
      jest.spyOn(sentEventListener.messageRepository, "getLatestMessageSent").mockResolvedValueOnce(null);

      const latestMessageSentBlockNumber = await sentEventListener.getLatestMessageSentBlockNumber(Direction.L2_TO_L1);
      expect(latestMessageSentBlockNumber).toStrictEqual(null);
    });

    it("should return null when DB query failed", async () => {
      jest
        .spyOn(sentEventListener.messageRepository, "getLatestMessageSent")
        .mockRejectedValueOnce(new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, {}));

      const latestMessageSentBlockNumber = await sentEventListener.getLatestMessageSentBlockNumber(Direction.L2_TO_L1);
      expect(latestMessageSentBlockNumber).toStrictEqual(null);
    });

    it("should return latest message sent block number", async () => {
      const message = generateMessageFromDb();
      jest.spyOn(sentEventListener.messageRepository, "getLatestMessageSent").mockResolvedValueOnce(message);

      const latestMessageSentBlockNumber = await sentEventListener.getLatestMessageSentBlockNumber(Direction.L2_TO_L1);
      expect(latestMessageSentBlockNumber).toStrictEqual(message.sentBlockNumber);
    });
  });

  describe("listenForMessageSentEvents", () => {
    beforeEach(() => {
      sentEventListener.shouldStopListening = true;
    });

    it("should catch any error and log it when ethers query failed", async () => {
      const loggerErrorSpy = jest.spyOn(sentEventListener.logger, "error");
      jest
        .spyOn(sentEventListener.messageServiceContract, "getEvents")
        .mockRejectedValueOnce(new Error("Ethers error"));

      await sentEventListener.listenForMessageSentEvents(1000, 100_000, 0);
      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenLastCalledWith(
        `Error found in listenForMessageSentEvents:\nFounded error: ${JSON.stringify(
          new Error("Ethers error"),
        )}\nParsed error: ${JSON.stringify({
          errorCode: "UNKNOWN_ERROR",
          mitigation: { shouldRetry: true, retryWithBlocking: true, retryPeriodInMs: 5000 },
        })}`,
      );
    });

    it("should catch DatabaseAccessError when DB query failed and restart from the current message block number", async () => {
      const loggerWarnSpy = jest.spyOn(sentEventListener.logger, "warn");
      jest.spyOn(sentEventListener.messageServiceContract, "getEvents").mockResolvedValueOnce([
        {
          args: {
            _from: TEST_ADDRESS_1,
            _to: TEST_ADDRESS_2,
            _fee: 0,
            _value: 1,
            _nonce: 1,
            _calldata: "0x",
            _messageHash: TEST_MESSAGE_HASH,
          } as unknown as Result,
          blockNumber: 5,
          logIndex: 0,
          contractAddress: TEST_CONTRACT_ADDRESS_1,
          transactionHash: "transaction-hash",
        },
      ]);

      const message = generateMessageFromDb();
      jest
        .spyOn(sentEventListener.messageRepository, "insertMessage")
        .mockRejectedValueOnce(
          new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Insert, {}, message),
        );

      await sentEventListener.listenForMessageSentEvents(1000, 100_000, 0);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
    });

    it("should insert messages in the DB", async () => {
      const loggerInfoSpy = jest.spyOn(sentEventListener.logger, "info");
      jest.spyOn(sentEventListener.messageServiceContract, "getCurrentBlockNumber").mockResolvedValueOnce(10);
      jest.spyOn(sentEventListener.messageServiceContract, "getEvents").mockResolvedValueOnce([
        {
          args: {
            _from: TEST_ADDRESS_1,
            _to: TEST_ADDRESS_2,
            _fee: 0,
            _value: 1,
            _nonce: 1,
            _calldata: "0x",
            _messageHash: TEST_MESSAGE_HASH,
          } as unknown as Result,
          blockNumber: 5,
          logIndex: 0,
          contractAddress: TEST_CONTRACT_ADDRESS_1,
          transactionHash: "transaction-hash",
        },
      ]);

      const repositorySpy = jest.spyOn(sentEventListener.messageRepository, "insertMessage").mockResolvedValueOnce();

      sentEventListener.shouldStopListening = false;

      await sentEventListener.listenForMessageSentEvents(1000, 100_000, 0);
      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(loggerInfoSpy).toHaveBeenLastCalledWith(`message: ${JSON.stringify(TEST_MESSAGE_HASH)}`);

      expect(repositorySpy).toHaveBeenCalledTimes(1);

      sentEventListener.shouldStopListening = true;
    });
  });

  describe("shouldExcludeMessage", () => {
    it("Should return false if onlyEOA flag is false", () => {
      const messageCallData = "0x";
      const messageHash = TEST_MESSAGE_HASH;
      expect(sentEventListener.shouldExcludeMessage(messageCallData, messageHash)).toBeFalsy();
    });

    it("Should return false if onlyEOA flag is true and messageCalldata is empty bytes", () => {
      const messageCallData = "0x";
      const messageHash = TEST_MESSAGE_HASH;

      sentEventListener.onlyEOA = true;

      expect(sentEventListener.shouldExcludeMessage(messageCallData, messageHash)).toBeFalsy();
    });

    it("Should return true if onlyEOA flag is true and messageCalldata is not empty bytes", () => {
      const messageCallData = "0x01";
      const messageHash = TEST_MESSAGE_HASH;

      const loggerSpy = jest.spyOn(sentEventListener.logger, "warn");
      sentEventListener.onlyEOA = true;

      expect(sentEventListener.shouldExcludeMessage(messageCallData, messageHash)).toBeTruthy();

      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith(
        `Message with hash ${messageHash} has been excluded because target address is not an EOA or calldata is not empty.`,
      );
    });
  });
});
