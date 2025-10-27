import { Address, TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { bigintReplacer, ILogger, attempt, msToSeconds } from "@consensys/linea-shared-utils";
import { wait } from "@consensys/linea-sdk";
import { ILazyOracle } from "../../core/clients/contracts/ILazyOracle.js";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient.js";
import { RebalanceDirection, RebalanceRequirement } from "../../core/entities/RebalanceRequirement.js";
import { ILineaRollupYieldExtension } from "../../core/clients/contracts/ILineaRollupYieldExtension.js";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { OperationMode } from "../../core/enums/OperationModeEnums.js";

// FIRST PRIORITY FOR UNIT TESTING
export class YieldReportingProcessor implements IOperationModeProcessor {
  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly lineaRollupYieldExtensionClient: ILineaRollupYieldExtension<TransactionReceipt>,
    private readonly lidoAccountingReportClient: ILidoAccountingReportClient,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly maxInactionMs: number,
    private readonly yieldProvider: Address,
    private readonly l2YieldRecipient: Address,
  ) {}

  /**
   * Executes one processing cycle:
   * - Waits for the next `VaultsReportDataUpdated` event **or** a timeout, whichever happens first.
   * - Once triggered, runs the main processing logic (`_process()`).
   * - Always cleans up the event watcher afterward.
   */
  public async process(): Promise<void> {
    const { unwatch, waitForEvent } = await this.lazyOracleContractClient.waitForVaultsReportDataUpdatedEvent();
    try {
      this.logger.info(`process - Waiting for VaultsReportDataUpdated event vs timeout race, timeout=${this.maxInactionMs}ms`)
      // Race: event vs. timeout
      const winner = await Promise.race([
        waitForEvent.then(() => "event" as const),
        wait(this.maxInactionMs).then(() => "timeout" as const),
      ]);
      this.logger.info(
        `process - race won by ${winner === "timeout" ?  `time out after ${this.maxInactionMs}ms` : "VaultsReportDataUpdated event"}`,
      );
      const startedAt = performance.now();
      await this._process();
      const durationMs = performance.now() - startedAt;
      this.metricsUpdater.recordOperationModeDuration(OperationMode.YIELD_REPORTING_MODE, msToSeconds(durationMs));
    } finally {
      // clean up watcher
      this.logger.debug("Cleaning up VaultsReportDataUpdated event watcher")
      unwatch();
    }
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
    await this.lidoAccountingReportClient.getLatestSubmitVaultReportParams();
    const [initialRebalanceRequirements, isSimulateSubmitLatestVaultReportSuccessful] = await Promise.all([
      this.yieldManagerContractClient.getRebalanceRequirements(),
      this.lidoAccountingReportClient.isSimulateSubmitLatestVaultReportSuccessful(),
    ]);
    this.logger.info(
      `_process - Initial data fetch: initialRebalanceRequirements=${JSON.stringify(initialRebalanceRequirements, bigintReplacer, 2)} isSimulateSubmitLatestVaultReportSuccessful=${isSimulateSubmitLatestVaultReportSuccessful}`,
    );

    // If we begin in DEFICIT, freeze beacon chain deposits to prevent further exacerbation
    if (initialRebalanceRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
      await attempt(this.logger, () => this.yieldManagerContractClient.pauseStakingIfNotAlready(this.yieldProvider), "_process - pause staking failed (tolerated)");
    }

    // Do primary rebalance +/- report submission
    await this._handleRebalance(initialRebalanceRequirements, isSimulateSubmitLatestVaultReportSuccessful);

    const postReportRebalanceRequirements = await this.yieldManagerContractClient.getRebalanceRequirements();
    this.logger.info(
      `_process - Post rebalance data fetch: postReportRebalanceRequirements=${JSON.stringify(postReportRebalanceRequirements, bigintReplacer, 2)}`
    );

    // Mid-cycle drift check:
    // If we *started* with EXCESS (STAKE) but external flows flipped us to DEFICIT,
    // immediately correct with a targeted UNSTAKE amendment.
    if (initialRebalanceRequirements.rebalanceDirection === RebalanceDirection.STAKE) {
      if (postReportRebalanceRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
        await this._handleUnstakingRebalance(postReportRebalanceRequirements.rebalanceAmount, false);
      } else {
        await attempt(this.logger, () => this.yieldManagerContractClient.unpauseStakingIfNotAlready(this.yieldProvider), "_process - unpause staking failed (tolerated)");
      }
    }

    // Beacon-chain withdrawals are last:
    // These have fulfillment latency beyond this method; queue them after local state is stable.
    const beaconChainWithdrawalRequirements = await this.yieldManagerContractClient.getRebalanceRequirements();
    this.logger.info(
      `_process - Beacon chain withdrawal data fetch: beaconChainWithdrawalRequirements=${JSON.stringify(beaconChainWithdrawalRequirements, bigintReplacer, 2)}`
    );
    if (beaconChainWithdrawalRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
      await this.beaconChainStakingClient.submitWithdrawalRequestsToFulfilAmount(
        beaconChainWithdrawalRequirements.rebalanceAmount,
      );
    }
  }

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

  // Surplus
  private async _handleStakingRebalance(
    rebalanceAmount: bigint,
    isSimulateSubmitLatestVaultReportSuccessful: boolean,
  ): Promise<void> {
      this.logger.info(`_handleStakingRebalance - reserve surplus, rebalanceAmount=${rebalanceAmount}`);
    // Rebalance first - tolerate failures because fresh vault report should not be blocked
    const transferFundsForNativeYieldResult = await attempt(this.logger, () => this.lineaRollupYieldExtensionClient.transferFundsForNativeYield(rebalanceAmount), "_handleStakingRebalance - transferFundsForNativeYield failed (tolerated)");
    // Only do YieldManager->YieldProvider, if L1MessageService->YieldManager succeeded
    if (transferFundsForNativeYieldResult.isOk()) {
      await attempt(this.logger, () => this.yieldManagerContractClient.fundYieldProvider(this.yieldProvider, rebalanceAmount), "_handleStakingRebalance - fundYieldProvider failed (tolerated)");
    }

    // Submit report last
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      this.logger.info("_handleStakingRebalance calling _handleSubmitLatestVaultReport");
      await this._handleSubmitLatestVaultReport();
    }
  }

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
    await attempt(this.logger, () => this.yieldManagerContractClient.safeAddToWithdrawalReserveIfAboveThreshold(this.yieldProvider, rebalanceAmount), "_handleUnstakingRebalance - safeAddToWithdrawalReserveIfAboveThreshold failed (tolerated)");
  }

  /**
   * @notice Submits the latest vault report and then reports yield to the yield manager.
   * @dev Uses `tryResult` to safely handle failures without throwing.
   *      - If submitting the vault report fails, the function logs the error and exits early,
   *        since reporting yield without a fresh vault report would be invalid.
   *      - If yield reporting fails, the function logs the error but continues without rethrowing,
   *        allowing the caller or scheduler to proceed without interruption.
   * @dev We tolerate report submission errors because they should not block rebalances
   */
  private async _handleSubmitLatestVaultReport() {
      // First call: submit vault report
      const vaultResult = await attempt(this.logger, () => this.lidoAccountingReportClient.submitLatestVaultReport(), "_handleSubmitLatestVaultReport: submitLatestVaultReport failed; skipping yield report");
      // Early return, no point reporting Linea yield without a new vault report beforehand
      if (vaultResult.isErr()) {
        return;
      }

      // Second call: report yield
      const yieldResult = await attempt(this.logger, () => this.yieldManagerContractClient.reportYield(this.yieldProvider, this.l2YieldRecipient), "_handleSubmitLatestVaultReport - submitLatestVaultReport succeeded but reportYield failed");
      if (yieldResult.isErr()) {
        return;
      }

      // Both calls succeeded
      this.logger.info("_handleSubmitLatestVaultReport: vault report + yield report succeeded");
      return yieldResult;
  }
}
