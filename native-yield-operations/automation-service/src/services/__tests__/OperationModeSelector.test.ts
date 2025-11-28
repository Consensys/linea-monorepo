import { mock, MockProxy } from "jest-mock-extended";
import type { ILogger } from "@consensys/linea-shared-utils";
import type { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import type { IValidatorDataClient } from "../../core/clients/IValidatorDataClient.js";
import type { ValidatorBalanceWithPendingWithdrawal } from "../../core/entities/ValidatorBalance.js";
import type { TransactionReceipt, Address } from "viem";
import type { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { OperationMode } from "../../core/enums/OperationModeEnums.js";
import { OperationModeExecutionStatus } from "../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";
import { ONE_GWEI } from "@consensys/linea-shared-utils";
import { ResultAsync } from "neverthrow";

jest.mock("@consensys/linea-shared-utils", () => {
  const actual = jest.requireActual("@consensys/linea-shared-utils");
  return {
    ...actual,
    wait: jest.fn(),
    attempt: jest.fn(),
  };
});

import { wait, attempt } from "@consensys/linea-shared-utils";
import { OperationModeSelector } from "../OperationModeSelector.js";

describe("OperationModeSelector", () => {
  const yieldProvider = "0x1111111111111111111111111111111111111111" as Address;

  let logger: MockProxy<ILogger>;
  let metricsUpdater: MockProxy<INativeYieldAutomationMetricsUpdater>;
  let yieldManager: MockProxy<IYieldManager<TransactionReceipt>>;
  let validatorDataClient: MockProxy<IValidatorDataClient>;
  let yieldReportingProcessor: MockProxy<IOperationModeProcessor>;
  let ossificationPendingProcessor: MockProxy<IOperationModeProcessor>;
  let ossificationCompleteProcessor: MockProxy<IOperationModeProcessor>;
  let waitMock: jest.MockedFunction<typeof wait>;
  let attemptMock: jest.MockedFunction<typeof attempt>;

  const contractReadRetryTimeMs = 123;

  const createSelector = (retryTime = contractReadRetryTimeMs) =>
    new OperationModeSelector(
      logger,
      metricsUpdater,
      yieldManager,
      validatorDataClient,
      yieldReportingProcessor,
      ossificationPendingProcessor,
      ossificationCompleteProcessor,
      yieldProvider,
      retryTime,
    );

  beforeEach(() => {
    jest.clearAllMocks();
    logger = mock<ILogger>();
    metricsUpdater = mock<INativeYieldAutomationMetricsUpdater>();
    yieldManager = mock<IYieldManager<TransactionReceipt>>();
    validatorDataClient = mock<IValidatorDataClient>();
    yieldReportingProcessor = mock<IOperationModeProcessor>();
    ossificationPendingProcessor = mock<IOperationModeProcessor>();
    ossificationCompleteProcessor = mock<IOperationModeProcessor>();

    waitMock = wait as jest.MockedFunction<typeof wait>;
    waitMock.mockResolvedValue(undefined);

    attemptMock = attempt as jest.MockedFunction<typeof attempt>;
    attemptMock.mockImplementation((logger, fn, msg) => {
      return ResultAsync.fromPromise(Promise.resolve().then(fn), (e) => {
        const error = e instanceof Error ? e : new Error(String(e));
        logger.warn(msg, { error });
        return error;
      });
    });

    // Default mock for validator data client
    validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue([]);
    validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(0n);
  });

  it("runs yield reporting mode when neither ossification flag is set", async () => {
    yieldManager.isOssificationInitiated.mockResolvedValueOnce(false);
    yieldManager.isOssified.mockResolvedValueOnce(false);

    const selector = createSelector();
    yieldReportingProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(yieldReportingProcessor.process).toHaveBeenCalledTimes(1);
    expect(ossificationPendingProcessor.process).not.toHaveBeenCalled();
    expect(ossificationCompleteProcessor.process).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.YIELD_REPORTING_MODE,
      OperationModeExecutionStatus.Success,
    );
    expect(waitMock).not.toHaveBeenCalled();
  });

  it("runs ossification pending mode when ossification is initiated", async () => {
    yieldManager.isOssificationInitiated.mockResolvedValueOnce(true);
    yieldManager.isOssified.mockResolvedValueOnce(false);

    const selector = createSelector();
    ossificationPendingProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(ossificationPendingProcessor.process).toHaveBeenCalledTimes(1);
    expect(yieldReportingProcessor.process).not.toHaveBeenCalled();
    expect(ossificationCompleteProcessor.process).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.OSSIFICATION_PENDING_MODE,
      OperationModeExecutionStatus.Success,
    );
  });

  it("prefers ossification complete mode when ossified", async () => {
    yieldManager.isOssificationInitiated.mockResolvedValueOnce(true);
    yieldManager.isOssified.mockResolvedValueOnce(true);

    const selector = createSelector();
    ossificationCompleteProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(ossificationCompleteProcessor.process).toHaveBeenCalledTimes(1);
    expect(yieldReportingProcessor.process).not.toHaveBeenCalled();
    expect(ossificationPendingProcessor.process).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.OSSIFICATION_COMPLETE_MODE,
      OperationModeExecutionStatus.Success,
    );
  });

  it("is a no-op when start is invoked while already running", async () => {
    yieldManager.isOssificationInitiated.mockResolvedValue(false);
    yieldManager.isOssified.mockResolvedValue(false);

    const selector = createSelector();
    yieldReportingProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await Promise.all([selector.start(), selector.start()]);

    expect(logger.debug).toHaveBeenCalledWith("OperationModeSelector.start() - already running, skipping");
    expect(yieldReportingProcessor.process).toHaveBeenCalledTimes(1);
  });

  it("does nothing when stop is called before start", () => {
    const selector = createSelector();

    selector.stop();

    expect(logger.debug).toHaveBeenCalledWith("OperationModeSelector.stop() - not running, skipping");
    expect(logger.info).not.toHaveBeenCalled();
  });

  it("logs errors, waits, and retries before succeeding", async () => {
    const error = new Error("boom");
    const retryTime = 321;

    yieldManager.isOssificationInitiated.mockRejectedValueOnce(error);
    yieldManager.isOssificationInitiated.mockResolvedValueOnce(false);
    yieldManager.isOssified.mockRejectedValueOnce(error);
    yieldManager.isOssified.mockResolvedValueOnce(false);

    const selector = createSelector(retryTime);
    yieldReportingProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(logger.error).toHaveBeenCalledWith(`selectOperationModeLoop error, retrying in ${retryTime}ms`, { error });
    expect(waitMock).toHaveBeenCalledWith(retryTime);
    expect(yieldReportingProcessor.process).toHaveBeenCalledTimes(1);
    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.YIELD_REPORTING_MODE,
      OperationModeExecutionStatus.Success,
    );
    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.UNKNOWN,
      OperationModeExecutionStatus.Failure,
    );
  });

  it("increments failure metric with UNKNOWN when error occurs during contract reads", async () => {
    const error = new Error("contract read failed");
    const retryTime = 100;

    yieldManager.isOssificationInitiated.mockRejectedValueOnce(error);
    yieldManager.isOssificationInitiated.mockResolvedValueOnce(false);
    yieldManager.isOssified.mockResolvedValueOnce(false);

    const selector = createSelector(retryTime);
    yieldReportingProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.UNKNOWN,
      OperationModeExecutionStatus.Failure,
    );
  });

  it("increments failure metric with specific mode when error occurs during processor execution", async () => {
    const error = new Error("processor failed");
    const retryTime = 100;

    yieldManager.isOssificationInitiated.mockResolvedValue(false);
    yieldManager.isOssified.mockResolvedValue(false);
    yieldReportingProcessor.process.mockRejectedValueOnce(error);

    const selector = createSelector(retryTime);
    yieldReportingProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.YIELD_REPORTING_MODE,
      OperationModeExecutionStatus.Failure,
    );
  });

  it("increments failure metric with OSSIFICATION_PENDING_MODE when error occurs during pending processor execution", async () => {
    const error = new Error("pending processor failed");
    const retryTime = 100;

    yieldManager.isOssificationInitiated.mockResolvedValue(true);
    yieldManager.isOssified.mockResolvedValue(false);
    ossificationPendingProcessor.process.mockRejectedValueOnce(error);

    const selector = createSelector(retryTime);
    ossificationPendingProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.OSSIFICATION_PENDING_MODE,
      OperationModeExecutionStatus.Failure,
    );
  });

  it("increments failure metric with OSSIFICATION_COMPLETE_MODE when error occurs during complete processor execution", async () => {
    const error = new Error("complete processor failed");
    const retryTime = 100;

    yieldManager.isOssificationInitiated.mockResolvedValue(true);
    yieldManager.isOssified.mockResolvedValue(true);
    ossificationCompleteProcessor.process.mockRejectedValueOnce(error);

    const selector = createSelector(retryTime);
    ossificationCompleteProcessor.process.mockImplementation(async () => {
      selector.stop();
    });

    await selector.start();

    expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
      OperationMode.OSSIFICATION_COMPLETE_MODE,
      OperationModeExecutionStatus.Failure,
    );
  });

  describe("refreshGaugeMetrics", () => {
    it("refreshes gauge metrics on each loop iteration", async () => {
      yieldManager.isOssificationInitiated.mockResolvedValue(false);
      yieldManager.isOssified.mockResolvedValue(false);

      const validators: ValidatorBalanceWithPendingWithdrawal[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
          pendingWithdrawalAmount: 3n,
          withdrawableAmount: 0n,
        },
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-2",
          validatorIndex: 2n,
          pendingWithdrawalAmount: 1n,
          withdrawableAmount: 0n,
        },
      ];

      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue(validators);
      validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(4n * ONE_GWEI);

      const selector = createSelector();
      yieldReportingProcessor.process.mockImplementation(async () => {
        selector.stop();
      });

      await selector.start();

      expect(attemptMock).toHaveBeenCalled();
      expect(validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending).toHaveBeenCalled();
      expect(validatorDataClient.getTotalPendingPartialWithdrawalsWei).toHaveBeenCalledWith(validators);
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalledWith(4);
    });

    it("handles undefined validator list gracefully", async () => {
      yieldManager.isOssificationInitiated.mockResolvedValue(false);
      yieldManager.isOssified.mockResolvedValue(false);

      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue(undefined);

      const selector = createSelector();
      yieldReportingProcessor.process.mockImplementation(async () => {
        selector.stop();
      });

      await selector.start();

      expect(attemptMock).toHaveBeenCalled();
      expect(validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending).toHaveBeenCalled();
      expect(validatorDataClient.getTotalPendingPartialWithdrawalsWei).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).not.toHaveBeenCalled();
    });

    it("logs warning and error details but continues loop when refreshGaugeMetrics fails", async () => {
      yieldManager.isOssificationInitiated.mockResolvedValue(false);
      yieldManager.isOssified.mockResolvedValue(false);

      const error = new Error("Failed to get validators");
      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockRejectedValue(error);

      const selector = createSelector();
      yieldReportingProcessor.process.mockImplementation(async () => {
        selector.stop();
      });

      await selector.start();

      expect(attemptMock).toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Failed to refresh gauge metrics", { error });
      expect(logger.error).toHaveBeenCalledWith("Failed to refresh gauge metrics with details", {
        error: expect.objectContaining({ message: "Failed to get validators" }),
        errorMessage: "Failed to get validators",
        errorStack: expect.any(String),
      });
      expect(yieldReportingProcessor.process).toHaveBeenCalledTimes(1);
      expect(metricsUpdater.incrementOperationModeExecution).toHaveBeenCalledWith(
        OperationMode.YIELD_REPORTING_MODE,
        OperationModeExecutionStatus.Success,
      );
    });
  });
});
