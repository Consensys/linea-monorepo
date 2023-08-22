import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { LoggerOptions } from "winston";
import { L1AnchoredEventListener } from "../";
import { L1MessageServiceContract } from "../../../contracts";
import { getTestL1Signer } from "../../../utils/testHelpers/contracts";
import { TEST_CONTRACT_ADDRESS_1, testL1NetworkConfig } from "../../../utils/testHelpers/constants";
import { generateMessageFromDb } from "../../../utils/testHelpers/helpers";
import { OnChainMessageStatus } from "../../../utils/enum";
import { Direction, MessageStatus } from "../../utils/enums";
import { AnchoredEventListener } from "../AnchoredEventListener";
import { L1NetworkConfig } from "../../utils/types";
import { LineaLogger, getLogger } from "../../../logger";
import { DEFAULT_LISTENER_INTERVAL, DEFAULT_MAX_FETCH_MESSAGES_FROM_DB } from "../../../utils/constants";

class TestAnchoredEventListener extends AnchoredEventListener<L1MessageServiceContract> {
  public logger: LineaLogger;
  public polling: number;
  public maxFetchMessagesFromDbOverride: number;

  constructor(
    dataSource: DataSource,
    messageServiceContract: L1MessageServiceContract,
    config: L1NetworkConfig,
    direction: Direction,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, messageServiceContract, config, direction);
    this.logger = getLogger(L1AnchoredEventListener.name, loggerOptions);
    this.polling = this.pollingInterval;
    this.maxFetchMessagesFromDbOverride = this.maxFetchMessagesFromDb;
  }
}

