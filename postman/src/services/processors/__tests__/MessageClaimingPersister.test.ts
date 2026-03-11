import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "../../../../src/core/metrics";
import { IProvider } from "../../../core/clients/blockchain/IProvider";
import { Message } from "../../../core/entities/Message";
import { Direction, OnChainMessageStatus, MessageStatus } from "../../../core/enums";
import { IMessageRepository } from "../../../core/persistence/IMessageRepository";
import { IMessageServiceContract } from "../../../core/services/contracts/IMessageServiceContract";
import { IReceiptPoller } from "../../../core/services/IReceiptPoller";
import { ITransactionRetrier } from "../../../core/services/ITransactionRetrier";
import { IMessageClaimingPersister } from "../../../core/services/processors/IMessageClaimingPersister";
import { TransactionReceipt, TransactionSubmission } from "../../../core/types";
import {
  TEST_TRANSACTION_HASH,
  testL2NetworkConfig,
  testPendingMessage,
  testPendingMessage2,
} from "../../../utils/testing/constants";
import { TestLogger } from "../../../utils/testing/helpers";
import { MessageClaimingPersister } from "../MessageClaimingPersister";

const generateTransactionReceipt = (overrides: Partial<TransactionReceipt> = {}): TransactionReceipt => ({
  hash: TEST_TRANSACTION_HASH,
  blockNumber: 100,
  status: "success",
  gasUsed: 50_000n,
  gasPrice: 100_000_000_000n,
  logs: [],
  ...overrides,
});

const generateTransactionSubmission = (overrides: Partial<TransactionSubmission> = {}): TransactionSubmission => ({
  hash: TEST_TRANSACTION_HASH,
  nonce: 1,
  gasLimit: 50_000n,
  maxFeePerGas: 100_000_000_000n,
  maxPriorityFeePerGas: 1_000_000_000n,
  ...overrides,
});

