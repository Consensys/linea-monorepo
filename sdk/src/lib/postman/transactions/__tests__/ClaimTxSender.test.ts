import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DataSource } from "typeorm";
import { JsonRpcProvider } from "@ethersproject/providers";
import { BigNumber } from "ethers";
import { LoggerOptions } from "winston";
import { L2MessageServiceContract } from "../../../contracts";
import { getTestL2Signer } from "../../../utils/testHelpers/contracts";
import { TEST_CONTRACT_ADDRESS_2, testL2NetworkConfig } from "../../../utils/testHelpers/constants";
import { ClaimTxSender } from "../ClaimTxSender";
import { DatabaseErrorType, DatabaseRepoName, Direction, MessageStatus } from "../../utils/enums";
import { LineaLogger, getLogger } from "../../../logger";
import { L2NetworkConfig, MessageInDb } from "../../utils/types";
import { MessageRepository } from "../../repositories/MessageRepository";
import { generateMessageFromDb } from "../../../utils/testHelpers/helpers";
import { DEFAULT_MAX_CLAIM_GAS_LIMIT } from "../../../utils/constants";
import { OnChainMessageStatus } from "../../../utils/enum";
import { DatabaseAccessError } from "../../utils/errors";

class TestClaimTxSender extends ClaimTxSender<L2MessageServiceContract> {
  public logger: LineaLogger;
  public repository: MessageRepository;

  constructor(
    dataSource: DataSource,
    messageServiceContract: L2MessageServiceContract,
    config: L2NetworkConfig,
    direction: Direction,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, messageServiceContract, config, direction);
    this.logger = getLogger("ClaimTxSender", loggerOptions);
    this.repository = this.messageRepository;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public async isRateLimitExceeded(_messageFee: string, _messageValue: string): Promise<boolean> {
    return false;
  }
  public async getFeesUsingOriginLayerBaseFeePerGas(): Promise<{
    maxFeePerGas?: BigNumber | undefined;
    maxPriorityFeePerGas?: BigNumber | undefined;
  }> {
    return {
      maxFeePerGas: BigNumber.from(100_000),
      maxPriorityFeePerGas: BigNumber.from(50_000),
    };
  }
  public async get1559Fees(): Promise<{
    maxFeePerGas?: BigNumber | undefined;
    maxPriorityFeePerGas?: BigNumber | undefined;
  }> {
    return {};
  }

  public async getNonce(): Promise<number | null> {
    return super.getNonce();
  }

  public async calculateGasEstimationAndThresHold(
    message: MessageInDb,
  ): Promise<{ threshold: number; estimatedGasLimit: BigNumber }> {
    return super.calculateGasEstimationAndThresHold(message);
  }

  public getGasLimit(gasLimit: BigNumber): BigNumber | null {
    return super.getGasLimit(gasLimit);
  }

  public async isTransactionUnderPriced(gasLimit: BigNumber, messageFee: string): Promise<boolean> {
    return super.isTransactionUnderPriced(gasLimit, messageFee);
  }

  public async executeClaimTransaction(message: MessageInDb, nonce: number, gasLimit: BigNumber) {
    super.executeClaimTransaction(message, nonce, gasLimit);
  }

  public async listenForReadyToBeClaimedMessages(interval: number) {
    await super.listenForReadyToBeClaimedMessages(interval);
  }
}

const pollingInterval = 1_000;