describe("AnchoredEventListener", () => {
  let anchoredEventListener: TestAnchoredEventListener;
  let messageServiceContract: L1MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L1MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_1,
      "read-write",
      getTestL1Signer(),
    );

    anchoredEventListener = new TestAnchoredEventListener(
      mock<DataSource>(),
      messageServiceContract,
      testL1NetworkConfig,
      Direction.L2_TO_L1,
      {
        silent: true,
      },
    );

    jest.spyOn(anchoredEventListener.messageRepository, "updateMessage").mockResolvedValue();
  });

  describe("constructor", () => {
    it("Should set the config with default value when there are not in the config ", () => {
      const anchoredEventListenerWithoutConfig = new TestAnchoredEventListener(
        mock<DataSource>(),
        messageServiceContract,
        { ...testL1NetworkConfig, listener: {} },
        Direction.L2_TO_L1,
        {
          silent: true,
        },
      );

      expect(anchoredEventListenerWithoutConfig.polling).toStrictEqual(DEFAULT_LISTENER_INTERVAL);
      expect(anchoredEventListenerWithoutConfig.maxFetchMessagesFromDbOverride).toStrictEqual(
        DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
      );
    });
  });

  describe("start", () => {
    it("Should start the AnchoredEventListener service", async () => {
      const loggerSpy = jest.spyOn(anchoredEventListener.logger, "info");
      jest.spyOn(anchoredEventListener, "listenForMessageAnchoringEvents").mockResolvedValueOnce();

      anchoredEventListener.start();

      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith("Starting AnchoredEventListener...");

      anchoredEventListener.stop();
    });
  });

  describe("stop", () => {
    it("Should start the AnchoredEventListener service", async () => {
      const loggerSpy = jest.spyOn(anchoredEventListener.logger, "info");

      anchoredEventListener.stop();

      expect(loggerSpy).toHaveBeenCalledTimes(2);
      expect(loggerSpy).toHaveBeenNthCalledWith(1, "Stopping AnchoredEventListener...");
      expect(loggerSpy).toHaveBeenNthCalledWith(2, "AnchoredEventListener stopped.");
    });
  });

  describe("listenForMessageAnchoringEvents", () => {
    it("should log a warning message if the number of messages fetched from DB is equal to maxFetchMessagesFromDb", async () => {
      const loggerSpy = jest.spyOn(anchoredEventListener.logger, "warn");
      const getNFirstMessageSentSpy = jest
        .spyOn(anchoredEventListener.messageRepository, "getNFirstMessageSent")
        .mockResolvedValueOnce([
          generateMessageFromDb(),
          generateMessageFromDb({ id: 2 }),
          generateMessageFromDb({ id: 3 }),
        ]);

      await anchoredEventListener.listenForMessageAnchoringEvents();
      expect(getNFirstMessageSentSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith("Limit of messages sent to listen reached (3).");
    });

    it("should do nothing when there are no messages in the DB", async () => {
      const getNFirstMessageSentSpy = jest
        .spyOn(anchoredEventListener.messageRepository, "getNFirstMessageSent")
        .mockResolvedValueOnce([]);

      const updateMessageSpy = jest.spyOn(anchoredEventListener.messageRepository, "updateMessage");

      await anchoredEventListener.listenForMessageAnchoringEvents();
      expect(getNFirstMessageSentSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledTimes(0);
    });

    it("should failed when there is an error in the message repository or message service contract", async () => {
      const loggerSpy = jest.spyOn(anchoredEventListener.logger, "error");
      const expectedError = new Error("Error");

      const getNFirstMessageSentSpy = jest
        .spyOn(anchoredEventListener.messageRepository, "getNFirstMessageSent")
        .mockResolvedValueOnce([generateMessageFromDb()]);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockRejectedValueOnce(expectedError);

      await anchoredEventListener.listenForMessageAnchoringEvents();

      expect(getNFirstMessageSentSpy).toHaveBeenCalledTimes(1);

      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith(
        `Error found in listenForMessageAnchoringEvents:\nFounded error: ${JSON.stringify(
          expectedError,
        )}\nParsed error: ${JSON.stringify({
          errorCode: "UNKNOWN_ERROR",
          mitigation: { shouldRetry: true, retryWithBlocking: true, retryPeriodInMs: 5000 },
        })}`,
      );
    });

    it("should update message status to ANCHORED when the message has been anchored on the destination layer", async () => {
      const message = generateMessageFromDb();

      const loggerSpy = jest.spyOn(anchoredEventListener.logger, "info");
      const getNFirstMessageSentSpy = jest
        .spyOn(anchoredEventListener.messageRepository, "getNFirstMessageSent")
        .mockResolvedValueOnce([message]);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValueOnce(OnChainMessageStatus.CLAIMABLE);

      const updateMessageSpy = jest.spyOn(anchoredEventListener.messageRepository, "updateMessage");

      await anchoredEventListener.listenForMessageAnchoringEvents();

      expect(getNFirstMessageSentSpy).toHaveBeenCalledTimes(1);

      expect(updateMessageSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledWith(message.messageHash, message.direction, {
        status: MessageStatus.ANCHORED,
      });

      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith(`Message hash ${message.messageHash} has been anchored.`);
    });

    it("should update message status to CLAIMED when the message has already been claimed on the destination layer", async () => {
      const message = generateMessageFromDb();

      const loggerSpy = jest.spyOn(anchoredEventListener.logger, "info");
      const getNFirstMessageSentSpy = jest
        .spyOn(anchoredEventListener.messageRepository, "getNFirstMessageSent")
        .mockResolvedValueOnce([message]);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValueOnce(OnChainMessageStatus.CLAIMED);

      const updateMessageSpy = jest.spyOn(anchoredEventListener.messageRepository, "updateMessage");

      await anchoredEventListener.listenForMessageAnchoringEvents();

      expect(getNFirstMessageSentSpy).toHaveBeenCalledTimes(1);

      expect(updateMessageSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledWith(message.messageHash, message.direction, {
        status: MessageStatus.CLAIMED_SUCCESS,
      });

      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith(`Message with hash ${message.messageHash} has already been claimed.`);
    });
  });
});
