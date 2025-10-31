import { Address, TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { ILogger, attempt, msToSeconds } from "@consensys/linea-shared-utils";
import { ILazyOracle } from "../../core/clients/contracts/ILazyOracle.js";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient.js";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { OperationMode } from "../../core/enums/OperationModeEnums.js";
import { IOperationModeMetricsRecorder } from "../../core/metrics/IOperationModeMetricsRecorder.js";

export class OssificationPendingProcessor implements IOperationModeProcessor {
  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly operationModeMetricsRecorder: IOperationModeMetricsRecorder,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly lidoAccountingReportClient: ILidoAccountingReportClient,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly yieldProvider: Address,
  ) {}

  /**
   * Executes one processing cycle:
   * - Waits for the next `VaultsReportDataUpdated` event **or** a timeout, whichever happens first.
   * - Once triggered, runs the main processing logic (`_process()`).
   * - Always cleans up the event watcher afterward.
   */
  public async process(): Promise<void> {
    const triggerEvent = await this.lazyOracleContractClient.waitForVaultsReportDataUpdatedEvent();
    this.metricsUpdater.incrementOperationModeTrigger(OperationMode.OSSIFICATION_PENDING_MODE, triggerEvent.result);
    const startedAt = performance.now();
    await this._process();
    const durationMs = performance.now() - startedAt;
    this.metricsUpdater.recordOperationModeDuration(OperationMode.OSSIFICATION_PENDING_MODE, msToSeconds(durationMs));
  }

  /**
   * Main processing loop:
   * 1. Submit vault report if available
   * 2. Perform processPendingOssifcation
   * 3. Max withdraw
   * 4. Max unstake
   */
  private async _process(): Promise<void> {
    // Max unstake
    this.logger.info("_process - performing max unstake from beacon chain");
    await attempt(
      this.logger,
      () => this.beaconChainStakingClient.submitMaxAvailableWithdrawalRequests(),
      "submitMaxAvailableWithdrawalRequests failed (tolerated)",
    );

    this.logger.info("_process - Fetching latest vault report");
    // Submit vault report if available
    const vault = await this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider);
    await this.lidoAccountingReportClient.getLatestSubmitVaultReportParams(vault);
    const isSimulateSubmitLatestVaultReportSuccessful =
      await this.lidoAccountingReportClient.isSimulateSubmitLatestVaultReportSuccessful(vault);
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      this.logger.info("_process - Submitting latest vault report");
      const vaultReportResult = await attempt(
        this.logger,
        () => this.lidoAccountingReportClient.submitLatestVaultReport(vault),
        "submitLatestVaultReport failed (tolerated)",
      );
      if (vaultReportResult.isOk()) {
        this.metricsUpdater.incrementLidoVaultAccountingReport(vault);
      }
    }

    // Process Pending Ossification

    const ossificationResult = await attempt(
      this.logger,
      () => this.yieldManagerContractClient.progressPendingOssification(this.yieldProvider),
      "_process - progressPendingOssification failed",
    );
    // Stop if failed.
    if (ossificationResult.isErr()) return;

    await this.operationModeMetricsRecorder.recordProgressOssificationMetrics(this.yieldProvider, ossificationResult);
    this.logger.info("_process - Ossification completed, performing max safe withdrawal");

    // Max withdraw if ossified
    if (await this.yieldManagerContractClient.isOssified(this.yieldProvider)) {
      const withdrawalResult = await attempt(
        this.logger,
        () => this.yieldManagerContractClient.safeMaxAddToWithdrawalReserve(this.yieldProvider),
        "_process - safeMaxAddToWithdrawalReserve failed",
      );
      await this.operationModeMetricsRecorder.recordSafeWithdrawalMetrics(this.yieldProvider, withdrawalResult);
    }
  }
}
