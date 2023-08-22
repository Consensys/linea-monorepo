import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { BigNumber } from "ethers";
import { JsonRpcProvider, TransactionReceipt } from "@ethersproject/providers";
import { LoggerOptions } from "winston";
import { L2MessageServiceContract } from "../../../contracts";
import { getTestL2Signer } from "../../../utils/testHelpers/contracts";
import { TEST_CONTRACT_ADDRESS_2, testL2NetworkConfig } from "../../../utils/testHelpers/constants";
import { ClaimStatusWatcher } from "../ClaimStatusWatcher";
import { Direction, MessageStatus } from "../../utils/enums";
import {
  generateMessageFromDb,
  generateTransactionReceipt,
  generateTransactionResponse,
} from "../../../utils/testHelpers/helpers";
import { LineaLogger, getLogger } from "../../../logger";
import { L2NetworkConfig } from "../../utils/types";
import { MessageRepository } from "../../repositories/MessageRepository";

class TestClaimStatusWatcher extends ClaimStatusWatcher<L2MessageServiceContract> {
  public logger: LineaLogger;
  public repository: MessageRepository;

  constructor(
    dataSource: DataSource,
    l2MessageServiceContract: L2MessageServiceContract,
    config: L2NetworkConfig,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l2MessageServiceContract, config, Direction.L1_TO_L2);
    this.logger = getLogger("ClaimStatusWatcher", loggerOptions);
    this.repository = this.messageRepository;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public async isRateLimitExceededError(_transactionHash: string): Promise<boolean> {
    return false;
  }

  protected async getFeesUsingOriginLayerBaseFeePerGas(): Promise<{
    maxFeePerGas?: BigNumber | undefined;
    maxPriorityFeePerGas?: BigNumber | undefined;
  }> {
    return {};
  }

  public async waitForReceipt(interval: number) {
    await super.waitForReceipt(interval);
  }

  public async updateReceiptStatus(receipt: TransactionReceipt) {
    await super.updateReceiptStatus(receipt);
  }
}

const pollingInterval = 1_000;

