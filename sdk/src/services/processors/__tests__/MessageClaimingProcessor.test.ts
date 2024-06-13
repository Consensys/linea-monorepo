import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { TestLogger } from "../../../utils/testing/helpers";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../../core/enums/MessageEnums";
import {
  TEST_L2_SIGNER_PRIVATE_KEY,
  testAnchoredMessage,
  testClaimedMessage,
  testL1NetworkConfig,
  testL2NetworkConfig,
  testUnderpricedAnchoredMessage,
  testZeroFeeAnchoredMessage,
} from "../../../utils/testing/constants";
import { IMessageRepository } from "../../../core/persistence/IMessageRepository";
import { IChainQuerier } from "../../../core/clients/blockchain/IChainQuerier";
import { IMessageClaimingProcessor } from "../../../core/services/processors/IMessageClaimingProcessor";
import { MessageClaimingProcessor } from "../MessageClaimingProcessor";
import { IMessageServiceContract } from "../../../core/services/contracts/IMessageServiceContract";
import { ContractTransactionResponse, EthersError, Overrides, TransactionReceipt, TransactionResponse } from "ethers";
import { IEIP1559GasProvider } from "../../../core/clients/blockchain/IEIP1559GasProvider";
import { Message } from "../../..";
import { ErrorParser } from "../../../utils/ErrorParser";

