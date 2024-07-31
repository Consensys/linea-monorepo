import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { TestLogger } from "../../../utils/testing/helpers";
import { Direction, MessageStatus } from "../../../core/enums/MessageEnums";
import {
  testL1NetworkConfig,
  testMessageSentEvent,
  testMessageSentEventWithCallData,
} from "../../../utils/testing/constants";
import { IChainQuerier } from "../../../core/clients/blockchain/IChainQuerier";
import { IMessageSentEventProcessor } from "../../../core/services/processors/IMessageSentEventProcessor";
import { MessageSentEventProcessor } from "../MessageSentEventProcessor";
import { ILineaRollupLogClient } from "../../../core/clients/blockchain/ethereum/ILineaRollupLogClient";
import { MessageFactory } from "../../../core/entities/MessageFactory";
import { Block, JsonRpcProvider, TransactionReceipt, TransactionRequest, TransactionResponse } from "ethers";
import { EthereumMessageDBService } from "../../persistence/EthereumMessageDBService";

describe("TestMessageSentEventProcessor", () => {
  let messageSentEventProcessor: IMessageSentEventProcessor;
  const databaseService = mock<EthereumMessageDBService>();
  const l1LogClientMock = mock<ILineaRollupLogClient>();
  const l1QuerierMock =
    mock<IChainQuerier<TransactionReceipt, Block, TransactionRequest, TransactionResponse, JsonRpcProvider>>();
  const logger = new TestLogger(MessageSentEventProcessor.name);

  beforeEach(() => {
    messageSentEventProcessor = new MessageSentEventProcessor(
      databaseService,
      l1LogClientMock,
      l1QuerierMock,
      {
        direction: Direction.L1_TO_L2,
        maxBlocksToFetchLogs: testL1NetworkConfig.listener.maxBlocksToFetchLogs,
        blockConfirmation: testL1NetworkConfig.listener.blockConfirmation,
        isEOAEnabled: testL1NetworkConfig.isEOAEnabled,
        isCalldataEnabled: testL1NetworkConfig.isCalldataEnabled,
      },
      logger,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("process", () => {
    it("Should insert message with status as sent into repository if the message is not excluded", async () => {
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryInsertSpy = jest.spyOn(databaseService, "insertMessage");
      jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockResolvedValue(100);
      jest.spyOn(l1LogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      const expectedMessageToInsert = MessageFactory.createMessage({
        ...testMessageSentEvent,
        sentBlockNumber: testMessageSentEvent.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
      });

      await messageSentEventProcessor.process(200, 0);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledWith(expectedMessageToInsert);
    });

    it("Should insert message with status as excluded into repository if the message is excluded", async () => {
      messageSentEventProcessor = new MessageSentEventProcessor(
        databaseService,
        l1LogClientMock,
        l1QuerierMock,
        {
          direction: Direction.L1_TO_L2,
          maxBlocksToFetchLogs: testL1NetworkConfig.listener.maxBlocksToFetchLogs,
          blockConfirmation: testL1NetworkConfig.listener.blockConfirmation,
          isEOAEnabled: !testL1NetworkConfig.isEOAEnabled,
          isCalldataEnabled: testL1NetworkConfig.isCalldataEnabled,
        },
        logger,
      );
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryInsertSpy = jest.spyOn(databaseService, "insertMessage");
      jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockResolvedValue(100);
      jest.spyOn(l1LogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      const expectedMessageToInsert = MessageFactory.createMessage({
        ...testMessageSentEvent,
        sentBlockNumber: testMessageSentEvent.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.EXCLUDED,
        claimNumberOfRetry: 0,
      });

      await messageSentEventProcessor.process(0, 0);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledWith(expectedMessageToInsert);
    });

    it("Should insert message with calldata with status as sent into repository if calldata is enabled", async () => {
      messageSentEventProcessor = new MessageSentEventProcessor(
        databaseService,
        l1LogClientMock,
        l1QuerierMock,
        {
          direction: Direction.L1_TO_L2,
          maxBlocksToFetchLogs: testL1NetworkConfig.listener.maxBlocksToFetchLogs,
          blockConfirmation: testL1NetworkConfig.listener.blockConfirmation,
          isEOAEnabled: testL1NetworkConfig.isEOAEnabled,
          isCalldataEnabled: !testL1NetworkConfig.isCalldataEnabled,
        },
        logger,
      );
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryInsertSpy = jest.spyOn(databaseService, "insertMessage");
      jest.spyOn(l1QuerierMock, "getCurrentBlockNumber").mockResolvedValue(100);
      jest.spyOn(l1LogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEventWithCallData]);
      const expectedMessageToInsert = MessageFactory.createMessage({
        ...testMessageSentEventWithCallData,
        sentBlockNumber: testMessageSentEventWithCallData.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
      });

      await messageSentEventProcessor.process(0, 0);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledWith(expectedMessageToInsert);
    });
  });
});
