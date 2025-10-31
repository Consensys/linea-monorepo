import { jest } from "@jest/globals";
import type { ILogger } from "@consensys/linea-shared-utils";
import type { INativeYieldAutomationMetricsUpdater } from "../../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { IOperationModeMetricsRecorder } from "../../../core/metrics/IOperationModeMetricsRecorder.js";
import type { IYieldManager } from "../../../core/clients/contracts/IYieldManager.js";
import type { ILazyOracle } from "../../../core/clients/contracts/ILazyOracle.js";
import type { ILidoAccountingReportClient } from "../../../core/clients/ILidoAccountingReportClient.js";
import type { IBeaconChainStakingClient } from "../../../core/clients/IBeaconChainStakingClient.js";
import type { TransactionReceipt, Address } from "viem";
import { OperationMode } from "../../../core/enums/OperationModeEnums.js";
import { OperationTrigger } from "../../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";
import { OssificationPendingProcessor } from "../OssificationPendingProcessor.js";
import { ResultAsync } from "neverthrow";

jest.mock("@consensys/linea-shared-utils", () => {
  const actual = jest.requireActual("@consensys/linea-shared-utils") as typeof import("@consensys/linea-shared-utils");
  return {
    ...actual,
    attempt: jest.fn(),
    msToSeconds: jest.fn(),
  };
});

import { attempt, msToSeconds } from "@consensys/linea-shared-utils";