describe("TestMessageClaimingProcessor", () => {
  let messageClaimingProcessor: IMessageClaimingProcessor;
  let mockedDate: Date;
  const messageRepositoryMock = mock<IMessageRepository<unknown>>();
  const l2MessageServiceContractMock = mock<
    IMessageServiceContract<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse> &
      IEIP1559GasProvider
  >();
  const l2QuerierMock = mock<IChainQuerier<unknown>>();
  const logger = new TestLogger(MessageClaimingProcessor.name);

  beforeEach(() => {
    messageClaimingProcessor = new MessageClaimingProcessor(
      messageRepositoryMock,
      l2MessageServiceContractMock,
      l2QuerierMock,
      testL2NetworkConfig,
      Direction.L1_TO_L2,
      testL1NetworkConfig.messageServiceContractAddress,
      logger,
    );

    mockedDate = new Date();
    jest.useFakeTimers();
    jest.setSystemTime(mockedDate);
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("getAndClaimAnchoredMessage", () => {
    it("Should return and log as error if claim tx nonce is higher than the max diff", async () => {
      const loggerErrorSpy = jest.spyOn(logger, "error");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(80);

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenCalledWith(
        "Nonce returned from getNonce is an invalid value (e.g. null or undefined)",
      );
    });

    it("Should return without calling any get message status if getFirstMessageToClaim return null", async () => {
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(null);

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(0);
    });

    it("Should log as warning and save message as zero fee if message has zero fee", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      const messageRepositorySaveSpy = jest.spyOn(messageRepositoryMock, "updateMessage");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(testZeroFeeAnchoredMessage);
      const expectedLoggingMessage = new Message(testZeroFeeAnchoredMessage);
      const expectedSavedMessage = new Message({
        ...testZeroFeeAnchoredMessage,
        status: MessageStatus.ZERO_FEE,
        updatedAt: mockedDate,
      });

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(0);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Found message with zero fee. This message will not be processed: messageHash=%s",
        expectedLoggingMessage.messageHash,
      );
      expect(messageRepositorySaveSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledWith(expectedSavedMessage);
    });

    it("Should log as info and save message as claimed if message was claimed", async () => {
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      const messageRepositorySaveSpy = jest.spyOn(messageRepositoryMock, "updateMessage");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(testClaimedMessage);
      jest.spyOn(l2MessageServiceContractMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMED);
      const expectedLoggingMessage = new Message(testClaimedMessage);
      const expectedSavedMessage = new Message({ ...testClaimedMessage, updatedAt: mockedDate });

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith(
        "Found already claimed message: messageHash=%s",
        expectedLoggingMessage.messageHash,
      );
      expect(messageRepositorySaveSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledWith(expectedSavedMessage);
    });

    it("Should log as warning and save message as non-executable if message gas limit was above max gas limit", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      const messageRepositorySaveSpy = jest.spyOn(messageRepositoryMock, "updateMessage");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(testAnchoredMessage);
      jest.spyOn(l2MessageServiceContractMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(l2MessageServiceContractMock, "estimateClaimGas").mockResolvedValue(200000n);
      const expectedLoggingMessage = new Message(testAnchoredMessage);
      const expectedSavedMessage = new Message({
        ...testAnchoredMessage,
        status: MessageStatus.NON_EXECUTABLE,
        updatedAt: mockedDate,
      });

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Estimated gas limit is higher than the max allowed gas limit for this message: messageHash=%s messageInfo=%s estimatedGasLimit=%s maxAllowedGasLimit=%s",
        expectedLoggingMessage.messageHash,
        expectedLoggingMessage.toString(),
        "200000",
        testL2NetworkConfig.claiming.maxClaimGasLimit!.toString(),
      );
      expect(messageRepositorySaveSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledWith(expectedSavedMessage);
    });

    it("Should log as warning and save message as fee underpriced if message fee was underpriced", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      const messageRepositorySaveSpy = jest.spyOn(messageRepositoryMock, "updateMessage");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(testUnderpricedAnchoredMessage);
      jest.spyOn(l2MessageServiceContractMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(l2MessageServiceContractMock, "estimateClaimGas").mockResolvedValue(100000n);
      const expectedLoggingMessage = new Message({
        ...testUnderpricedAnchoredMessage,
        claimGasEstimationThreshold: 10,
        updatedAt: mockedDate,
      });
      const expectedSavedMessage = new Message({
        ...testUnderpricedAnchoredMessage,
        claimGasEstimationThreshold: 10,
        status: MessageStatus.FEE_UNDERPRICED,
        updatedAt: mockedDate,
      });

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Fee underpriced found in this message: messageHash=%s messageInfo=%s transactionGasLimit=%s maxFeePerGas=%s",
        expectedLoggingMessage.messageHash,
        expectedLoggingMessage.toString(),
        "100000",
        "1000000000",
      );
      expect(messageRepositorySaveSpy).toHaveBeenCalledTimes(2);
      //expect(messageRepositorySaveSpy).toHaveBeenNthCalledWith(1, [expectedLoggingMessage]);
      expect(messageRepositorySaveSpy).toHaveBeenNthCalledWith(2, expectedSavedMessage);
    });

    it("Should log as warning and save message with reset claimGasEstimationThreshold if rate limit exceeded on L1", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      const messageRepositorySaveSpy = jest.spyOn(messageRepositoryMock, "updateMessage");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(testAnchoredMessage);
      jest.spyOn(l2MessageServiceContractMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(l2MessageServiceContractMock, "estimateClaimGas").mockResolvedValue(100000n);
      jest.spyOn(l2MessageServiceContractMock, "isRateLimitExceeded").mockResolvedValue(true);
      const expectedLoggingMessage = new Message({
        ...testAnchoredMessage,
        claimGasEstimationThreshold: 10000000000,
        updatedAt: mockedDate,
      });

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledWith(expectedLoggingMessage);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Rate limit exceeded for this message. It will be reprocessed later: messageHash=%s",
        expectedLoggingMessage.messageHash,
      );
    });

    it("Should update message if successful", async () => {
      messageClaimingProcessor = new MessageClaimingProcessor(
        messageRepositoryMock,
        l2MessageServiceContractMock,
        l2QuerierMock,
        {
          ...testL2NetworkConfig,
          claiming: {
            signerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
            messageSubmissionTimeout: 300_000,
          },
        },
        Direction.L1_TO_L2,
        testL1NetworkConfig.messageServiceContractAddress,
        logger,
      );
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      const l2MessageServiceContractClaimSpy = jest.spyOn(l2MessageServiceContractMock, "claim");
      const messageRepositorySaveSpy = jest.spyOn(messageRepositoryMock, "updateMessage");
      const messageRepositoryUpdateAtomicSpy = jest.spyOn(messageRepositoryMock, "updateMessageWithClaimTxAtomic");
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(testAnchoredMessage);
      jest.spyOn(l2MessageServiceContractMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(l2MessageServiceContractMock, "estimateClaimGas").mockResolvedValue(100000n);
      jest.spyOn(l2MessageServiceContractMock, "isRateLimitExceeded").mockResolvedValue(false);
      const expectedLoggingMessage = new Message({
        ...testAnchoredMessage,
        claimGasEstimationThreshold: 10000000000,
        updatedAt: mockedDate,
      });

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledWith(expectedLoggingMessage);
      expect(l2MessageServiceContractClaimSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryUpdateAtomicSpy).toHaveBeenCalledTimes(1);
    });

    it("Should log as warn or error if update message throws a ACTION_REJECTED error", async () => {
      const loggerWarnOrErrorSpy = jest.spyOn(logger, "warnOrError");
      const l2MessageServiceContractMsgStatusSpy = jest.spyOn(l2MessageServiceContractMock, "getMessageStatus");
      const l2MessageServiceContractClaimSpy = jest.spyOn(l2MessageServiceContractMock, "claim");
      const messageRepositorySaveSpy = jest.spyOn(messageRepositoryMock, "updateMessage");
      const actionRejectedError = {
        code: "ACTION_REJECTED",
        shortMessage: "action rejected error for testing",
      };
      const messageRepositoryUpdateAtomicSpy = jest
        .spyOn(messageRepositoryMock, "updateMessageWithClaimTxAtomic")
        .mockRejectedValue(actionRejectedError);
      jest.spyOn(messageRepositoryMock, "getLastClaimTxNonce").mockResolvedValue(100);
      jest.spyOn(l2QuerierMock, "getCurrentNonce").mockResolvedValue(99);
      jest.spyOn(l2MessageServiceContractMock, "get1559Fees").mockResolvedValue({ maxFeePerGas: 1000000000n });
      jest.spyOn(messageRepositoryMock, "getFirstMessageToClaim").mockResolvedValue(testAnchoredMessage);
      jest.spyOn(l2MessageServiceContractMock, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(l2MessageServiceContractMock, "estimateClaimGas").mockResolvedValue(100000n);
      jest.spyOn(l2MessageServiceContractMock, "isRateLimitExceeded").mockResolvedValue(false);
      const expectedLoggingMessage = new Message({
        ...testAnchoredMessage,
        claimGasEstimationThreshold: 10000000000,
        updatedAt: mockedDate,
      });
      const expectedSavedMessage = new Message({
        ...testAnchoredMessage,
        claimGasEstimationThreshold: 10000000000,
        status: MessageStatus.NON_EXECUTABLE,
        updatedAt: mockedDate,
      });

      await messageClaimingProcessor.getAndClaimAnchoredMessage();

      expect(l2MessageServiceContractMsgStatusSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositorySaveSpy).toHaveBeenCalledTimes(2);
      expect(messageRepositorySaveSpy).toHaveBeenNthCalledWith(1, expectedLoggingMessage);
      expect(messageRepositorySaveSpy).toHaveBeenNthCalledWith(2, expectedSavedMessage);
      expect(l2MessageServiceContractClaimSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryUpdateAtomicSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnOrErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnOrErrorSpy).toHaveBeenCalledWith(actionRejectedError, {
        parsedError: ErrorParser.parseErrorWithMitigation(actionRejectedError as EthersError),
      });
    });
  });
});
