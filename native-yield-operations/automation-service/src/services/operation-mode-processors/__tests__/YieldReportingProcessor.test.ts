import { jest } from "@jest/globals";
import { ResultAsync } from "neverthrow";
import type { ILogger } from "@consensys/linea-shared-utils";
import type { INativeYieldAutomationMetricsUpdater } from "../../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { IOperationModeMetricsRecorder } from "../../../core/metrics/IOperationModeMetricsRecorder.js";
import type { IYieldManager } from "../../../core/clients/contracts/IYieldManager.js";
import type { ILazyOracle } from "../../../core/clients/contracts/ILazyOracle.js";
import type { ILidoAccountingReportClient } from "../../../core/clients/ILidoAccountingReportClient.js";
import type { ILineaRollupYieldExtension } from "../../../core/clients/contracts/ILineaRollupYieldExtension.js";
import type { IBeaconChainStakingClient } from "../../../core/clients/IBeaconChainStakingClient.js";
import type { Address, TransactionReceipt, Hex } from "viem";
import { OperationTrigger } from "../../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";
import { OperationMode } from "../../../core/enums/OperationModeEnums.js";
import { RebalanceDirection } from "../../../core/entities/RebalanceRequirement.js";
import { YieldReportingProcessor } from "../YieldReportingProcessor.js";
import type { UpdateVaultDataParams } from "../../../core/clients/contracts/ILazyOracle.js";

jest.mock("@consensys/linea-shared-utils", () => {
  const actual = jest.requireActual("@consensys/linea-shared-utils") as typeof import("@consensys/linea-shared-utils");
  return {
    ...actual,
    attempt: jest.fn(),
    msToSeconds: jest.fn(),
    weiToGweiNumber: jest.fn(),
  };
});

import { attempt, msToSeconds, weiToGweiNumber } from "@consensys/linea-shared-utils";

