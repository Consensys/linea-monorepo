import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { TestLogger } from "../../../utils/testing/helpers";
import { Direction } from "../../../core/enums/MessageEnums";
import { rejectedMessageProps, testL2NetworkConfig, testMessage } from "../../../utils/testing/constants";
import { IPoller } from "../../../core/services/pollers/IPoller";
import { MessageSentEventPoller } from "../MessageSentEventPoller";
import { IMessageSentEventProcessor } from "../../../core/services/processors/IMessageSentEventProcessor";
import { IChainQuerier } from "../../../core/clients/blockchain/IChainQuerier";
import { IMessageRepository } from "../../../core/persistence/IMessageRepository";
import { wait } from "../../../core/utils/shared";
import { DatabaseAccessError } from "../../../core/errors/DatabaseErrors";
import { DatabaseErrorType, DatabaseRepoName } from "../../../core/enums/DatabaseEnums";

describe("TestMessageSentEventPoller", () => {
  let testMessageSentEventPoller: IPoller;
  const eventProcessorMock = mock<IMessageSentEventProcessor>();
  const l1QuerierMock = mock<IChainQuerier<unknown>>();
  const messageRepositoryMock = mock<IMessageRepository<unknown>>();
  const logger = new TestLogger(MessageSentEventPoller.name);

  beforeEach(() => {
    testMessageSentEventPoller = new MessageSentEventPoller(
      eventProcessorMock,
      l1QuerierMock,
      messageRepositoryMock,
      Direction.L1_TO_L2,
      testL2NetworkConfig,
      logger,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("start", () => {
    it("Should return and log as warning if it has been started", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockResolvedValue(10);
      jest.spyOn(messageRepositoryMock, "getLatestMessageSent").mockResolvedValue(null);
      jest.spyOn(eventProcessorMock, "getAndStoreMessageSentEvents").mockResolvedValue({
        nextFromBlock: 20,
        nextFromBlockLogIndex: 0,
      });

      testMessageSentEventPoller.start();
      await wait(500);
      await testMessageSentEventPoller.start();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith("%s has already started.", MessageSentEventPoller.name);
    });

    it("Should call getAndStoreMessageSentEvents and log as info if it started successfully", async () => {
      const l1QuerierMockSpy = jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockResolvedValue(10);
      const messageRepositoryMockSpy = jest
        .spyOn(messageRepositoryMock, "getLatestMessageSent")
        .mockResolvedValue(testMessage);
      const eventProcessorMockSpy = jest.spyOn(eventProcessorMock, "getAndStoreMessageSentEvents").mockResolvedValue({
        nextFromBlock: 20,
        nextFromBlockLogIndex: 0,
      });
      const loggerInfoSpy = jest.spyOn(logger, "info");

      testMessageSentEventPoller.start();
      await wait(500);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith("Starting %s %s...", Direction.L1_TO_L2, MessageSentEventPoller.name);
      expect(l1QuerierMockSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryMockSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryMockSpy).toHaveBeenCalledWith(
        Direction.L1_TO_L2,
        testL2NetworkConfig.messageServiceContractAddress,
      );
      expect(eventProcessorMockSpy).toHaveBeenCalled();
      expect(eventProcessorMockSpy).toHaveBeenCalledWith(20, 0);
    });

    it("Should log as warning if getCurrentBlockNumber throws error", async () => {
      const error = new Error("Other error for testing");
      const l1QuerierMockSpy = jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockRejectedValue(error);
      const loggerErrorSpy = jest.spyOn(logger, "error");

      await testMessageSentEventPoller.start();
      await wait(500);

      expect(loggerErrorSpy).toHaveBeenCalled();
      expect(loggerErrorSpy).toHaveBeenCalledWith(error);
      expect(l1QuerierMockSpy).toHaveBeenCalled();
    });

    it("Should log as warning if getAndStoreMessageSentEvents throws DatabaseAccessError", async () => {
      const l1QuerierMockSpy = jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockResolvedValue(10);
      const messageRepositoryMockSpy = jest
        .spyOn(messageRepositoryMock, "getLatestMessageSent")
        .mockResolvedValue(testMessage);
      const eventProcessorMockSpy = jest
        .spyOn(eventProcessorMock, "getAndStoreMessageSentEvents")
        .mockRejectedValue(
          new DatabaseAccessError(
            DatabaseRepoName.MessageRepository,
            DatabaseErrorType.Read,
            new Error("read database error."),
            { ...rejectedMessageProps, logIndex: 1 },
          ),
        );
      const loggerWarnSpy = jest.spyOn(logger, "warn");

      await testMessageSentEventPoller.start();
      await wait(500);

      expect(loggerWarnSpy).toHaveBeenCalled();
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Something went wrong with database access. Restarting fromBlockNum=%s and fromLogIndex=%s and errorMessage=%s",
        0,
        1,
        new DatabaseAccessError(
          DatabaseRepoName.MessageRepository,
          DatabaseErrorType.Read,
          new Error("read database error."),
        ).message,
      );
      expect(l1QuerierMockSpy).toHaveBeenCalled();
      expect(messageRepositoryMockSpy).toHaveBeenCalled();
      expect(messageRepositoryMockSpy).toHaveBeenCalledWith(
        Direction.L1_TO_L2,
        testL2NetworkConfig.messageServiceContractAddress,
      );
      expect(eventProcessorMockSpy).toHaveBeenCalled();
      expect(eventProcessorMockSpy).toHaveBeenCalledWith(10, 0);
    });

    it("Should log as warning or error if getAndStoreMessageSentEvents throws Error", async () => {
      const l1QuerierMockSpy = jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockResolvedValue(10);
      const messageRepositoryMockSpy = jest
        .spyOn(messageRepositoryMock, "getLatestMessageSent")
        .mockResolvedValue(testMessage);
      const error = new Error("Other error for testing");
      const eventProcessorMockSpy = jest
        .spyOn(eventProcessorMock, "getAndStoreMessageSentEvents")
        .mockRejectedValue(error);
      const loggerWarnOrErrorSpy = jest.spyOn(logger, "warnOrError");

      await testMessageSentEventPoller.start();
      await wait(500);

      expect(loggerWarnOrErrorSpy).toHaveBeenCalled();
      expect(loggerWarnOrErrorSpy).toHaveBeenCalledWith(error);
      expect(l1QuerierMockSpy).toHaveBeenCalled();
      expect(messageRepositoryMockSpy).toHaveBeenCalled();
      expect(messageRepositoryMockSpy).toHaveBeenCalledWith(
        Direction.L1_TO_L2,
        testL2NetworkConfig.messageServiceContractAddress,
      );
      expect(eventProcessorMockSpy).toHaveBeenCalled();
      expect(eventProcessorMockSpy).toHaveBeenCalledWith(10, 0);
    });
  });

  describe("stop", () => {
    it("Should return and log as info if it stopped successfully", async () => {
      const loggerInfoSpy = jest.spyOn(logger, "info");
      testMessageSentEventPoller = new MessageSentEventPoller(
        eventProcessorMock,
        l1QuerierMock,
        messageRepositoryMock,
        Direction.L1_TO_L2,
        {
          ...testL2NetworkConfig,
          listener: {},
        },
        logger,
      );

      testMessageSentEventPoller.start();
      testMessageSentEventPoller.stop();

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(
        1,
        "Starting %s %s...",
        Direction.L1_TO_L2,
        MessageSentEventPoller.name,
      );
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(
        2,
        "Stopping %s %s...",
        Direction.L1_TO_L2,
        MessageSentEventPoller.name,
      );
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(
        3,
        "%s %s stopped.",
        Direction.L1_TO_L2,
        MessageSentEventPoller.name,
      );
    });
  });
});
