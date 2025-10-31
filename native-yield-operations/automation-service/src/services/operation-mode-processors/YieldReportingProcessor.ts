import { Address, TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { bigintReplacer, ILogger, attempt, msToSeconds, weiToGweiNumber } from "@consensys/linea-shared-utils";
import { ILazyOracle } from "../../core/clients/contracts/ILazyOracle.js";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient.js";
import { RebalanceDirection, RebalanceRequirement } from "../../core/entities/RebalanceRequirement.js";
import { ILineaRollupYieldExtension } from "../../core/clients/contracts/ILineaRollupYieldExtension.js";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { OperationMode } from "../../core/enums/OperationModeEnums.js";
import { IOperationModeMetricsRecorder } from "../../core/metrics/IOperationModeMetricsRecorder.js";

/**
 * Processor for YIELD_REPORTING_MODE operations.
 * Handles native yield reporting cycles including rebalancing, vault report submission, and beacon chain withdrawals.
 */
export class YieldReportingProcessor implements IOperationModeProcessor {
  private vault: Address;

  /**
   * Creates a new YieldReportingProcessor instance.
   *
   * @param {ILogger} logger - Logger instance for logging operations.
   * @param {INativeYieldAutomationMetricsUpdater} metricsUpdater - Service for updating operation mode metrics.
   * @param {IOperationModeMetricsRecorder} operationModeMetricsRecorder - Service for recording operation mode metrics from transaction receipts.
   * @param {IYieldManager<TransactionReceipt>} yieldManagerContractClient - Client for interacting with YieldManager contracts.
   * @param {ILazyOracle<TransactionReceipt>} lazyOracleContractClient - Client for waiting on LazyOracle events.
   * @param {ILineaRollupYieldExtension<TransactionReceipt>} lineaRollupYieldExtensionClient - Client for interacting with LineaRollupYieldExtension contracts.
   * @param {ILidoAccountingReportClient} lidoAccountingReportClient - Client for submitting Lido accounting reports.
   * @param {IBeaconChainStakingClient} beaconChainStakingClient - Client for managing beacon chain staking operations.
   * @param {Address} yieldProvider - The yield provider address to process.
   * @param {Address} l2YieldRecipient - The L2 yield recipient address for yield reporting.
   */
  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly operationModeMetricsRecorder: IOperationModeMetricsRecorder,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly lineaRollupYieldExtensionClient: ILineaRollupYieldExtension<TransactionReceipt>,
    private readonly lidoAccountingReportClient: ILidoAccountingReportClient,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly yieldProvider: Address,
    private readonly l2YieldRecipient: Address,
  ) {}

  /**
   * Executes one processing cycle:
   * - Waits for the next `VaultsReportDataUpdated` event **or** a timeout, whichever happens first.
   * - Once triggered, runs the main processing logic (`_process()`).
   * - Always cleans up the event watcher afterward.
   * Records operation mode trigger metrics and execution duration metrics.
   *
   * @returns {Promise<void>} A promise that resolves when the processing cycle completes.
   */
  public async process(): Promise<void> {
    const triggerEvent = await this.lazyOracleContractClient.waitForVaultsReportDataUpdatedEvent();
    this.metricsUpdater.incrementOperationModeTrigger(OperationMode.YIELD_REPORTING_MODE, triggerEvent.result);
    const startedAt = performance.now();
    await this._process();
    const durationMs = performance.now() - startedAt;
    this.metricsUpdater.recordOperationModeDuration(OperationMode.YIELD_REPORTING_MODE, msToSeconds(durationMs));
  }

  /**
   * Orchestrates a single Native Yield reporting cycle.
   *
   * High-level flow
   *  1) Read current state:
   *     - Fetch latest Lido report params
   *     - Determine rebalance direction/amount (stake vs. unstake vs. no-op)
   *     - Dry-run report submission to avoid reverting later
   *  2) Front-running / safety:
   *     - If we're in DEFICIT (need to UNSTAKE), pause staking up-front to
   *       prevent new deposits worsening the shortfall while we act
   *  3) Primary action:
   *     - Perform the primary rebalance and (if possible) submit the report
   *  4) Mid-cycle drift fix:
   *     - If we *started* needing STAKE but, during processing, external flows
   *       (e.g., bridge withdrawals) flipped us into DEFICIT, perform an
   *       *amendment* UNSTAKE to restore reserve targets
   *  5) Resume normal operations:
   *     - If we started in EXCESS (STAKE) and did *not* end in DEFICIT, unpause staking
   *  6) Beacon chain withdrawals:
   *     - Defer actual validator withdrawals to the very end since fulfillment
   *       extends beyond this method’s synchronous runtime
   *
   * Invariants and intent:
   *  - Never leave staking enabled while in a known deficit
   *  - Prefer to submit the accounting report in the same cycle as the primary rebalance
   *  - Make at most one amendment pass to handle mid-cycle state changes
   *
   * Side effects:
   *  - May pause/unpause staking
   *  - May submit Lido vault report
   *  - May stake/unstake on the vault
   *  - May queue beacon-chain withdrawals
   */
  private async _process(): Promise<void> {
    // Fetch initial data
    this.vault = await this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider);
    await this.lidoAccountingReportClient.getLatestSubmitVaultReportParams(this.vault);
    const [initialRebalanceRequirements, isSimulateSubmitLatestVaultReportSuccessful] = await Promise.all([
      this.yieldManagerContractClient.getRebalanceRequirements(),
      this.lidoAccountingReportClient.isSimulateSubmitLatestVaultReportSuccessful(this.vault),
    ]);
    this.logger.info(
      `_process - Initial data fetch: initialRebalanceRequirements=${JSON.stringify(initialRebalanceRequirements, bigintReplacer, 2)} isSimulateSubmitLatestVaultReportSuccessful=${isSimulateSubmitLatestVaultReportSuccessful}`,
    );

    // If we begin in DEFICIT, freeze beacon chain deposits to prevent further exacerbation
    if (initialRebalanceRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
      await attempt(
        this.logger,
        () => this.yieldManagerContractClient.pauseStakingIfNotAlready(this.yieldProvider),
        "_process - pause staking failed (tolerated)",
      );
    }

    // Do primary rebalance +/- report submission
    await this._handleRebalance(initialRebalanceRequirements, isSimulateSubmitLatestVaultReportSuccessful);

    const postReportRebalanceRequirements = await this.yieldManagerContractClient.getRebalanceRequirements();
    this.logger.info(
      `_process - Post rebalance data fetch: postReportRebalanceRequirements=${JSON.stringify(postReportRebalanceRequirements, bigintReplacer, 2)}`,
    );

    // Mid-cycle drift check:
    // If we *started* with EXCESS (STAKE) but external flows flipped us to DEFICIT,
    // immediately correct with a targeted UNSTAKE amendment.
    if (initialRebalanceRequirements.rebalanceDirection === RebalanceDirection.STAKE) {
      if (postReportRebalanceRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
        await this._handleUnstakingRebalance(postReportRebalanceRequirements.rebalanceAmount, false);
      } else {
        await attempt(
          this.logger,
          () => this.yieldManagerContractClient.unpauseStakingIfNotAlready(this.yieldProvider),
          "_process - unpause staking failed (tolerated)",
        );
      }
    }

    // Beacon-chain withdrawals are last:
    // These have fulfillment latency beyond this method; queue them after local state is stable.
    const beaconChainWithdrawalRequirements = await this.yieldManagerContractClient.getRebalanceRequirements();
    this.logger.info(
      `_process - Beacon chain withdrawal data fetch: beaconChainWithdrawalRequirements=${JSON.stringify(beaconChainWithdrawalRequirements, bigintReplacer, 2)}`,
    );
    if (beaconChainWithdrawalRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
      await this.beaconChainStakingClient.submitWithdrawalRequestsToFulfilAmount(
        beaconChainWithdrawalRequirements.rebalanceAmount,
      );
    }
  }

  /**
   * Handles rebalancing operations based on rebalance requirements.
   * Routes to appropriate handler based on rebalance direction:
   * - NONE: No-op if report simulation fails, otherwise simple submit report
   * - STAKE: Handles staking rebalance (surplus)
   * - UNSTAKE: Handles unstaking rebalance (deficit)
   *
   * @param {RebalanceRequirement} rebalanceRequirements - The rebalance requirements containing direction and amount.
   * @param {boolean} isSimulateSubmitLatestVaultReportSuccessful - Whether the vault report submission simulation was successful.
   * @returns {Promise<void>} A promise that resolves when rebalancing is handled.
   */
  private async _handleRebalance(
    rebalanceRequirements: RebalanceRequirement,
    isSimulateSubmitLatestVaultReportSuccessful: boolean,
  ): Promise<void> {
    if (rebalanceRequirements.rebalanceDirection === RebalanceDirection.NONE) {
      // No-op
      if (!isSimulateSubmitLatestVaultReportSuccessful) {
        this.logger.info("_handleRebalance - no-op");
        return;
        // Simple submit report
      } else {
        this.logger.info("_handleRebalance - no rebalance pathway, calling _handleSubmitLatestVaultReport");
        await this._handleSubmitLatestVaultReport();
        return;
      }
    } else if (rebalanceRequirements.rebalanceDirection === RebalanceDirection.STAKE) {
      await this._handleStakingRebalance(
        rebalanceRequirements.rebalanceAmount,
        isSimulateSubmitLatestVaultReportSuccessful,
      );
    } else {
      await this._handleUnstakingRebalance(
        rebalanceRequirements.rebalanceAmount,
        isSimulateSubmitLatestVaultReportSuccessful,
      );
    }
  }

  /**
   * Handles staking rebalance operations when there is a reserve surplus.
   * Rebalance first - tolerate failures because fresh vault report should not be blocked.
   * Only do YieldManager->YieldProvider, if L1MessageService->YieldManager succeeded.
   * Submit report last.
   *
   * Assumptions:
   * - i.) We count rebalance once funds have been moved away from the L1MessageService
   * - ii.) Only the initial rebalance will call this fn
   *
   * @param {bigint} rebalanceAmount - The amount to rebalance in wei.
   * @param {boolean} isSimulateSubmitLatestVaultReportSuccessful - Whether the vault report submission simulation was successful.
   * @returns {Promise<void>} A promise that resolves when staking rebalance is handled.
   */
  // Surplus
  private async _handleStakingRebalance(
    rebalanceAmount: bigint,
    isSimulateSubmitLatestVaultReportSuccessful: boolean,
  ): Promise<void> {
    this.logger.info(`_handleStakingRebalance - reserve surplus, rebalanceAmount=${rebalanceAmount}`);
    // Rebalance first - tolerate failures because fresh vault report should not be blocked
    const transferFundsForNativeYieldResult = await attempt(
      this.logger,
      () => this.lineaRollupYieldExtensionClient.transferFundsForNativeYield(rebalanceAmount),
      "_handleStakingRebalance - transferFundsForNativeYield failed (tolerated)",
    );
    // Only do YieldManager->YieldProvider, if L1MessageService->YieldManager succeeded
    if (transferFundsForNativeYieldResult.isOk()) {
      // Assumptions
      // i.) We count rebalance once funds have been moved away from the L1MessageService
      // ii.) Only the initial rebalance will call this fn
      this.metricsUpdater.recordRebalance(RebalanceDirection.STAKE, weiToGweiNumber(rebalanceAmount));
      const transferFundsResult = await attempt(
        this.logger,
        () => this.yieldManagerContractClient.fundYieldProvider(this.yieldProvider, rebalanceAmount),
        "_handleStakingRebalance - fundYieldProvider failed (tolerated)",
      );
      await this.operationModeMetricsRecorder.recordTransferFundsMetrics(this.yieldProvider, transferFundsResult);
    }

    // Submit report last
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      this.logger.info("_handleStakingRebalance calling _handleSubmitLatestVaultReport");
      await this._handleSubmitLatestVaultReport();
    }
  }

  /**
   * Handles unstaking rebalance operations when there is a reserve deficit.
   * Submit report first, then perform rebalance.
   *
   * @param {bigint} rebalanceAmount - The amount to rebalance in wei.
   * @param {boolean} isSimulateSubmitLatestVaultReportSuccessful - Whether the vault report submission simulation was successful.
   * @returns {Promise<void>} A promise that resolves when unstaking rebalance is handled.
   */
  // Deficit
  private async _handleUnstakingRebalance(
    rebalanceAmount: bigint,
    isSimulateSubmitLatestVaultReportSuccessful: boolean,
  ): Promise<void> {
    // Submit report first
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      this.logger.info("_handleUnstakingRebalance calling _handleSubmitLatestVaultReport");
      await this._handleSubmitLatestVaultReport();
    }

    this.logger.info(`_handleUnstakingRebalance - reserve deficit, rebalanceAmount=${rebalanceAmount}`);
    // Then perform rebalance
    const withdrawalResult = await attempt(
      this.logger,
      () =>
        this.yieldManagerContractClient.safeAddToWithdrawalReserveIfAboveThreshold(this.yieldProvider, rebalanceAmount),
      "_handleUnstakingRebalance - safeAddToWithdrawalReserveIfAboveThreshold failed (tolerated)",
    );
    await this.operationModeMetricsRecorder.recordSafeWithdrawalMetrics(this.yieldProvider, withdrawalResult);
  }

  /**
   * @notice Submits the latest vault report and then reports yield to the yield manager.
   * @dev Uses `tryResult` to safely handle failures without throwing.
   *      - If submitting the vault report fails, the function logs the error and exits early,
   *        since reporting yield without a fresh vault report would be invalid.
   *      - If yield reporting fails, the function logs the error but continues without rethrowing,
   *        allowing the caller or scheduler to proceed without interruption.
   * @dev We tolerate report submission errors because they should not block rebalances
   * @returns {Promise<void>} A promise that resolves when the vault report and yield report are submitted (or early returns on failure).
   */
  private async _handleSubmitLatestVaultReport() {
    // First call: submit vault report
    const vaultResult = await attempt(
      this.logger,
      () => this.lidoAccountingReportClient.submitLatestVaultReport(this.vault),
      "_handleSubmitLatestVaultReport: submitLatestVaultReport failed; skipping yield report",
    );
    // Early return, no point reporting Linea yield without a new vault report beforehand
    if (vaultResult.isErr()) {
      return;
    } else {
      this.metricsUpdater.incrementLidoVaultAccountingReport(this.vault);
    }

    // Second call: report yield
    const yieldResult = await attempt(
      this.logger,
      () => this.yieldManagerContractClient.reportYield(this.yieldProvider, this.l2YieldRecipient),
      "_handleSubmitLatestVaultReport - submitLatestVaultReport succeeded but reportYield failed",
    );
    if (yieldResult.isErr()) {
      return;
    } else {
      await this.operationModeMetricsRecorder.recordReportYieldMetrics(this.yieldProvider, yieldResult);
    }

    // Both calls succeeded
    this.logger.info("_handleSubmitLatestVaultReport: vault report + yield report succeeded");
    return;
  }
}