describe("YieldReportingProcessor", () => {
  const yieldProvider = "0x1111111111111111111111111111111111111111" as Address;
  const l2Recipient = "0x2222222222222222222222222222222222222222" as Address;
  const vaultAddress = "0x3333333333333333333333333333333333333333" as Address;
  const stakeAmount = 10n;
  const submitParams: UpdateVaultDataParams = {
    vault: vaultAddress,
    totalValue: 0n,
    cumulativeLidoFees: 0n,
    liabilityShares: 0n,
    maxLiabilityShares: 0n,
    slashingReserve: 0n,
    proof: [] as Hex[],
  };

  let logger: jest.Mocked<ILogger>;
  let metricsUpdater: jest.Mocked<INativeYieldAutomationMetricsUpdater>;
  let metricsRecorder: jest.Mocked<IOperationModeMetricsRecorder>;
  let yieldManager: jest.Mocked<IYieldManager<TransactionReceipt>>;
  let lazyOracle: jest.Mocked<ILazyOracle<TransactionReceipt>>;
  let lidoReportClient: jest.Mocked<ILidoAccountingReportClient>;
  let yieldExtension: jest.Mocked<ILineaRollupYieldExtension<TransactionReceipt>>;
  let beaconClient: jest.Mocked<IBeaconChainStakingClient>;
  const attemptMock = attempt as jest.MockedFunction<typeof attempt>;
  const msToSecondsMock = msToSeconds as jest.MockedFunction<typeof msToSeconds>;
  const weiToGweiNumberMock = weiToGweiNumber as jest.MockedFunction<typeof weiToGweiNumber>;

  beforeEach(() => {
    jest.clearAllMocks();

    logger = {
      name: "logger",
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
    } as unknown as jest.Mocked<INativeYieldAutomationMetricsUpdater>;

    metricsRecorder = {
      recordTransferFundsMetrics: jest.fn(),
      recordSafeWithdrawalMetrics: jest.fn(),
      recordReportYieldMetrics: jest.fn(),
    } as unknown as jest.Mocked<IOperationModeMetricsRecorder>;

    yieldManager = {
      getLidoStakingVaultAddress: jest.fn(),
      getRebalanceRequirements: jest.fn(),
      pauseStakingIfNotAlready: jest.fn(),
      unpauseStakingIfNotAlready: jest.fn(),
      fundYieldProvider: jest.fn(),
      safeAddToWithdrawalReserveIfAboveThreshold: jest.fn(),
      reportYield: jest.fn(),
    } as unknown as jest.Mocked<IYieldManager<TransactionReceipt>>;

    lazyOracle = {
      waitForVaultsReportDataUpdatedEvent: jest.fn(),
    } as unknown as jest.Mocked<ILazyOracle<TransactionReceipt>>;

    lidoReportClient = {
      getLatestSubmitVaultReportParams: jest.fn(),
      isSimulateSubmitLatestVaultReportSuccessful: jest.fn(),
      submitLatestVaultReport: jest.fn(),
    } as unknown as jest.Mocked<ILidoAccountingReportClient>;

    yieldExtension = {
      transferFundsForNativeYield: jest.fn(),
    } as unknown as jest.Mocked<ILineaRollupYieldExtension<TransactionReceipt>>;

    beaconClient = {
      submitWithdrawalRequestsToFulfilAmount: jest.fn(),
    } as unknown as jest.Mocked<IBeaconChainStakingClient>;

    lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      result: OperationTrigger.TIMEOUT,
    });
    yieldManager.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);
    lidoReportClient.getLatestSubmitVaultReportParams.mockResolvedValue(submitParams);
    lidoReportClient.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(true);
    lidoReportClient.submitLatestVaultReport.mockResolvedValue(undefined);
    yieldManager.reportYield.mockResolvedValue({ transactionHash: "0xyield" } as unknown as TransactionReceipt);
    yieldExtension.transferFundsForNativeYield.mockResolvedValue({
      transactionHash: "0xtransfer",
    } as unknown as TransactionReceipt);
    yieldManager.fundYieldProvider.mockResolvedValue({ transactionHash: "0xfund" } as unknown as TransactionReceipt);
    yieldManager.safeAddToWithdrawalReserveIfAboveThreshold.mockResolvedValue(
      undefined as unknown as TransactionReceipt,
    );

    attemptMock.mockImplementation(((loggerArg: ILogger, fn: () => unknown | Promise<unknown>) =>
      ResultAsync.fromPromise((async () => fn())(), (error) => error as Error)) as typeof attempt);
    msToSecondsMock.mockImplementation((ms: number) => ms / 1_000);
    weiToGweiNumberMock.mockImplementation((value: bigint) => Number(value));
  });

  const createProcessor = () =>
    new YieldReportingProcessor(
      logger,
      metricsUpdater,
      metricsRecorder,
      yieldManager,
      lazyOracle,
      yieldExtension,
      lidoReportClient,
      beaconClient,
      yieldProvider,
      l2Recipient,
    );

  it("processes staking surplus flow and records metrics", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: stakeAmount })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 5n });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(100).mockReturnValueOnce(250);

    const processor = createProcessor();
    await processor.process();

    expect(lazyOracle.waitForVaultsReportDataUpdatedEvent).toHaveBeenCalledTimes(1);
    expect(metricsUpdater.incrementOperationModeTrigger).toHaveBeenCalledWith(
      OperationMode.YIELD_REPORTING_MODE,
      OperationTrigger.TIMEOUT,
    );
    expect(yieldManager.pauseStakingIfNotAlready).not.toHaveBeenCalled();
    expect(yieldExtension.transferFundsForNativeYield).toHaveBeenCalledWith(stakeAmount);
    expect(metricsUpdater.recordRebalance).toHaveBeenCalledWith(RebalanceDirection.STAKE, Number(stakeAmount));
    expect(metricsRecorder.recordTransferFundsMetrics).toHaveBeenCalledTimes(1);
    const transferResult = metricsRecorder.recordTransferFundsMetrics.mock.calls[0][1];
    expect(transferResult.isOk()).toBe(true);
    expect(lidoReportClient.submitLatestVaultReport).toHaveBeenCalledWith(vaultAddress);
    expect(metricsUpdater.incrementLidoVaultAccountingReport).toHaveBeenCalledWith(vaultAddress);
    expect(metricsRecorder.recordReportYieldMetrics).toHaveBeenCalledTimes(1);
    expect(yieldManager.unpauseStakingIfNotAlready).toHaveBeenCalledWith(yieldProvider);
    expect(beaconClient.submitWithdrawalRequestsToFulfilAmount).toHaveBeenCalledWith(5n);
    expect(msToSeconds).toHaveBeenCalledWith(150);
    expect(metricsUpdater.recordOperationModeDuration).toHaveBeenCalledWith(OperationMode.YIELD_REPORTING_MODE, 0.15);

    performanceSpy.mockRestore();
  });

  it("pauses staking when starting in deficit and skips unpause", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 8n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(200).mockReturnValueOnce(260);

    const processor = createProcessor();
    await processor.process();

    expect(yieldManager.pauseStakingIfNotAlready).toHaveBeenCalledWith(yieldProvider);
    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledTimes(1);
    expect(yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();
    expect(beaconClient.submitWithdrawalRequestsToFulfilAmount).not.toHaveBeenCalled();

    performanceSpy.mockRestore();
  });

  it("performs an amendment unstake when stake flow flips to deficit mid-cycle", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: 4n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 6n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(50).mockReturnValueOnce(200);
    const processor = createProcessor();
    const amendmentSpy = jest.spyOn(
      processor as unknown as { _handleUnstakingRebalance(amount: bigint, success: boolean): Promise<void> },
      "_handleUnstakingRebalance",
    );

    await processor.process();

    expect(amendmentSpy).toHaveBeenCalledWith(6n, false);
    expect(yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();

    performanceSpy.mockRestore();
    amendmentSpy.mockRestore();
  });

  it("_handleRebalance handles no-op without submission when simulation fails", async () => {
    const processor = createProcessor();
    const submitSpy = jest.spyOn(
      processor as unknown as { _handleSubmitLatestVaultReport(): Promise<unknown> },
      "_handleSubmitLatestVaultReport",
    );

    await (
      processor as unknown as {
        _handleRebalance(req: unknown, successful: boolean): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n }, false);

    expect(logger.info).toHaveBeenCalledWith("_handleRebalance - no-op");
    expect(submitSpy).not.toHaveBeenCalled();
  });

  it("_handleRebalance submits report when no rebalance is needed but simulation succeeds", async () => {
    const processor = createProcessor();
    const submitSpy = jest
      .spyOn(
        processor as unknown as { _handleSubmitLatestVaultReport(): Promise<unknown> },
        "_handleSubmitLatestVaultReport",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown, successful: boolean): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n }, true);

    expect(submitSpy).toHaveBeenCalledTimes(1);
  });

  it("_handleRebalance routes staking directions to _handleStakingRebalance", async () => {
    const processor = createProcessor();
    const stakingSpy = jest
      .spyOn(
        processor as unknown as { _handleStakingRebalance(amount: bigint, success: boolean): Promise<void> },
        "_handleStakingRebalance",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown, successful: boolean): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: 42n }, true);

    expect(stakingSpy).toHaveBeenCalledWith(42n, true);
  });

  it("_handleRebalance routes deficit directions to _handleUnstakingRebalance", async () => {
    const processor = createProcessor();
    const unstakeSpy = jest
      .spyOn(
        processor as unknown as { _handleUnstakingRebalance(amount: bigint, success: boolean): Promise<void> },
        "_handleUnstakingRebalance",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown, successful: boolean): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 17n }, true);

    expect(unstakeSpy).toHaveBeenCalledWith(17n, true);
  });

  it("_handleStakingRebalance skips transfer metrics when the first attempt fails", async () => {
    yieldExtension.transferFundsForNativeYield.mockRejectedValueOnce(new Error("transfer failed"));

    const processor = createProcessor();
    await (
      processor as unknown as {
        _handleStakingRebalance(amount: bigint, successful: boolean): Promise<void>;
      }
    )._handleStakingRebalance(9n, true);

    expect(metricsUpdater.recordRebalance).not.toHaveBeenCalled();
    expect(yieldManager.fundYieldProvider).not.toHaveBeenCalled();
    expect(metricsRecorder.recordTransferFundsMetrics).not.toHaveBeenCalled();
  });

  it("_handleStakingRebalance omits report submission when simulation fails", async () => {
    const processor = createProcessor();
    const submitSpy = jest.spyOn(
      processor as unknown as { _handleSubmitLatestVaultReport(): Promise<unknown> },
      "_handleSubmitLatestVaultReport",
    );

    await (
      processor as unknown as {
        _handleStakingRebalance(amount: bigint, successful: boolean): Promise<void>;
      }
    )._handleStakingRebalance(12n, false);

    expect(submitSpy).not.toHaveBeenCalled();
    expect(metricsUpdater.recordRebalance).toHaveBeenCalled();
  });

  it("_handleUnstakingRebalance submits report before withdrawing when simulation succeeds", async () => {
    const processor = createProcessor();
    const submitSpy = jest
      .spyOn(
        processor as unknown as { _handleSubmitLatestVaultReport(): Promise<unknown> },
        "_handleSubmitLatestVaultReport",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleUnstakingRebalance(amount: bigint, successful: boolean): Promise<void>;
      }
    )._handleUnstakingRebalance(15n, true);

    expect(submitSpy).toHaveBeenCalledTimes(1);
    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledTimes(1);
  });

  it("_handleUnstakingRebalance skips report submission when simulation fails", async () => {
    const processor = createProcessor();
    const submitSpy = jest.spyOn(
      processor as unknown as { _handleSubmitLatestVaultReport(): Promise<unknown> },
      "_handleSubmitLatestVaultReport",
    );

    await (
      processor as unknown as {
        _handleUnstakingRebalance(amount: bigint, successful: boolean): Promise<void>;
      }
    )._handleUnstakingRebalance(7n, false);

    expect(submitSpy).not.toHaveBeenCalled();
    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledTimes(1);
  });

  it("_handleSubmitLatestVaultReport returns early when vault submission fails", async () => {
    lidoReportClient.submitLatestVaultReport.mockRejectedValueOnce(new Error("vault fail"));

    const processor = createProcessor();
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    const result = await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<unknown>;
      }
    )._handleSubmitLatestVaultReport();

    expect(result).toBeUndefined();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(metricsRecorder.recordReportYieldMetrics).not.toHaveBeenCalled();
  });

  it("_handleSubmitLatestVaultReport stops when yield reporting fails", async () => {
    yieldManager.reportYield.mockRejectedValueOnce(new Error("yield fail"));

    const processor = createProcessor();
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    const result = await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<unknown>;
      }
    )._handleSubmitLatestVaultReport();

    expect(result).toBeUndefined();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).toHaveBeenCalledWith(vaultAddress);
    expect(metricsRecorder.recordReportYieldMetrics).not.toHaveBeenCalled();
  });

  it("_handleSubmitLatestVaultReport logs success when both steps succeed", async () => {
    const processor = createProcessor();
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    const result = await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<unknown>;
      }
    )._handleSubmitLatestVaultReport();

    expect(result).toBeDefined();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).toHaveBeenCalledWith(vaultAddress);
    expect(metricsRecorder.recordReportYieldMetrics).toHaveBeenCalledWith(
      yieldProvider,
      expect.objectContaining({ isOk: expect.any(Function) }),
    );
    expect(logger.info).toHaveBeenCalledWith("_handleSubmitLatestVaultReport: vault report + yield report succeeded");
  });
});