describe("ClaimTxSender", () => {
  let claimTxSender: TestClaimTxSender;
  let messageServiceContract: L2MessageServiceContract;

  beforeEach(() => {
    messageServiceContract = new L2MessageServiceContract(
      mock<JsonRpcProvider>(),
      TEST_CONTRACT_ADDRESS_2,
      "read-write",
      getTestL2Signer(),
    );

    claimTxSender = new TestClaimTxSender(
      mock<DataSource>(),
      messageServiceContract,
      testL2NetworkConfig,
      Direction.L1_TO_L2,
      {
        silent: true,
      },
    );
    jest.spyOn(messageServiceContract, "get1559Fees").mockResolvedValue({ maxFeePerGas: BigNumber.from(50_000) });
  });

  describe("getNonce", () => {
    it("should return on chain account nonce when there is no tx nonce is the DB ", async () => {
      const onChainAccountNonce = 10;
      jest.spyOn(claimTxSender.repository, "getLastTxNonce").mockResolvedValueOnce(null);
      jest.spyOn(messageServiceContract, "getCurrentNonce").mockResolvedValueOnce(onChainAccountNonce);

      expect(await claimTxSender.getNonce()).toStrictEqual(onChainAccountNonce);
    });

    it("should return null and log a warning message when the difference between on chain account nonce and last tx nonce in the DB is higher than maxNonceDiff", async () => {
      const onChainAccountNonce = 10;
      const lastTxNonceInDB = 15_000;
      const maxNonceDiff = 10_000;

      jest.spyOn(claimTxSender.repository, "getLastTxNonce").mockResolvedValueOnce(lastTxNonceInDB);
      jest.spyOn(messageServiceContract, "getCurrentNonce").mockResolvedValueOnce(onChainAccountNonce);
      const loggerWarnSpy = jest.spyOn(claimTxSender.logger, "warn");

      expect(await claimTxSender.getNonce()).toStrictEqual(null);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        `Last recorded nonce in db (${lastTxNonceInDB}) is higher than the latest nonce from blockchain (${onChainAccountNonce}) and exceeds the limit (${maxNonceDiff}), paused the claim message process now`,
      );
    });

    it("should return max value between lastTxNonceInDB and onChainAccountNonce when the difference between on chain account nonce and last tx nonce in the DB is less than maxNonceDiff", async () => {
      const onChainAccountNonce = 10_000;
      const lastTxNonceInDB = 15_000;

      jest.spyOn(claimTxSender.repository, "getLastTxNonce").mockResolvedValueOnce(lastTxNonceInDB);
      jest.spyOn(messageServiceContract, "getCurrentNonce").mockResolvedValueOnce(onChainAccountNonce);
      const loggerWarnSpy = jest.spyOn(claimTxSender.logger, "warn");

      expect(await claimTxSender.getNonce()).toStrictEqual(lastTxNonceInDB + 1);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(0);
    });
  });

  describe("calculateGasEstimationAndThresHold", () => {
    it("should return transaction gas limit and threshold", async () => {
      const txGasLimit = BigNumber.from(70_000);
      const message = generateMessageFromDb();
      jest.spyOn(messageServiceContract, "estimateClaimGas").mockResolvedValueOnce(txGasLimit);

      expect(await claimTxSender.calculateGasEstimationAndThresHold(message)).toStrictEqual({
        threshold: parseFloat(message.fee) / txGasLimit.toNumber(),
        estimatedGasLimit: txGasLimit,
      });
    });
  });

  describe("getGasLimit", () => {
    it("should return estimated gas limit when estimatedGasLimit <= maxClaimGasLimit", async () => {
      const estimatedGasLimit = BigNumber.from(70_000);
      expect(claimTxSender.getGasLimit(estimatedGasLimit)).toStrictEqual(estimatedGasLimit);
    });

    it("should return null when estimatedGasLimit > maxClaimGasLimit", async () => {
      const estimatedGasLimit = BigNumber.from(DEFAULT_MAX_CLAIM_GAS_LIMIT).mul(2);
      expect(claimTxSender.getGasLimit(estimatedGasLimit)).toStrictEqual(null);
    });
  });

  describe("isTransactionUnderPriced", () => {
    it("should return true when gasLimit * maxFeePerGas * profitMargin > messageFee", async () => {
      const txGasLimit = BigNumber.from(70_000);
      const message = generateMessageFromDb();

      expect(await claimTxSender.isTransactionUnderPriced(txGasLimit, message.fee)).toStrictEqual(true);
    });

    it("should return false when gasLimit * maxFeePerGas * profitMargin <= messageFee", async () => {
      const txGasLimit = BigNumber.from(70_000);
      const message = generateMessageFromDb({ fee: "10000000000000" });

      jest.spyOn(messageServiceContract, "get1559Fees").mockResolvedValueOnce({
        maxFeePerGas: BigNumber.from(100_000),
        maxPriorityFeePerGas: BigNumber.from(50_000),
      });

      expect(await claimTxSender.isTransactionUnderPriced(txGasLimit, message.fee)).toStrictEqual(false);
    });
  });

  describe("listenForReadyToBeClaimedMessages", () => {
    const onChainAccountNonce = 10;
    beforeEach(() => {
      jest.spyOn(claimTxSender.repository, "getLastTxNonce").mockResolvedValue(null);
      jest.spyOn(messageServiceContract, "getCurrentNonce").mockResolvedValue(onChainAccountNonce);
    });

    it("should do nothing and log a warning message when nonce returned from getNonce() is an invalid value", async () => {
      const lastTxNonceInDB = 15_000;
      jest.spyOn(claimTxSender.repository, "getLastTxNonce").mockResolvedValueOnce(lastTxNonceInDB);

      const loggerErrorSpy = jest.spyOn(claimTxSender.logger, "error");

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);
      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenCalledWith(
        "Nonce returned from getNonce is an invalid value (e.g. null or undefined)",
      );
    });

    it("should do nothing when there is no message to claim in the DB", async () => {
      const getFirstMessageToClaimSpy = jest
        .spyOn(claimTxSender.repository, "getFirstMessageToClaim")
        .mockResolvedValueOnce(null);

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);

      expect(getFirstMessageToClaimSpy).toHaveBeenCalledTimes(1);
    });

    it("should update the message status to ZERO_FEE when message fee === 0", async () => {
      const message = generateMessageFromDb({ fee: "0" });

      jest.spyOn(claimTxSender.repository, "getFirstMessageToClaim").mockResolvedValueOnce(message);
      const loggerWarnSpy = jest.spyOn(claimTxSender.logger, "warn");
      const updateMessageSpy = jest.spyOn(claimTxSender.repository, "updateMessage").mockResolvedValueOnce();

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(`Zero fee found in this message: ${JSON.stringify(message)}`);

      expect(updateMessageSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledWith(message.messageHash, message.direction, {
        status: MessageStatus.ZERO_FEE,
      });
    });

    it("should update the message status to CLAIM_SUCCEES when the message has already been claimed", async () => {
      const message = generateMessageFromDb();

      jest.spyOn(claimTxSender.repository, "getFirstMessageToClaim").mockResolvedValueOnce(message);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValueOnce(OnChainMessageStatus.CLAIMED);

      const loggerInfoSpy = jest.spyOn(claimTxSender.logger, "info");
      const updateMessageSpy = jest.spyOn(claimTxSender.repository, "updateMessage").mockResolvedValueOnce();

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);
      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith(`Found already claimed message: ${JSON.stringify(message)}`);

      expect(updateMessageSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledWith(message.messageHash, message.direction, {
        status: MessageStatus.CLAIMED_SUCCESS,
      });
    });

    it("should update the message status to NON_EXECUTABLE when the estimated gas limit is greater than maxClaimGasLimit", async () => {
      const message = generateMessageFromDb();
      const estimatedGasLimit = BigNumber.from(120_000);
      const maxClaimGasLimit = DEFAULT_MAX_CLAIM_GAS_LIMIT;

      jest.spyOn(claimTxSender.repository, "getFirstMessageToClaim").mockResolvedValueOnce(message);
      jest.spyOn(messageServiceContract, "estimateClaimGas").mockResolvedValueOnce(estimatedGasLimit);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValueOnce(OnChainMessageStatus.CLAIMABLE);

      const loggerWarnSpy = jest.spyOn(claimTxSender.logger, "warn");
      const updateMessageSpy = jest.spyOn(claimTxSender.repository, "updateMessage").mockResolvedValueOnce();

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        `Estimated gas limit (${estimatedGasLimit}) is higher than the max allowed gas limit (${maxClaimGasLimit}) for this message: ${JSON.stringify(
          message,
        )}`,
      );

      expect(updateMessageSpy).toHaveBeenCalledTimes(1);
      expect(updateMessageSpy).toHaveBeenCalledWith(message.messageHash, message.direction, {
        status: MessageStatus.NON_EXECUTABLE,
      });
    });

    it("should update the message status to FEE_UNDERPRICED when the transaction is underpriced", async () => {
      const message = generateMessageFromDb();
      const estimatedGasLimit = BigNumber.from(70_000);

      jest.spyOn(claimTxSender.repository, "getFirstMessageToClaim").mockResolvedValueOnce(message);
      jest.spyOn(messageServiceContract, "estimateClaimGas").mockResolvedValueOnce(estimatedGasLimit);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValueOnce(OnChainMessageStatus.CLAIMABLE);

      const loggerWarnSpy = jest.spyOn(claimTxSender.logger, "warn");
      const updateMessageSpy = jest.spyOn(claimTxSender.repository, "updateMessage").mockResolvedValue();

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(`Fee underpriced found in this message: ${JSON.stringify(message)}`);

      expect(updateMessageSpy).toHaveBeenCalledTimes(2);
      expect(updateMessageSpy).toHaveBeenLastCalledWith(message.messageHash, message.direction, {
        status: MessageStatus.FEE_UNDERPRICED,
      });
    });

    it("should log a warning message and reset the claimGasEstimationThreshold when the withdrawal rate limit exceeded", async () => {
      const message = generateMessageFromDb({ fee: "1000000000000" });
      const estimatedGasLimit = BigNumber.from(50_000);

      jest.spyOn(claimTxSender.repository, "getFirstMessageToClaim").mockResolvedValueOnce(message);
      jest.spyOn(messageServiceContract, "estimateClaimGas").mockResolvedValueOnce(estimatedGasLimit);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValueOnce(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(claimTxSender, "isRateLimitExceeded").mockResolvedValueOnce(true);

      const loggerWarnSpy = jest.spyOn(claimTxSender.logger, "warn");
      const updateMessageSpy = jest.spyOn(claimTxSender.repository, "updateMessage").mockResolvedValue();

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);
      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith(
        `Rate limit exceeded on L1 for this message: ${JSON.stringify(message)}`,
      );

      expect(updateMessageSpy).toHaveBeenCalledTimes(2);
      expect(updateMessageSpy).toHaveBeenLastCalledWith(message.messageHash, message.direction, {
        claimGasEstimationThreshold: undefined,
      });
    });

    it("should catch any http error and log an error", async () => {
      jest
        .spyOn(claimTxSender.repository, "getFirstMessageToClaim")
        .mockRejectedValueOnce(new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, {}));
      const loggerErrorSpy = jest.spyOn(claimTxSender.logger, "error");

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);
      expect(loggerErrorSpy).toHaveBeenCalledTimes(1);
      expect(loggerErrorSpy).toHaveBeenCalledWith(
        `Error found in listenForReadyToBeClaimedMessages:\nFailed message: ${JSON.stringify(
          null,
        )}\nFounded error: ${JSON.stringify({ name: "DatabaseAccessError" })}\nParsed error:${JSON.stringify({
          errorCode: "UNKNOWN_ERROR",
          mitigation: { shouldRetry: true, retryWithBlocking: true, retryPeriodInMs: 5000 },
        })}`,
      );
    });

    it("should claim message", async () => {
      const message = generateMessageFromDb({ fee: "1000000000000" });
      const estimatedGasLimit = BigNumber.from(70_000);

      jest.spyOn(claimTxSender.repository, "getFirstMessageToClaim").mockResolvedValueOnce(message);
      jest.spyOn(messageServiceContract, "estimateClaimGas").mockResolvedValueOnce(estimatedGasLimit);
      jest.spyOn(messageServiceContract, "getMessageStatus").mockResolvedValueOnce(OnChainMessageStatus.CLAIMABLE);
      jest.spyOn(claimTxSender.repository, "updateMessage").mockResolvedValue();
      const executeClaimTransactionSpy = jest.spyOn(claimTxSender, "executeClaimTransaction").mockResolvedValue();

      await claimTxSender.listenForReadyToBeClaimedMessages(pollingInterval);

      expect(executeClaimTransactionSpy).toHaveBeenCalledTimes(1);
      expect(executeClaimTransactionSpy).toHaveBeenCalledWith(message, onChainAccountNonce, estimatedGasLimit);
    });
  });
});