describe("TestMessageClaimingPersister", () => {
  let messageClaimingPersister: IMessageClaimingPersister;
  let mockedDate: Date;
  const databaseService = mock<IMessageRepository>();
  const messageServiceContract = mock<IMessageServiceContract>();
  const sponsorshipMetricsUpdater = mock<ISponsorshipMetricsUpdater>();
  const transactionMetricsUpdater = mock<ITransactionMetricsUpdater>();
  const provider = mock<IProvider>();
  const transactionRetrier = mock<ITransactionRetrier>();
  const receiptPoller = mock<IReceiptPoller>();
  const logger = new TestLogger(MessageClaimingPersister.name);

  beforeEach(() => {
    messageClaimingPersister = new MessageClaimingPersister(
      databaseService,
      messageServiceContract,
      sponsorshipMetricsUpdater,
      transactionMetricsUpdater,
      provider,
      transactionRetrier,
      receiptPoller,
      {
        direction: Direction.L1_TO_L2,
        messageSubmissionTimeout: testL2NetworkConfig.claiming.messageSubmissionTimeout,
        maxBumpsPerCycle: testL2NetworkConfig.claiming.maxBumpsPerCycle,
        maxCycles: testL2NetworkConfig.claiming.maxRetryCycles,
        receiptPollingTimeout: 120_000,
        receiptPollingInterval: 0,
      },
      logger,
    );

    mockedDate = new Date();
    jest.useFakeTimers();
    jest.setSystemTime(mockedDate);
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("process", () => {
    it("Should early return immediately if no pending message found", async () => {
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(null);
      const getReceiptSpy = jest.spyOn(provider, "getTransactionReceipt");

      await messageClaimingPersister.process();

      expect(getReceiptSpy).toHaveBeenCalledTimes(0);
    });

    it("Should log as error if getTransactionReceipt throws error", async () => {
      const getTxReceiptError = new Error("error for testing");
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(new Message(testPendingMessage));
      jest.spyOn(provider, "getTransactionReceipt").mockRejectedValue(getTxReceiptError);
      const loggerErrorSpy = jest.spyOn(logger, "error");

      await messageClaimingPersister.process();

      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenCalledWith(getTxReceiptError, {
        messageHash: testPendingMessage.messageHash,
      });
    });

    it("Should log as info and update message as claimed success if successful", async () => {
      const txReceipt = generateTransactionReceipt({ status: "success" });
      const testPendingMessageLocal = new Message(testPendingMessage);
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(testPendingMessageLocal);
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(txReceipt);
      jest.spyOn(sponsorshipMetricsUpdater, "incrementSponsorshipFeePaid").mockResolvedValue();
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(messageRepositoryUpdateSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith("Message has been SUCCESSFULLY claimed.", {
        messageHash: testPendingMessageLocal.messageHash,
        transactionHash: testPendingMessageLocal.claimTxHash,
      });
    });

    it("Should return and update message as sent if receipt status is reverted and rate limit exceeded", async () => {
      const txReceipt = generateTransactionReceipt({ status: "reverted" });
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(new Message(testPendingMessage));
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(txReceipt);
      jest.spyOn(messageServiceContract, "isRateLimitExceededError").mockResolvedValue(true);
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(messageRepositoryUpdateSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryUpdateSpy).toHaveBeenCalledWith(expect.objectContaining({ status: MessageStatus.SENT }));
    });

    it("Should log as warning and update message as claim reverted if receipt status is reverted", async () => {
      const txReceipt = generateTransactionReceipt({ status: "reverted" });
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(new Message(testPendingMessage));
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(txReceipt);
      jest.spyOn(messageServiceContract, "isRateLimitExceededError").mockResolvedValue(false);
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(messageRepositoryUpdateSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryUpdateSpy).toHaveBeenCalledWith(
        expect.objectContaining({ status: MessageStatus.CLAIMED_REVERTED }),
      );
      expect(loggerWarnSpy).toHaveBeenCalledWith("Message claim transaction has been REVERTED.", expect.any(Object));
    });

    it("Should update message as claimed if message claimed on-chain and receipt found on retry", async () => {
      const retryTxReceipt = generateTransactionReceipt({ status: "success" });
      const testPendingMessageLocal = new Message(testPendingMessage);
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(testPendingMessageLocal);
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValueOnce(null).mockResolvedValueOnce(retryTxReceipt);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMED);
      jest.spyOn(sponsorshipMetricsUpdater, "incrementSponsorshipFeePaid").mockResolvedValue();
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(messageRepositoryUpdateSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryUpdateSpy).toHaveBeenCalledWith(
        expect.objectContaining({ status: MessageStatus.CLAIMED_SUCCESS }),
      );
      expect(loggerWarnSpy).toHaveBeenCalledWith("Retrying to claim message.", expect.any(Object));
      expect(loggerWarnSpy).toHaveBeenCalledWith("Retried claim message transaction succeed.", expect.any(Object));
    });

    it("Should return and log as warning if message is claimed but receipt returned as null", async () => {
      const testPendingMessageLocal = new Message(testPendingMessage);
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(testPendingMessageLocal);
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(null);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMED);
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(messageRepositoryUpdateSpy).toHaveBeenCalledTimes(0);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Message was claimed on-chain but transaction receipt is not available yet.",
        expect.any(Object),
      );
    });

    it("Should bump fee and update DB when message is claimable and retry tx succeeds", async () => {
      const retryTxReceipt = generateTransactionReceipt({ status: "success" });
      const retryTxResponse = generateTransactionSubmission({
        maxPriorityFeePerGas: undefined,
        maxFeePerGas: undefined,
      });
      const testPendingMessageLocal = new Message(testPendingMessage);
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(testPendingMessageLocal);
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(null);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(transactionRetrier, "retryWithHigherFee").mockResolvedValue(retryTxResponse);
      jest.spyOn(receiptPoller, "poll").mockResolvedValue(retryTxReceipt);
      jest.spyOn(sponsorshipMetricsUpdater, "incrementSponsorshipFeePaid").mockResolvedValue();
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");
      const loggerWarnSpy = jest.spyOn(logger, "warn");

      await messageClaimingPersister.process();

      // First call: update after retry tx, second call: update receipt status
      expect(messageRepositoryUpdateSpy).toHaveBeenCalledTimes(2);
      expect(messageRepositoryUpdateSpy).toHaveBeenNthCalledWith(1, expect.objectContaining({ claimNumberOfRetry: 1 }));
      expect(loggerWarnSpy).toHaveBeenCalledWith("Bumping fee for claim transaction.", expect.any(Object));
      expect(loggerWarnSpy).toHaveBeenCalledWith("Retried claim message transaction succeed.", expect.any(Object));
    });

    it("Should return null when retry tx throws error", async () => {
      messageClaimingPersister = new MessageClaimingPersister(
        databaseService,
        messageServiceContract,
        sponsorshipMetricsUpdater,
        transactionMetricsUpdater,
        provider,
        transactionRetrier,
        receiptPoller,
        {
          direction: Direction.L1_TO_L2,
          messageSubmissionTimeout: 0,
          maxBumpsPerCycle: 5,
          maxCycles: 2,
          receiptPollingTimeout: 120_000,
          receiptPollingInterval: 0,
        },
        logger,
      );

      const retryError = new Error("error for testing");
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(new Message(testPendingMessage));
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(null);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(transactionRetrier, "retryWithHigherFee").mockRejectedValue(retryError);
      const loggerErrorSpy = jest.spyOn(logger, "error");

      await messageClaimingPersister.process();

      expect(loggerErrorSpy).toHaveBeenCalledWith(retryError, { messageHash: testPendingMessage.messageHash });
    });

    it("Should cancel and reset message when max bumps per cycle exceeded", async () => {
      const pendingMsg = new Message({
        ...testPendingMessage,
        claimNumberOfRetry: 5,
        claimCycleCount: 0,
        claimTxNonce: 42,
      });
      messageClaimingPersister = new MessageClaimingPersister(
        databaseService,
        messageServiceContract,
        sponsorshipMetricsUpdater,
        transactionMetricsUpdater,
        provider,
        transactionRetrier,
        receiptPoller,
        {
          direction: Direction.L1_TO_L2,
          messageSubmissionTimeout: 0,
          maxBumpsPerCycle: 5,
          maxCycles: 2,
          receiptPollingTimeout: 120_000,
          receiptPollingInterval: 0,
        },
        logger,
      );

      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(pendingMsg);
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(null);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValue(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(transactionRetrier, "cancelTransaction").mockResolvedValue("0xcancelhash" as `0x${string}`);
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(transactionRetrier.cancelTransaction).toHaveBeenCalledWith(42);
      expect(messageRepositoryUpdateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          status: MessageStatus.SENT,
          claimNumberOfRetry: 0,
          claimCycleCount: 1,
        }),
      );
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Max fee bumps exhausted, cancelling stuck transaction and resetting message.",
        expect.any(Object),
      );
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Message reset to SENT for re-claiming with fresh nonce.",
        expect.any(Object),
      );
    });

    it("Should set NEEDS_MANUAL_INTERVENTION when max cycles exceeded", async () => {
      const pendingMsg = new Message({
        ...testPendingMessage,
        claimNumberOfRetry: 5,
        claimCycleCount: 2,
      });
      messageClaimingPersister = new MessageClaimingPersister(
        databaseService,
        messageServiceContract,
        sponsorshipMetricsUpdater,
        transactionMetricsUpdater,
        provider,
        transactionRetrier,
        receiptPoller,
        {
          direction: Direction.L1_TO_L2,
          messageSubmissionTimeout: 0,
          maxBumpsPerCycle: 5,
          maxCycles: 2,
          receiptPollingTimeout: 120_000,
          receiptPollingInterval: 0,
        },
        logger,
      );

      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(pendingMsg);
      jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(null);
      const loggerErrorSpy = jest.spyOn(logger, "error");
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(messageRepositoryUpdateSpy).toHaveBeenCalledWith(
        expect.objectContaining({ status: MessageStatus.NEEDS_MANUAL_INTERVENTION }),
      );
      expect(loggerErrorSpy).toHaveBeenCalledWith(
        "Max retry cycles exceeded. Manual intervention is needed.",
        expect.any(Object),
      );
    });

    it("Should reset pending message without claim tx hash to SENT", async () => {
      const pendingMsg = new Message({
        ...testPendingMessage,
        claimTxHash: undefined,
      });
      jest.spyOn(databaseService, "getFirstPendingMessage").mockResolvedValue(pendingMsg);
      const loggerWarnSpy = jest.spyOn(logger, "warn");
      const messageRepositoryUpdateSpy = jest.spyOn(databaseService, "updateMessage");

      await messageClaimingPersister.process();

      expect(messageRepositoryUpdateSpy).toHaveBeenCalledWith(expect.objectContaining({ status: MessageStatus.SENT }));
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        "Found pending message without claim tx hash, resetting to allow retry.",
        expect.any(Object),
      );
    });

    it("Should handle two consecutive process calls correctly", async () => {
      const retryTxReceipt = generateTransactionReceipt({ status: "success" });
      const retryTxResponse = generateTransactionSubmission({
        maxPriorityFeePerGas: undefined,
        maxFeePerGas: undefined,
      });
      const testPendingMessageLocal = new Message(testPendingMessage);
      const testPendingMessageLocal2 = new Message(testPendingMessage2);

      jest
        .spyOn(databaseService, "getFirstPendingMessage")
        .mockResolvedValueOnce(testPendingMessageLocal)
        .mockResolvedValueOnce(testPendingMessageLocal2);
      jest
        .spyOn(provider, "getTransactionReceipt")
        .mockResolvedValueOnce(null) // first call: no receipt
        .mockResolvedValueOnce(null) // retryWithBump: getMessageStatus path (CLAIMED check)
        .mockResolvedValueOnce(null) // second call: no receipt
        .mockResolvedValueOnce(null); // second call: CLAIMED check
      jest
        .spyOn(messageServiceContract, "getMessageStatus")
        .mockResolvedValueOnce(OnChainMessageStatus.CLAIMABLE)
        .mockResolvedValueOnce(OnChainMessageStatus.CLAIMED);
      jest.spyOn(transactionRetrier, "retryWithHigherFee").mockResolvedValue(retryTxResponse);
      jest.spyOn(receiptPoller, "poll").mockResolvedValue(retryTxReceipt);
      jest.spyOn(sponsorshipMetricsUpdater, "incrementSponsorshipFeePaid").mockResolvedValue();

      await messageClaimingPersister.process();
      await messageClaimingPersister.process();

      // First process: bump + receipt update, second process: claimed on-chain receipt
      expect(databaseService.updateMessage).toHaveBeenCalled();
    });
  });
});