describe("OssificationPendingProcessor", () => {
  const yieldProvider = "0x1111111111111111111111111111111111111111" as Address;
  const vaultAddress = "0x2222222222222222222222222222222222222222" as Address;

  let logger: jest.Mocked<ILogger>;
  let metricsUpdater: jest.Mocked<INativeYieldAutomationMetricsUpdater>;
  let metricsRecorder: jest.Mocked<IOperationModeMetricsRecorder>;
  let yieldManager: jest.Mocked<IYieldManager<TransactionReceipt>>;
  let lazyOracle: jest.Mocked<ILazyOracle<TransactionReceipt>>;
  let lidoReportClient: jest.Mocked<ILidoAccountingReportClient>;
  let beaconClient: jest.Mocked<IBeaconChainStakingClient>;
  const attemptMock = attempt as jest.MockedFunction<typeof attempt>;
  const msToSecondsMock = msToSeconds as jest.MockedFunction<typeof msToSeconds>;

  beforeEach(() => {
    jest.clearAllMocks();
    logger = {
      name: "test",
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
      warnOrError: jest.fn(),
    } as unknown as jest.Mocked<ILogger>;

    metricsUpdater = {
      incrementOperationModeTrigger: jest.fn(),
      recordOperationModeDuration: jest.fn(),
      incrementLidoVaultAccountingReport: jest.fn(),
      recordRebalance: jest.fn(),
      addNodeOperatorFeesPaid: jest.fn(),
      addLiabilitiesPaid: jest.fn(),
      addLidoFeesPaid: jest.fn(),
      incrementOperationModeExecution: jest.fn(),
      addValidatorPartialUnstakeAmount: jest.fn(),
      incrementValidatorExit: jest.fn(),
      incrementReportYield: jest.fn(),
      addReportedYieldAmount: jest.fn(),
      setCurrentNegativeYieldLastReport: jest.fn(),
    } as unknown as jest.Mocked<INativeYieldAutomationMetricsUpdater>;

    metricsRecorder = {
      recordSafeWithdrawalMetrics: jest.fn(),
      recordProgressOssificationMetrics: jest.fn(),
      recordReportYieldMetrics: jest.fn(),
      recordTransferFundsMetrics: jest.fn(),
    } as unknown as jest.Mocked<IOperationModeMetricsRecorder>;

    yieldManager = {
      progressPendingOssification: jest.fn(),
      safeMaxAddToWithdrawalReserve: jest.fn(),
      isOssified: jest.fn(),
      getLidoStakingVaultAddress: jest.fn(),
    } as unknown as jest.Mocked<IYieldManager<TransactionReceipt>>;

    lazyOracle = {
      waitForVaultsReportDataUpdatedEvent: jest.fn(),
    } as unknown as jest.Mocked<ILazyOracle<TransactionReceipt>>;

    lidoReportClient = {
      getLatestSubmitVaultReportParams: jest.fn(),
      isSimulateSubmitLatestVaultReportSuccessful: jest.fn(),
      submitLatestVaultReport: jest.fn(),
    } as unknown as jest.Mocked<ILidoAccountingReportClient>;

    beaconClient = {
      submitMaxAvailableWithdrawalRequests: jest.fn(),
    } as unknown as jest.Mocked<IBeaconChainStakingClient>;

    lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      result: OperationTrigger.VAULTS_REPORT_DATA_UPDATED_EVENT,
      report: {
        timestamp: 0n,
        refSlot: 0n,
        treeRoot: "0x" as `0x${string}`,
        reportCid: "cid",
      },
      txHash: "0xhash" as `0x${string}`,
    });
    yieldManager.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);
    lidoReportClient.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(true);
    yieldManager.progressPendingOssification.mockResolvedValue({
      transactionHash: "0xhash",
    } as unknown as TransactionReceipt);
    yieldManager.isOssified.mockResolvedValue(false);
    yieldManager.safeMaxAddToWithdrawalReserve.mockResolvedValue(undefined as unknown as TransactionReceipt);
    msToSecondsMock.mockImplementation((ms: number) => ms / 1_000);
    attemptMock.mockImplementation(((logger: ILogger, fn: () => unknown | Promise<unknown>) =>
      ResultAsync.fromPromise(
        Promise.resolve().then(() => fn()),
        (error) => error as Error,
      )) as typeof attempt);
  });

  const createProcessor = () =>
    new OssificationPendingProcessor(
      logger,
      metricsUpdater,
      metricsRecorder,
      yieldManager,
      lazyOracle,
      lidoReportClient,
      beaconClient,
      yieldProvider,
    );

  it("processes trigger events, submits reports, and progresses ossification", async () => {
    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(1000).mockReturnValueOnce(1600);

    const processor = createProcessor();

    await processor.process();

    expect(lazyOracle.waitForVaultsReportDataUpdatedEvent).toHaveBeenCalledTimes(1);
    expect(metricsUpdater.incrementOperationModeTrigger).toHaveBeenCalledWith(
      OperationMode.OSSIFICATION_PENDING_MODE,
      OperationTrigger.VAULTS_REPORT_DATA_UPDATED_EVENT,
    );
    expect(beaconClient.submitMaxAvailableWithdrawalRequests).toHaveBeenCalledTimes(1);
    expect(yieldManager.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);
    expect(lidoReportClient.getLatestSubmitVaultReportParams).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.submitLatestVaultReport).toHaveBeenCalledWith(vaultAddress);
    expect(metricsUpdater.incrementLidoVaultAccountingReport).toHaveBeenCalledWith(vaultAddress);
    expect(yieldManager.progressPendingOssification).toHaveBeenCalledWith(yieldProvider);
    expect(metricsRecorder.recordProgressOssificationMetrics).toHaveBeenCalledWith(yieldProvider, expect.anything());
    expect(yieldManager.isOssified).toHaveBeenCalledWith(yieldProvider);
    expect(metricsRecorder.recordSafeWithdrawalMetrics).not.toHaveBeenCalled();
    expect(msToSeconds).toHaveBeenCalledWith(600);
    expect(metricsUpdater.recordOperationModeDuration).toHaveBeenCalledWith(
      OperationMode.OSSIFICATION_PENDING_MODE,
      0.6,
    );
    performanceSpy.mockRestore();
  });

  it("skips submitting reports when simulation fails", async () => {
    lidoReportClient.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(false);

    const processor = createProcessor();
    await processor.process();

    expect(lidoReportClient.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
  });

  it("returns early when progressPendingOssification fails", async () => {
    yieldManager.progressPendingOssification.mockRejectedValue(new Error("progress failed"));

    const processor = createProcessor();
    await processor.process();

    expect(metricsRecorder.recordProgressOssificationMetrics).not.toHaveBeenCalled();
    expect(yieldManager.safeMaxAddToWithdrawalReserve).not.toHaveBeenCalled();
  });

  it("performs safe withdrawal when ossification completes", async () => {
    yieldManager.isOssified.mockResolvedValue(true);

    const processor = createProcessor();
    await processor.process();

    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledWith(yieldProvider, expect.anything());
  });
});