describe("ClaimStatusWatcher", () => {
  let claimStatusWatcher: TestClaimStatusWatcher;
  let messageServiceContract: L2MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L2MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_2,
      "read-write",
      getTestL2Signer(),
    );

    claimStatusWatcher = new TestClaimStatusWatcher(mock<DataSource>(), messageServiceContract, testL2NetworkConfig, {
      silent: true,
    });
  });

  describe("updateReceiptStatus", () => {
    it("should update message status with SENT when transaction receipt status === 0 and error === RateLimitExceeded", async () => {
      jest.spyOn(claimStatusWatcher, "isRateLimitExceededError").mockResolvedValueOnce(true);
      const repositorySpy = jest
        .spyOn(claimStatusWatcher.repository, "updateMessageByTransactionHash")
        .mockResolvedValueOnce();

      const transactionReceipt = generateTransactionReceipt({ status: 0 });
      await claimStatusWatcher.updateReceiptStatus(transactionReceipt);

      expect(repositorySpy).toHaveBeenCalledTimes(1);
      expect(repositorySpy).toHaveBeenCalledWith(transactionReceipt.transactionHash, Direction.L1_TO_L2, {
        status: MessageStatus.SENT,
        claimGasEstimationThreshold: undefined,
      });
    });

    it("should update message status with CLAIMED_REVERTED when transaction receipt status === 0 and error !== RateLimitExceeded", async () => {
      const loggerWarnSpy = jest.spyOn(claimStatusWatcher.logger, "warn");
      const repositorySpy = jest
        .spyOn(claimStatusWatcher.repository, "updateMessageByTransactionHash")
        .mockResolvedValueOnce();

      const transactionReceipt = generateTransactionReceipt({ status: 0 });
      await claimStatusWatcher.updateReceiptStatus(transactionReceipt);

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        `CLAIMED_REVERTED: Message with tx hash ${transactionReceipt.transactionHash} has been reverted.`,
      );

      expect(repositorySpy).toHaveBeenCalledTimes(1);
      expect(repositorySpy).toHaveBeenCalledWith(transactionReceipt.transactionHash, Direction.L1_TO_L2, {
        status: MessageStatus.CLAIMED_REVERTED,
      });
    });

    it("should update message status with CLAIM_SUCCESS when transaction receipt status === 1", async () => {
      const loggerInfoSpy = jest.spyOn(claimStatusWatcher.logger, "info");
      const repositorySpy = jest
        .spyOn(claimStatusWatcher.repository, "updateMessageByTransactionHash")
        .mockResolvedValueOnce();

      const transactionReceipt = generateTransactionReceipt();
      await claimStatusWatcher.updateReceiptStatus(transactionReceipt);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith(
        `CLAIMED_SUCCESS: Message with tx hash ${transactionReceipt.transactionHash} has been claimed.`,
      );

      expect(repositorySpy).toHaveBeenCalledTimes(1);
      expect(repositorySpy).toHaveBeenCalledWith(transactionReceipt.transactionHash, Direction.L1_TO_L2, {
        status: MessageStatus.CLAIMED_SUCCESS,
      });
    });
  });

  describe("waitForReceipt", () => {
    it("should do nothing when there is no pending messages", async () => {
      const loggerInfoSpy = jest.spyOn(claimStatusWatcher.logger, "info");
      const loggerWarnSpy = jest.spyOn(claimStatusWatcher.logger, "warn");
      const loggerErrorSpy = jest.spyOn(claimStatusWatcher.logger, "error");

      const repositorySpy = jest
        .spyOn(claimStatusWatcher.repository, "getFirstPendingMessage")
        .mockResolvedValueOnce(null);
      await claimStatusWatcher.waitForReceipt(pollingInterval);

      expect(repositorySpy).toHaveBeenCalledTimes(1);
      expect(repositorySpy).toHaveBeenCalledWith(Direction.L1_TO_L2);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(0);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(0);
      expect(loggerErrorSpy).toHaveBeenCalledTimes(0);
    });

    it("should retry claim transaction when transaction has reached submission timeout and there is no transaction receipt", async () => {
      const loggerWarnSpy = jest.spyOn(claimStatusWatcher.logger, "warn");
      const messageFromDB = generateMessageFromDb({
        updatedAt: new Date("2023-06-02"),
        status: MessageStatus.PENDING,
        claimTxHash: "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
      });
      const transactionResponse = generateTransactionResponse();

      const getFirstPendingMessageSpy = jest
        .spyOn(claimStatusWatcher.repository, "getFirstPendingMessage")
        .mockResolvedValueOnce(messageFromDB);
      const updateMessageSpy = jest.spyOn(claimStatusWatcher.repository, "updateMessage").mockResolvedValueOnce();

      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      jest.spyOn(messageServiceContract.provider, "getTransactionReceipt").mockResolvedValueOnce(undefined);
      jest.spyOn(messageServiceContract, "retryTransactionWithHigherFee").mockResolvedValueOnce(transactionResponse);
      jest
        .spyOn(messageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });

      await claimStatusWatcher.waitForReceipt(pollingInterval);

      expect(getFirstPendingMessageSpy).toHaveBeenCalledTimes(1);
      expect(getFirstPendingMessageSpy).toHaveBeenCalledWith(Direction.L1_TO_L2);

      expect(updateMessageSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledWith(messageFromDB.messageHash, Direction.L1_TO_L2, {
        claimTxGasLimit: transactionResponse.gasLimit.toNumber(),
        claimTxMaxFeePerGas: transactionResponse.maxFeePerGas?.toBigInt(),
        claimTxMaxPriorityFeePerGas: transactionResponse.maxPriorityFeePerGas?.toBigInt(),
        claimTxHash: transactionResponse.hash,
      });

      expect(loggerWarnSpy).toHaveBeenCalledTimes(2);
      expect(loggerWarnSpy).toHaveBeenNthCalledWith(1, `Retring to claim:\nmessage: ${JSON.stringify(messageFromDB)}`);
      expect(loggerWarnSpy).toHaveBeenNthCalledWith(
        2,
        `Retried to claim done:\nmessage: ${JSON.stringify(messageFromDB)}\ntx: ${JSON.stringify(transactionResponse)}`,
      );
    });

    it("should catch the error and update the message status to NON_EXECUTABLE when the retry claim transaction failed", async () => {
      const loggerErrorSpy = jest.spyOn(claimStatusWatcher.logger, "error");
      const messageFromDB = generateMessageFromDb({
        updatedAt: new Date("2023-06-02"),
        status: MessageStatus.PENDING,
        claimTxHash: "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
      });

      const getFirstPendingMessageSpy = jest
        .spyOn(claimStatusWatcher.repository, "getFirstPendingMessage")
        .mockResolvedValueOnce(messageFromDB);
      const updateMessageSpy = jest.spyOn(claimStatusWatcher.repository, "updateMessage").mockResolvedValueOnce();

      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      jest.spyOn(messageServiceContract.provider, "getTransactionReceipt").mockResolvedValueOnce(undefined);
      jest
        .spyOn(messageServiceContract, "retryTransactionWithHigherFee")
        .mockRejectedValueOnce(new Error("Retry failed"));
      jest
        .spyOn(messageServiceContract, "get1559Fees")
        .mockResolvedValueOnce({ maxFeePerGas: BigNumber.from(100_000_000) });

      await claimStatusWatcher.waitForReceipt(pollingInterval);

      expect(getFirstPendingMessageSpy).toHaveBeenCalledTimes(1);
      expect(getFirstPendingMessageSpy).toHaveBeenCalledWith(Direction.L1_TO_L2);

      expect(updateMessageSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledWith(messageFromDB.messageHash, Direction.L1_TO_L2, {
        status: MessageStatus.NON_EXECUTABLE,
      });

      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenCalledWith(
        `Error found in retryTransactionWithHigherFee:\nFailed message: ${JSON.stringify(
          messageFromDB,
        )}\nFounded error: ${JSON.stringify(new Error("Retry failed"))}`,
      );
    });

    it("should catch the error and retry message when something went wrong (http error etc)", async () => {
      const loggerErrorSpy = jest.spyOn(claimStatusWatcher.logger, "error");

      const getFirstPendingMessageSpy = jest
        .spyOn(claimStatusWatcher.repository, "getFirstPendingMessage")
        .mockRejectedValueOnce(new Error("Http error"));

      const updateMessageSpy = jest.spyOn(claimStatusWatcher.repository, "updateMessage").mockResolvedValueOnce();

      await claimStatusWatcher.waitForReceipt(pollingInterval);

      expect(getFirstPendingMessageSpy).toHaveBeenCalledTimes(1);
      expect(getFirstPendingMessageSpy).toHaveBeenCalledWith(Direction.L1_TO_L2);

      expect(updateMessageSpy).toHaveBeenCalledTimes(0);

      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenCalledWith(
        `Error found in waitForReceipt:\nFailed message: ${JSON.stringify(null)}\nFounded error: ${JSON.stringify(
          new Error("Http error"),
        )}\nParsed error: ${JSON.stringify({
          errorCode: "UNKNOWN_ERROR",
          mitigation: { shouldRetry: true, retryWithBlocking: true, retryPeriodInMs: 5000 },
        })}`,
      );
    });

    it("should update message status when there is a receipt", async () => {
      const loggerInfoSpy = jest.spyOn(claimStatusWatcher.logger, "info");
      const messageFromDB = generateMessageFromDb({
        status: MessageStatus.PENDING,
        claimTxHash: "0x825a7f1aa4453735597ddf7e9062413c906a7ad49bf17ff32c2cf42f41d438d9",
      });

      const getFirstPendingMessageSpy = jest
        .spyOn(claimStatusWatcher.repository, "getFirstPendingMessage")
        .mockResolvedValueOnce(messageFromDB);
      const updateMessageByTransactionHashSpy = jest
        .spyOn(claimStatusWatcher.repository, "updateMessageByTransactionHash")
        .mockResolvedValueOnce();

      const transactionReceipt = generateTransactionReceipt();
      jest.spyOn(messageServiceContract.provider, "getTransactionReceipt").mockResolvedValueOnce(transactionReceipt);

      await claimStatusWatcher.waitForReceipt(pollingInterval);

      expect(getFirstPendingMessageSpy).toHaveBeenCalledTimes(1);
      expect(getFirstPendingMessageSpy).toHaveBeenCalledWith(Direction.L1_TO_L2);

      expect(updateMessageByTransactionHashSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageByTransactionHashSpy).toHaveBeenCalledWith(
        transactionReceipt.transactionHash,
        Direction.L1_TO_L2,
        {
          status: MessageStatus.CLAIMED_SUCCESS,
        },
      );

      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith(
        `CLAIMED_SUCCESS: Message with tx hash ${transactionReceipt.transactionHash} has been claimed.`,
      );
    });
  });
});
