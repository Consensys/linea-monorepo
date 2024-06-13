import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { TestLogger } from "../../../utils/testing/helpers";
import { IMessageAnchoringProcessor } from "../../../core/services/processors/IMessageAnchoringProcessor";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../../core/enums/MessageEnums";
import { testL1NetworkConfig, testL2NetworkConfig, testMessage } from "../../../utils/testing/constants";
import { MessageAnchoringProcessor } from "../MessageAnchoringProcessor";
import { IMessageRepository } from "../../../core/persistence/IMessageRepository";
import { IMessageServiceContract } from "../../../core/services/contracts/IMessageServiceContract";
import { IChainQuerier } from "../../../core/clients/blockchain/IChainQuerier";
import { ContractTransactionResponse, Overrides, TransactionReceipt, TransactionResponse } from "ethers";

describe("TestMessageAnchoringProcessor", () => {
  let anchoringProcessor: IMessageAnchoringProcessor;
  const messageRepositoryMock = mock<IMessageRepository<unknown>>();
  const l2ContractClientMock =
    mock<IMessageServiceContract<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse>>();
  const l2QuerierMock = mock<IChainQuerier<unknown>>();
  const logger = new TestLogger(MessageAnchoringProcessor.name);

  beforeEach(() => {
    anchoringProcessor = new MessageAnchoringProcessor(
      messageRepositoryMock,
      l2ContractClientMock,
      l2QuerierMock,
      testL2NetworkConfig,
      Direction.L1_TO_L2,
      testL1NetworkConfig.messageServiceContractAddress,
      logger,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("getAndUpdateAnchoredMessageStatus", () => {
    it("Should return if getNFirstMessageSent returns empty list", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      jest.spyOn(messageRepositoryMock, "getNFirstMessageSent").mockResolvedValue([]);

      await anchoringProcessor.getAndUpdateAnchoredMessageStatus();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(0);
    });

    it("Should log as warning if getNFirstMessageSent returns a list longer than maxFetchMessagesFromDb", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const maxFetchMessagesFromDb = testL2NetworkConfig.listener.maxFetchMessagesFromDb;
      jest
        .spyOn(messageRepositoryMock, "getNFirstMessageSent")
        .mockResolvedValue(Array(maxFetchMessagesFromDb).map(() => testMessage));

      await anchoringProcessor.getAndUpdateAnchoredMessageStatus();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Limit of messages sent to listen reached (%s).",
        maxFetchMessagesFromDb,
      );
    });

    it("Should log as info and call saveMessages if returned messageStatus is CLAIMABLE", async () => {
      const loggerInfoSpy = jest.spyOn(logger, "info");
      jest.spyOn(messageRepositoryMock, "getNFirstMessageSent").mockResolvedValue([testMessage]);
      jest.spyOn(l2QuerierMock, "getCurrentBlockNumber").mockResolvedValue(10);
      jest.spyOn(l2ContractClientMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      const testMessageEditSpy = jest.spyOn(testMessage, "edit");
      const messageRepositoryMockSaveSpy = jest.spyOn(messageRepositoryMock, "saveMessages");

      await anchoringProcessor.getAndUpdateAnchoredMessageStatus();

      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith("Message has been anchored: messageHash=%s", testMessage.messageHash);
      expect(testMessageEditSpy).toHaveBeenCalledTimes(1);
      expect(testMessageEditSpy).toHaveBeenCalledWith({ status: MessageStatus.ANCHORED });
      expect(messageRepositoryMockSaveSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryMockSaveSpy).toHaveBeenCalledWith([testMessage]);
    });

    it("Should log as info and call saveMessages if returned messageStatus is CLAIMED", async () => {
      const loggerInfoSpy = jest.spyOn(logger, "info");
      jest.spyOn(messageRepositoryMock, "getNFirstMessageSent").mockResolvedValue([testMessage]);
      jest.spyOn(l2QuerierMock, "getCurrentBlockNumber").mockResolvedValue(10);
      jest.spyOn(l2ContractClientMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMED);
      const testMessageEditSpy = jest.spyOn(testMessage, "edit");
      const messageRepositoryMockSaveSpy = jest.spyOn(messageRepositoryMock, "saveMessages");

      await anchoringProcessor.getAndUpdateAnchoredMessageStatus();

      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith(
        "Message has already been claimed: messageHash=%s",
        testMessage.messageHash,
      );
      expect(testMessageEditSpy).toHaveBeenCalledTimes(1);
      expect(testMessageEditSpy).toHaveBeenCalledWith({ status: MessageStatus.CLAIMED_SUCCESS });
      expect(messageRepositoryMockSaveSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryMockSaveSpy).toHaveBeenCalledWith([testMessage]);
    });

    it("Should log as error if getCurrentBlockNumber throws error", async () => {
      anchoringProcessor = new MessageAnchoringProcessor(
        messageRepositoryMock,
        l2ContractClientMock,
        l2QuerierMock,
        {
          ...testL2NetworkConfig,
          listener: {},
        },
        Direction.L1_TO_L2,
        testL1NetworkConfig.messageServiceContractAddress,
        logger,
      );
      const loggerErrorSpy = jest.spyOn(logger, "error");
      const error = new Error("Error for testing");
      jest.spyOn(messageRepositoryMock, "getNFirstMessageSent").mockResolvedValue([testMessage]);
      jest.spyOn(l2QuerierMock, "getCurrentBlockNumber").mockRejectedValue(error);
      const messageRepositoryMockSaveSpy = jest.spyOn(messageRepositoryMock, "saveMessages");

      await anchoringProcessor.getAndUpdateAnchoredMessageStatus();

      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenCalledWith(error);
      expect(messageRepositoryMockSaveSpy).toHaveBeenCalledTimes(0);
    });
  });
});
