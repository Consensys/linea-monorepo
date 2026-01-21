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
import type { IVaultHub } from "../../../core/clients/contracts/IVaultHub.js";
import type { Address, TransactionReceipt, Hex } from "viem";
import { OperationTrigger } from "../../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";
import { OperationMode } from "../../../core/enums/OperationModeEnums.js";
import { RebalanceDirection } from "../../../core/entities/RebalanceRequirement.js";
import { YieldReportingProcessor } from "../YieldReportingProcessor.js";
import type { UpdateVaultDataParams } from "../../../core/clients/contracts/ILazyOracle.js";
import type { YieldReport } from "../../../core/entities/YieldReport.js";

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
  let vaultHubClient: jest.Mocked<IVaultHub<TransactionReceipt>>;
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
    } as unknown as jest.Mocked<ILogger>;

    metricsUpdater = {
      recordOperationModeDuration: jest.fn(),
      incrementLidoVaultAccountingReport: jest.fn(),
      recordRebalance: jest.fn(),
      setLastPeekedNegativeYieldReport: jest.fn(),
      setLastPeekedPositiveYieldReport: jest.fn(),
      setLastSettleableLidoFees: jest.fn(),
      setLastVaultReportTimestamp: jest.fn(),
      setYieldReportedCumulative: jest.fn(),
      setLstLiabilityPrincipalGwei: jest.fn(),
      setLastReportedNegativeYield: jest.fn(),
      setLidoLstLiabilityGwei: jest.fn(),
      setLastTotalPendingPartialWithdrawalsGwei: jest.fn(),
      setPendingPartialWithdrawalQueueAmountGwei: jest.fn(),
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
      peekYieldReport: jest.fn(),
      getYieldProviderData: jest.fn(),
      getBalance: jest.fn(),
      getTargetReserveDeficit: jest.fn(),
      safeWithdrawFromYieldProvider: jest.fn(),
    } as unknown as jest.Mocked<IYieldManager<TransactionReceipt>>;

    lazyOracle = {
      waitForVaultsReportDataUpdatedEvent: jest.fn(),
    } as unknown as jest.Mocked<ILazyOracle<TransactionReceipt>>;

    lidoReportClient = {
      getLatestSubmitVaultReportParams: jest.fn(),
      submitLatestVaultReport: jest.fn(),
    } as unknown as jest.Mocked<ILidoAccountingReportClient>;

    yieldExtension = {
      transferFundsForNativeYield: jest.fn(),
    } as unknown as jest.Mocked<ILineaRollupYieldExtension<TransactionReceipt>>;

    beaconClient = {
      submitWithdrawalRequestsToFulfilAmount: jest.fn(),
    } as unknown as jest.Mocked<IBeaconChainStakingClient>;

    vaultHubClient = {
      settleableLidoFeesValue: jest.fn(),
      isReportFresh: jest.fn(),
      isVaultConnected: jest.fn(),
    } as unknown as jest.Mocked<IVaultHub<TransactionReceipt>>;

    lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      result: OperationTrigger.TIMEOUT,
    });
    yieldManager.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);
    yieldManager.getBalance.mockResolvedValue(0n);
    yieldManager.getTargetReserveDeficit.mockResolvedValue(0n);
    yieldManager.safeWithdrawFromYieldProvider.mockResolvedValue({
      transactionHash: "0xwithdraw",
    } as unknown as TransactionReceipt);
    lidoReportClient.getLatestSubmitVaultReportParams.mockResolvedValue(submitParams);
    lidoReportClient.submitLatestVaultReport.mockResolvedValue(undefined);
    yieldManager.reportYield.mockResolvedValue({ transactionHash: "0xyield" } as unknown as TransactionReceipt);
    yieldExtension.transferFundsForNativeYield.mockResolvedValue({
      transactionHash: "0xtransfer",
    } as unknown as TransactionReceipt);
    yieldManager.fundYieldProvider.mockResolvedValue({ transactionHash: "0xfund" } as unknown as TransactionReceipt);
    yieldManager.safeAddToWithdrawalReserveIfAboveThreshold.mockResolvedValue(
      undefined as unknown as TransactionReceipt,
    );

    vaultHubClient.isVaultConnected.mockResolvedValue(true);
    vaultHubClient.isReportFresh.mockResolvedValue(false);

    attemptMock.mockImplementation(((loggerArg: ILogger, fn: () => unknown | Promise<unknown>) =>
      ResultAsync.fromPromise((async () => fn())(), (error) => error as Error)) as typeof attempt);
    msToSecondsMock.mockImplementation((ms: number) => ms / 1_000);
    weiToGweiNumberMock.mockImplementation((value: bigint) => Number(value));
  });

  const createProcessor = (
    shouldSubmitVaultReport: boolean = true,
    shouldReportYield: boolean = true,
    isUnpauseStakingEnabled: boolean = true,
    minNegativeYieldDiffToReportYieldWei: bigint = 1000000000000000000n,
    minWithdrawalThresholdEth: bigint = 0n,
    cyclesPerYieldReport: number = 12,
  ) =>
    new YieldReportingProcessor(
      logger,
      metricsUpdater,
      metricsRecorder,
      yieldManager,
      lazyOracle,
      yieldExtension,
      lidoReportClient,
      beaconClient,
      vaultHubClient,
      yieldProvider,
      l2Recipient,
      shouldSubmitVaultReport,
      shouldReportYield,
      isUnpauseStakingEnabled,
      minNegativeYieldDiffToReportYieldWei,
      minWithdrawalThresholdEth,
      cyclesPerYieldReport,
    );

  it("_process - processes staking surplus flow and records metrics", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: stakeAmount })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 5n });

    // Mock _shouldReportYield to return true by triggering cycle-based reporting
    vaultHubClient.settleableLidoFeesValue.mockResolvedValue(600000000000000000n);
    yieldManager.peekYieldReport.mockResolvedValue({
      yieldAmount: 2000000000000000000n,
      outstandingNegativeYield: 0n,
      yieldProvider,
    });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(100).mockReturnValueOnce(250);

    const processor = createProcessor();
    // Set cycleCount to 11 so that when _process() increments it to 12, cycle-based reporting will trigger (12 % 12 === 0)
    (processor as any).cycleCount = 11;
    await processor.process();

    expect(lazyOracle.waitForVaultsReportDataUpdatedEvent).toHaveBeenCalledTimes(1);
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

  it("_process - pauses staking when starting in deficit and skips unpause", async () => {
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

  it("_process - skips vault report submission when shouldSubmitVaultReport is false", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    // Mock _shouldReportYield to return true by triggering cycle-based reporting
    vaultHubClient.settleableLidoFeesValue.mockResolvedValue(600000000000000000n);
    yieldManager.peekYieldReport.mockResolvedValue({
      yieldAmount: 2000000000000000000n,
      outstandingNegativeYield: 0n,
      yieldProvider,
    });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(100).mockReturnValueOnce(200);

    const processor = createProcessor(false, true, true);
    // Set cycleCount to 11 so that when _process() increments it to 12, cycle-based reporting will trigger (12 % 12 === 0)
    (processor as any).cycleCount = 11;
    await processor.process();

    expect(lidoReportClient.getLatestSubmitVaultReportParams).not.toHaveBeenCalled();
    expect(lidoReportClient.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(
      "_handleSubmitLatestVaultReport - Skipping vault report submission (SHOULD_SUBMIT_VAULT_REPORT=false)",
    );
    expect(yieldManager.reportYield).toHaveBeenCalledWith(yieldProvider, l2Recipient);
    expect(metricsRecorder.recordReportYieldMetrics).toHaveBeenCalledTimes(1);

    performanceSpy.mockRestore();
  });

  it("_process - performs an amendment unstake when flow flips to deficit mid-cycle (STAKE→UNSTAKE)", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: 4n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 6n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(50).mockReturnValueOnce(200);
    const processor = createProcessor();
    const amendmentSpy = jest.spyOn(
      processor as unknown as { _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void> },
      "_handleUnstakingRebalance",
    );

    await processor.process();

    expect(amendmentSpy).toHaveBeenCalledWith(6n, false);
    expect(yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();

    performanceSpy.mockRestore();
    amendmentSpy.mockRestore();
  });

  it("_process - performs an amendment unstake when flow flips to deficit mid-cycle (NONE→UNSTAKE)", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 5n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(100).mockReturnValueOnce(250);
    const processor = createProcessor();
    const amendmentSpy = jest.spyOn(
      processor as unknown as { _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void> },
      "_handleUnstakingRebalance",
    );

    await processor.process();

    expect(amendmentSpy).toHaveBeenCalledWith(5n, false);
    expect(yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();

    performanceSpy.mockRestore();
    amendmentSpy.mockRestore();
  });

  it("_process - if start in excess, and no amendment unstake is required, will perform unpause staking", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: 3n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    const processor = createProcessor();
    const amendmentSpy = jest.spyOn(
      processor as unknown as { _handleUnstakingRebalance(amount: bigint, success: boolean): Promise<void> },
      "_handleUnstakingRebalance",
    );

    await processor.process();

    expect(amendmentSpy).not.toHaveBeenCalled();
    expect(yieldManager.unpauseStakingIfNotAlready).toHaveBeenCalledWith(yieldProvider);

    amendmentSpy.mockRestore();
  });

  it("_process - does not perform ending beacon chain withdrawal if there is no ending deficit", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: 2n });

    const processor = createProcessor();
    await processor.process();

    expect(beaconClient.submitWithdrawalRequestsToFulfilAmount).not.toHaveBeenCalled();
    expect(yieldManager.unpauseStakingIfNotAlready).toHaveBeenCalledWith(yieldProvider);
  });

  it("_process - unpauses staking when starting and ending in non-deficit state (NONE→NONE)", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    // Mock _shouldReportYield to return true
    vaultHubClient.settleableLidoFeesValue.mockResolvedValue(600000000000000000n); // above threshold
    yieldManager.peekYieldReport.mockResolvedValue({
      yieldAmount: 2000000000000000000n,
      outstandingNegativeYield: 0n,
      yieldProvider,
    });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(100).mockReturnValueOnce(200);
    const processor = createProcessor();
    const amendmentSpy = jest.spyOn(
      processor as unknown as { _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void> },
      "_handleUnstakingRebalance",
    );

    await processor.process();

    expect(amendmentSpy).not.toHaveBeenCalled();
    expect(yieldManager.unpauseStakingIfNotAlready).toHaveBeenCalledWith(yieldProvider);
    expect(beaconClient.submitWithdrawalRequestsToFulfilAmount).not.toHaveBeenCalled();

    performanceSpy.mockRestore();
    amendmentSpy.mockRestore();
  });

  it("_process - does not unpause staking when isUnpauseStakingEnabled is false (STAKE→NONE)", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: 3n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(100).mockReturnValueOnce(200);
    const processor = createProcessor(true, true, false); // isUnpauseStakingEnabled = false
    const amendmentSpy = jest.spyOn(
      processor as unknown as { _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void> },
      "_handleUnstakingRebalance",
    );

    await processor.process();

    expect(amendmentSpy).not.toHaveBeenCalled();
    expect(yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();

    performanceSpy.mockRestore();
    amendmentSpy.mockRestore();
  });

  it("_process - does not unpause staking when isUnpauseStakingEnabled is false (NONE→NONE)", async () => {
    yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n })
      .mockResolvedValueOnce({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    // Mock _shouldReportYield to return true
    vaultHubClient.settleableLidoFeesValue.mockResolvedValue(600000000000000000n); // above threshold
    yieldManager.peekYieldReport.mockResolvedValue({
      yieldAmount: 2000000000000000000n,
      outstandingNegativeYield: 0n,
      yieldProvider,
    });

    const performanceSpy = jest.spyOn(performance, "now").mockReturnValueOnce(100).mockReturnValueOnce(200);
    const processor = createProcessor(true, true, false); // isUnpauseStakingEnabled = false
    const amendmentSpy = jest.spyOn(
      processor as unknown as { _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void> },
      "_handleUnstakingRebalance",
    );

    await processor.process();

    expect(amendmentSpy).not.toHaveBeenCalled();
    expect(yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();
    expect(beaconClient.submitWithdrawalRequestsToFulfilAmount).not.toHaveBeenCalled();

    performanceSpy.mockRestore();
    amendmentSpy.mockRestore();
  });

  it("_handleRebalance reports yield when no rebalance is needed", async () => {
    const processor = createProcessor();
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
  });

  it("_handleNoRebalance transfers YieldManager balance when above threshold", async () => {
    const minWithdrawalThresholdEth = 1n; // 1 ETH threshold
    const yieldManagerBalance = 2n * 1000000000000000000n; // 2 ETH (above threshold)
    const processor = createProcessor(true, true, true, 1000000000000000000n, minWithdrawalThresholdEth);
    
    yieldManager.getBalance.mockResolvedValueOnce(yieldManagerBalance);
    yieldManager.getTargetReserveDeficit.mockResolvedValueOnce(0n);
    yieldManager.fundYieldProvider.mockResolvedValueOnce({ transactionHash: "0xtransfer" } as unknown as TransactionReceipt);
    
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    expect(yieldManager.getBalance).toHaveBeenCalledTimes(1);
    expect(yieldManager.fundYieldProvider).toHaveBeenCalledWith(yieldProvider, yieldManagerBalance);
    expect(metricsRecorder.recordTransferFundsMetrics).toHaveBeenCalledWith(yieldProvider, expect.any(Object));
    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
  });

  it("_handleNoRebalance skips transfer when YieldManager balance is below threshold", async () => {
    const minWithdrawalThresholdEth = 1n; // 1 ETH threshold
    const yieldManagerBalance = 500000000000000000n; // 0.5 ETH (below threshold)
    const processor = createProcessor(true, true, true, 1000000000000000000n, minWithdrawalThresholdEth);
    
    yieldManager.getBalance.mockResolvedValueOnce(yieldManagerBalance);
    yieldManager.getTargetReserveDeficit.mockResolvedValueOnce(0n);
    
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    expect(yieldManager.getBalance).toHaveBeenCalledTimes(1);
    expect(yieldManager.fundYieldProvider).not.toHaveBeenCalled();
    expect(metricsRecorder.recordTransferFundsMetrics).not.toHaveBeenCalled();
    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
  });

  it("_handleNoRebalance withdraws from yield provider when targetReserveDeficit > 0", async () => {
    const minWithdrawalThresholdEth = 1n; // 1 ETH threshold
    const yieldManagerBalance = 2n * 1000000000000000000n; // 2 ETH (above threshold)
    const targetReserveDeficit = 500000000000000000n; // 0.5 ETH deficit
    const processor = createProcessor(true, true, true, 1000000000000000000n, minWithdrawalThresholdEth);
    
    yieldManager.getBalance.mockResolvedValueOnce(yieldManagerBalance);
    yieldManager.getTargetReserveDeficit.mockResolvedValueOnce(targetReserveDeficit);
    yieldManager.safeWithdrawFromYieldProvider.mockResolvedValueOnce({
      transactionHash: "0xwithdraw",
    } as unknown as TransactionReceipt);
    yieldManager.fundYieldProvider.mockResolvedValueOnce({ transactionHash: "0xtransfer" } as unknown as TransactionReceipt);
    
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.NONE, rebalanceAmount: 0n });

    expect(yieldManager.getBalance).toHaveBeenCalledTimes(1);
    expect(yieldManager.getTargetReserveDeficit).toHaveBeenCalledTimes(1);
    expect(yieldManager.safeWithdrawFromYieldProvider).toHaveBeenCalledWith(yieldProvider, targetReserveDeficit);
    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledWith(yieldProvider, expect.any(Object));
    expect(yieldManager.fundYieldProvider).toHaveBeenCalledWith(yieldProvider, yieldManagerBalance);
    expect(metricsRecorder.recordTransferFundsMetrics).toHaveBeenCalledWith(yieldProvider, expect.any(Object));
    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
  });

  it("_handleRebalance routes excess directions to _handleStakingRebalance", async () => {
    const processor = createProcessor();
    const stakingSpy = jest
      .spyOn(
        processor as unknown as { _handleStakingRebalance(amount: bigint): Promise<void> },
        "_handleStakingRebalance",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.STAKE, rebalanceAmount: 42n });

    expect(stakingSpy).toHaveBeenCalledWith(42n);
  });

  it("_handleRebalance routes deficit directions to _handleUnstakingRebalance", async () => {
    const processor = createProcessor();
    const unstakeSpy = jest
      .spyOn(
        processor as unknown as { _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void> },
        "_handleUnstakingRebalance",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleRebalance(req: unknown): Promise<void>;
      }
    )._handleRebalance({ rebalanceDirection: RebalanceDirection.UNSTAKE, rebalanceAmount: 17n });

    expect(unstakeSpy).toHaveBeenCalledWith(17n, true);
  });

  it("_handleStakingRebalance successfully calls rebalance functions and reports yield", async () => {
    const processor = createProcessor();
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleStakingRebalance(amount: bigint): Promise<void>;
      }
    )._handleStakingRebalance(18n);

    expect(yieldExtension.transferFundsForNativeYield).toHaveBeenCalledWith(18n);
    expect(metricsUpdater.recordRebalance).toHaveBeenCalledWith(RebalanceDirection.STAKE, Number(18n));
    expect(yieldManager.fundYieldProvider).toHaveBeenCalledWith(yieldProvider, 18n);
    expect(metricsRecorder.recordTransferFundsMetrics).toHaveBeenCalledTimes(1);
    const transferResult = metricsRecorder.recordTransferFundsMetrics.mock.calls[0][1];
    expect(transferResult.isOk()).toBe(true);
    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
    reportYieldSpy.mockRestore();
  });

  it("_handleStakingRebalance tolerates failure of fundYieldProvider", async () => {
    yieldManager.fundYieldProvider.mockRejectedValueOnce(new Error("fund fail"));

    const processor = createProcessor();
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleStakingRebalance(amount: bigint): Promise<void>;
      }
    )._handleStakingRebalance(11n);

    expect(yieldExtension.transferFundsForNativeYield).toHaveBeenCalledWith(11n);
    expect(metricsUpdater.recordRebalance).toHaveBeenCalledWith(RebalanceDirection.STAKE, Number(11n));
    expect(metricsRecorder.recordTransferFundsMetrics).toHaveBeenCalledTimes(1);
    const result = metricsRecorder.recordTransferFundsMetrics.mock.calls[0][1];
    expect(result.isErr()).toBe(true);
    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
    reportYieldSpy.mockRestore();
  });

  it("_handleStakingRebalance tolerates failure of transferFundsForNativeYieldResult, but will skip fundYieldProvider and metrics update", async () => {
    yieldExtension.transferFundsForNativeYield.mockRejectedValueOnce(new Error("transfer fail"));

    const processor = createProcessor();
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleStakingRebalance(amount: bigint): Promise<void>;
      }
    )._handleStakingRebalance(9n);

    expect(metricsUpdater.recordRebalance).not.toHaveBeenCalled();
    expect(yieldManager.fundYieldProvider).not.toHaveBeenCalled();
    expect(metricsRecorder.recordTransferFundsMetrics).not.toHaveBeenCalled();
    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
    reportYieldSpy.mockRestore();
  });

  it("_handleUnstakingRebalance reports yield before withdrawing when shouldReportYield is true", async () => {
    const processor = createProcessor();
    const reportYieldSpy = jest
      .spyOn(
        processor as unknown as { _handleReportYield(): Promise<unknown> },
        "_handleReportYield",
      )
      .mockResolvedValue(undefined);

    await (
      processor as unknown as {
        _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void>;
      }
    )._handleUnstakingRebalance(15n, true);

    expect(reportYieldSpy).toHaveBeenCalledTimes(1);
    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledTimes(1);
  });

  it("_handleUnstakingRebalance skips yield reporting when shouldReportYield is false", async () => {
    const processor = createProcessor();
    const reportYieldSpy = jest.spyOn(
      processor as unknown as { _handleReportYield(): Promise<unknown> },
      "_handleReportYield",
    );

    await (
      processor as unknown as {
        _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void>;
      }
    )._handleUnstakingRebalance(7n, false);

    expect(reportYieldSpy).not.toHaveBeenCalled();
    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledTimes(1);
  });

  it("_handleUnstakingRebalance tolerates failure of safeAddToWithdrawalReserveIfAboveThreshold", async () => {
    const error = new Error("withdrawal failed");
    yieldManager.safeAddToWithdrawalReserveIfAboveThreshold.mockRejectedValueOnce(error);
    const processor = createProcessor();

    await (
      processor as unknown as {
        _handleUnstakingRebalance(amount: bigint, shouldReportYield: boolean): Promise<void>;
      }
    )._handleUnstakingRebalance(20n, false);

    expect(metricsRecorder.recordSafeWithdrawalMetrics).toHaveBeenCalledTimes(1);
    const result = metricsRecorder.recordSafeWithdrawalMetrics.mock.calls[0][1];
    expect(result.isErr()).toBe(true);
  });

  describe("_handleReportYield", () => {
    it("reports yield when _shouldReportYield returns true", async () => {
      const processor = createProcessor();
      (processor as unknown as { vault: Address }).vault = vaultAddress;
      const shouldReportYieldSpy = jest
        .spyOn(processor as unknown as { _shouldReportYield(): Promise<boolean> }, "_shouldReportYield")
        .mockResolvedValue(true);

      await (
        processor as unknown as {
          _handleReportYield(): Promise<void>;
        }
      )._handleReportYield();

      expect(shouldReportYieldSpy).toHaveBeenCalledTimes(1);
      expect(yieldManager.reportYield).toHaveBeenCalledWith(yieldProvider, l2Recipient);
      expect(metricsRecorder.recordReportYieldMetrics).toHaveBeenCalledWith(
        yieldProvider,
        expect.objectContaining({ isOk: expect.any(Function) }),
      );
      expect(logger.info).toHaveBeenCalledWith("_handleReportYield: yield report succeeded");
      shouldReportYieldSpy.mockRestore();
    });

    it("skips yield reporting when _shouldReportYield returns false", async () => {
      const processor = createProcessor();
      (processor as unknown as { vault: Address }).vault = vaultAddress;
      const shouldReportYieldSpy = jest
        .spyOn(processor as unknown as { _shouldReportYield(): Promise<boolean> }, "_shouldReportYield")
        .mockResolvedValue(false);

      await (
        processor as unknown as {
          _handleReportYield(): Promise<void>;
        }
      )._handleReportYield();

      expect(shouldReportYieldSpy).toHaveBeenCalledTimes(1);
      expect(yieldManager.reportYield).not.toHaveBeenCalled();
      expect(metricsRecorder.recordReportYieldMetrics).not.toHaveBeenCalled();
      shouldReportYieldSpy.mockRestore();
    });

    it("tolerates failure of reportYield", async () => {
      yieldManager.reportYield.mockRejectedValueOnce(new Error("yield fail"));

      const processor = createProcessor();
      (processor as unknown as { vault: Address }).vault = vaultAddress;
      const shouldReportYieldSpy = jest
        .spyOn(processor as unknown as { _shouldReportYield(): Promise<boolean> }, "_shouldReportYield")
        .mockResolvedValue(true);

      await (
        processor as unknown as {
          _handleReportYield(): Promise<void>;
        }
      )._handleReportYield();

      expect(shouldReportYieldSpy).toHaveBeenCalledTimes(1);
      expect(yieldManager.reportYield).toHaveBeenCalledWith(yieldProvider, l2Recipient);
      expect(metricsRecorder.recordReportYieldMetrics).not.toHaveBeenCalled();
      shouldReportYieldSpy.mockRestore();
    });
  });

  it("_handleSubmitLatestVaultReport handles vault submission failure gracefully", async () => {
    vaultHubClient.isReportFresh.mockResolvedValueOnce(false);
    lidoReportClient.submitLatestVaultReport.mockRejectedValueOnce(new Error("vault fail"));

    const processor = createProcessor();
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<void>;
      }
    )._handleSubmitLatestVaultReport();

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(yieldManager.reportYield).not.toHaveBeenCalled();
    expect(metricsRecorder.recordReportYieldMetrics).not.toHaveBeenCalled();
  });

  it("_handleSubmitLatestVaultReport logs success when vault report succeeds", async () => {
    vaultHubClient.isReportFresh.mockResolvedValueOnce(false);

    const processor = createProcessor();
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<void>;
      }
    )._handleSubmitLatestVaultReport();

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(metricsUpdater.incrementLidoVaultAccountingReport).toHaveBeenCalledWith(vaultAddress);
    expect(logger.info).toHaveBeenCalledWith("_handleSubmitLatestVaultReport - Vault report submission succeeded");
    expect(yieldManager.reportYield).not.toHaveBeenCalled();
    expect(metricsRecorder.recordReportYieldMetrics).not.toHaveBeenCalled();
  });

  it("_handleSubmitLatestVaultReport skips submission when report is fresh", async () => {
    vaultHubClient.isReportFresh.mockResolvedValueOnce(true);

    const processor = createProcessor();
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<void>;
      }
    )._handleSubmitLatestVaultReport();

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.getLatestSubmitVaultReportParams).not.toHaveBeenCalled();
    expect(lidoReportClient.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith("_handleSubmitLatestVaultReport - Skipping vault report submission (report is fresh)");
  });

  it("_handleSubmitLatestVaultReport proceeds with submission when isReportFresh check fails", async () => {
    vaultHubClient.isReportFresh.mockRejectedValueOnce(new Error("check failed"));

    const processor = createProcessor();
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<void>;
      }
    )._handleSubmitLatestVaultReport();

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(logger.warn).toHaveBeenCalledWith(
      expect.stringContaining("Failed to check if report is fresh, proceeding with submission attempt"),
    );
    expect(lidoReportClient.getLatestSubmitVaultReportParams).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.submitLatestVaultReport).toHaveBeenCalled();
  });


  it("_handleSubmitLatestVaultReport skips vault report submission when shouldSubmitVaultReport is false", async () => {
    const processor = createProcessor(false, true, true);
    (processor as unknown as { vault: Address }).vault = vaultAddress;

    await (
      processor as unknown as {
        _handleSubmitLatestVaultReport(): Promise<void>;
      }
    )._handleSubmitLatestVaultReport();

    expect(lidoReportClient.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(
      "_handleSubmitLatestVaultReport - Skipping vault report submission (SHOULD_SUBMIT_VAULT_REPORT=false)",
    );
    expect(yieldManager.reportYield).not.toHaveBeenCalled();
    expect(metricsRecorder.recordReportYieldMetrics).not.toHaveBeenCalled();
  });


  describe("_shouldReportYield", () => {
    const minNegativeYieldDiffToReportYieldWei = 1000000000000000000n;

    beforeEach(() => {
      // Default mock for getYieldProviderData - can be overridden in individual tests
      yieldManager.getYieldProviderData.mockResolvedValue({
        yieldProviderVendor: 0,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        yieldProviderIndex: 0n,
        userFunds: 0n,
        yieldReportedCumulative: 0n,
        lstLiabilityPrincipal: 0n,
        lastReportedNegativeYield: 0n,
      });
    });

    it("returns true when negative yield diff threshold is met", async () => {
      const unpaidFees = 300000000000000000n; // below threshold
      const onStateNegativeYield = 1000000000000000000n; // 1 ETH
      const peekedNegativeYield = 2500000000000000000n; // 2.5 ETH
      // negativeYieldDiff = peekedNegativeYield - onStateNegativeYield = 1.5 ETH, above threshold
      const yieldReport: YieldReport = {
        yieldAmount: 500000000000000000n, // below threshold
        outstandingNegativeYield: peekedNegativeYield,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);
      yieldManager.getYieldProviderData.mockResolvedValue({
        yieldProviderVendor: 0,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        yieldProviderIndex: 0n,
        userFunds: 0n,
        yieldReportedCumulative: 0n,
        lstLiabilityPrincipal: 0n,
        lastReportedNegativeYield: onStateNegativeYield,
      });

      const processor = createProcessor(
        true,
        true,
        true,
        minNegativeYieldDiffToReportYieldWei,
      );
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(true);
      expect(yieldManager.getYieldProviderData).toHaveBeenCalledWith(yieldProvider);
    });

    it("returns false when negative yield diff is below threshold", async () => {
      const unpaidFees = 300000000000000000n; // below threshold
      const onStateNegativeYield = 1000000000000000000n; // 1 ETH
      const peekedNegativeYield = 1500000000000000000n; // 1.5 ETH
      // negativeYieldDiff = peekedNegativeYield - onStateNegativeYield = 0.5 ETH, below threshold
      const yieldReport: YieldReport = {
        yieldAmount: 500000000000000000n, // below threshold
        outstandingNegativeYield: peekedNegativeYield,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);
      yieldManager.getYieldProviderData.mockResolvedValue({
        yieldProviderVendor: 0,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        yieldProviderIndex: 0n,
        userFunds: 0n,
        yieldReportedCumulative: 0n,
        lstLiabilityPrincipal: 0n,
        lastReportedNegativeYield: onStateNegativeYield,
      });

      const processor = createProcessor(
        true,
        true,
        true,
        minNegativeYieldDiffToReportYieldWei,
      );
      // Set cycleCount to a value that's not divisible by cyclesPerYieldReport (default 12)
      (processor as any).cycleCount = 1; // 1 % 12 !== 0
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(false);
    });

    it("returns false when neither threshold is met", async () => {
      const unpaidFees = 300000000000000000n; // below threshold
      const yieldReport: YieldReport = {
        yieldAmount: 500000000000000000n, // below threshold
        outstandingNegativeYield: 0n,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);

      const processor = createProcessor(
        true,
        true,
        true,
        minNegativeYieldDiffToReportYieldWei,
      );
      // Set cycleCount to a value that's not divisible by cyclesPerYieldReport (default 12)
      (processor as any).cycleCount = 1; // 1 % 12 !== 0
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(false);
    });

    it("handles undefined yieldReport gracefully by treating yieldAmount as 0n", async () => {
      const unpaidFees = 300000000000000000n; // below threshold

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(undefined);

      const processor = createProcessor(
        true,
        true,
        true,
        minNegativeYieldDiffToReportYieldWei,
      );
      // Set cycleCount to a value that's not divisible by cyclesPerYieldReport (default 12)
      (processor as any).cycleCount = 1; // 1 % 12 !== 0
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(false);
    });

    it("returns false when both values are zero", async () => {
      const unpaidFees = 0n;
      const yieldReport: YieldReport = {
        yieldAmount: 0n,
        outstandingNegativeYield: 0n,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);

      const processor = createProcessor(true, true);
      // Set cycleCount to a value that's not divisible by cyclesPerYieldReport (default 12)
      (processor as any).cycleCount = 1; // 1 % 12 !== 0
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(false);
    });

    it("logs both results correctly", async () => {
      const unpaidFees = 600000000000000000n;
      const yieldReport: YieldReport = {
        yieldAmount: 2000000000000000000n,
        outstandingNegativeYield: 0n,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);

      const processor = createProcessor(true, true);
      // Set cycleCount to trigger cycle-based reporting
      (processor as any).cycleCount = 12; // 12 % 12 === 0
      await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("_shouldReportYield - shouldReportYield=true"),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining('settleableLidoFees="600000000000000000"'),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining('"yieldAmount":"2000000000000000000"'),
      );
    });

    it("sets all three gauge metrics when reads succeed", async () => {
      const unpaidFees = 600000000000000000n;
      const yieldReport: YieldReport = {
        yieldAmount: 2000000000000000000n,
        outstandingNegativeYield: 1000000000000000000n,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);

      const processor = createProcessor(true, true);
      (processor as unknown as { vault: Address }).vault = vaultAddress;
      await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(metricsUpdater.setLastPeekedNegativeYieldReport).toHaveBeenCalledWith(vaultAddress, 1000000000000000000);
      expect(metricsUpdater.setLastPeekedPositiveYieldReport).toHaveBeenCalledWith(vaultAddress, 2000000000000000000);
      expect(metricsUpdater.setLastSettleableLidoFees).toHaveBeenCalledWith(vaultAddress, 600000000000000000);
    });

    it("sets metrics with zero values when yieldReport is undefined", async () => {
      const unpaidFees = 300000000000000000n;

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(undefined);

      const processor = createProcessor(true, true);
      (processor as unknown as { vault: Address }).vault = vaultAddress;
      await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      // When yieldReport is undefined, these metrics should not be called
      expect(metricsUpdater.setLastPeekedNegativeYieldReport).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastPeekedPositiveYieldReport).not.toHaveBeenCalled();
      // Only unpaid fees metric should be set when unpaidFees is defined
      expect(metricsUpdater.setLastSettleableLidoFees).toHaveBeenCalledWith(vaultAddress, 300000000000000000);
    });

    it("returns true when cycle count is divisible by cyclesPerYieldReport", async () => {
      const unpaidFees = 300000000000000000n; // below threshold
      const yieldReport: YieldReport = {
        yieldAmount: 500000000000000000n, // below threshold
        outstandingNegativeYield: 0n,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);
      yieldManager.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);

      const cyclesPerYieldReport = 12;
      const processor = createProcessor(
        true,
        true,
        true,
        minNegativeYieldDiffToReportYieldWei,
        0n,
        cyclesPerYieldReport,
      );
      (processor as any).vault = vaultAddress;

      // Set cycleCount to a multiple of cyclesPerYieldReport
      (processor as any).cycleCount = 24; // 24 % 12 === 0
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(true);
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("cycleBasedReportingDue=true"),
      );
    });

    it("returns false when shouldReportYield is false even if thresholds are met", async () => {
      const unpaidFees = 600000000000000000n;
      const onStateNegativeYield = 0n;
      const peekedNegativeYield = 2000000000000000000n; // 2 ETH difference, above threshold
      const yieldReport: YieldReport = {
        yieldAmount: 0n,
        outstandingNegativeYield: peekedNegativeYield,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);
      yieldManager.getYieldProviderData.mockResolvedValue({
        yieldProviderVendor: 0,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        yieldProviderIndex: 0n,
        userFunds: 0n,
        yieldReportedCumulative: 0n,
        lstLiabilityPrincipal: 0n,
        lastReportedNegativeYield: onStateNegativeYield,
      });

      const processor = createProcessor(true, false); // shouldReportYield = false
      (processor as any).vault = vaultAddress;
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(false);
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("_shouldReportYield - shouldReportYield=false"),
      );
    });

    it("returns false when cycle count is not divisible by cyclesPerYieldReport and thresholds not met", async () => {
      const unpaidFees = 300000000000000000n; // below threshold
      const yieldReport: YieldReport = {
        yieldAmount: 500000000000000000n, // below threshold
        outstandingNegativeYield: 0n,
        yieldProvider,
      };

      vaultHubClient.settleableLidoFeesValue.mockResolvedValue(unpaidFees);
      yieldManager.peekYieldReport.mockResolvedValue(yieldReport);
      yieldManager.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);

      const cyclesPerYieldReport = 12;
      const processor = createProcessor(
        true,
        true,
        true,
        minNegativeYieldDiffToReportYieldWei,
        0n,
        cyclesPerYieldReport,
      );
      (processor as any).vault = vaultAddress;

      // Set cycleCount to not be a multiple of cyclesPerYieldReport
      (processor as any).cycleCount = 11; // 11 % 12 !== 0
      const result = await (processor as unknown as { _shouldReportYield(): Promise<boolean> })._shouldReportYield();

      expect(result).toBe(false);
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("cycleBasedReportingDue=false"),
      );
    });
  });
});
