import { Address, TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";
import { ILogger } from "ts-libs/linea-shared-utils";
import { wait } from "sdk/sdk-ethers/dist";
import { ILazyOracle, UpdateVaultDataParams } from "../../core/services/contracts/ILazyOracle";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient";
import { RebalanceDirection, RebalanceRequirement } from "../../core/entities/RebalanceRequirement";
import { ILineaRollupYieldExtension } from "../../core/services/contracts/ILineaRollupYieldExtension";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient";

export class YieldReportingOperationModeProcessor implements IOperationModeProcessor {
  constructor(
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly lineaRollupYieldExtensionClient: ILineaRollupYieldExtension<TransactionReceipt>,
    private readonly lidoAccountingReportClient: ILidoAccountingReportClient,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly logger: ILogger,
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
      // Race: event vs. timeout
      const winner = await Promise.race([
        waitForEvent.then(() => "event" as const),
        wait(this.maxInactionMs).then(() => "timeout" as const),
      ]);
      await this._process();
      this.logger.info(`poll(): finished via ${winner}`);
    } finally {
      // clean up watcher
      unwatch();
    }
  }

  // Attempt to submitReport for latest report -> If it works, then we should submit, if simulation fails, then nah.

  /**
   * Main processing loop:
   * 1. Determine whether a new report should be submitted.
   * 2. Decide whether a rebalance is required.
   * 3. If rebalancing, calculate the amount and direction.
   * 4. Execute.
   * 5. Evaluate execution result, amend if needed.
   */
  private async _process(): Promise<void> {
    const vaultAddress = await this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider);
    const latestSubmitVaultReportParams =
      await this.lidoAccountingReportClient.getSubmitVaultReportParams(vaultAddress);
    const [initialRebalanceRequirements, isSimulateSubmitLatestVaultReportSuccessful] = await Promise.all([
      this.yieldManagerContractClient.getRebalanceRequirements(),
      this.lidoAccountingReportClient.isSimulateSubmitLatestVaultReportSuccessful(latestSubmitVaultReportParams),
    ]);

    // If deficit, pause staking first
    if (initialRebalanceRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
      await this.yieldManagerContractClient.pauseStakingIfNotAlready(this.yieldProvider);
    }
    // Do primary rebalance
    await this._handleRebalance(
      initialRebalanceRequirements,
      isSimulateSubmitLatestVaultReportSuccessful,
      latestSubmitVaultReportParams,
    );

    // Do amendments (in-case of front-running, or reportYield leaving system in deficit)
    const postReportRebalanceRequirements = await this.yieldManagerContractClient.getRebalanceRequirements();
    this._handleRebalance(postReportRebalanceRequirements, false, latestSubmitVaultReportParams);
    // Only unpause at end if we had excess to start with, and no deficit after
    if (
      initialRebalanceRequirements.rebalanceDirection === RebalanceDirection.STAKE &&
      postReportRebalanceRequirements.rebalanceDirection !== RebalanceDirection.UNSTAKE
    ) {
      await this.yieldManagerContractClient.unpauseStakingIfNotAlready(this.yieldProvider);
    }

    // Leave beacon chain withdrawals last
    const beaconChainWithdrawalRequirements = await this.yieldManagerContractClient.getRebalanceRequirements();
    if (beaconChainWithdrawalRequirements.rebalanceDirection === RebalanceDirection.UNSTAKE) {
      await this.beaconChainStakingClient.submitWithdrawalRequestsToFulfilAmount(
        beaconChainWithdrawalRequirements.rebalanceAmount,
      );
    }
  }

  private async _handleRebalance(
    rebalanceRequirements: RebalanceRequirement,
    isSimulateSubmitLatestVaultReportSuccessful: boolean,
    latestSubmitVaultReportParams: UpdateVaultDataParams,
  ): Promise<void> {
    if (rebalanceRequirements.rebalanceDirection === RebalanceDirection.NONE) {
      // No-op
      if (!isSimulateSubmitLatestVaultReportSuccessful) {
        return;
        // Simple submit report
      } else {
        await this.lidoAccountingReportClient.submitLatestVaultReport(latestSubmitVaultReportParams);
        await this.yieldManagerContractClient.reportYield(this.yieldProvider, this.l2YieldRecipient);
        return;
      }
    } else if (rebalanceRequirements.rebalanceDirection === RebalanceDirection.STAKE) {
      await this._handleStakingRebalance(
        rebalanceRequirements.rebalanceAmount,
        isSimulateSubmitLatestVaultReportSuccessful,
        latestSubmitVaultReportParams,
      );
    } else {
      await this._handleUnstakingRebalance(
        rebalanceRequirements.rebalanceAmount,
        isSimulateSubmitLatestVaultReportSuccessful,
        latestSubmitVaultReportParams,
      );
    }
  }

  // Surplus
  private async _handleStakingRebalance(
    rebalanceAmount: bigint,
    isSimulateSubmitLatestVaultReportSuccessful: boolean,
    latestSubmitVaultReportParams: UpdateVaultDataParams,
  ): Promise<void> {
    // Rebalance first
    await this.lineaRollupYieldExtensionClient.transferFundsForNativeYield(rebalanceAmount);
    await this.yieldManagerContractClient.fundYieldProvider(this.yieldProvider, rebalanceAmount);
    // Submit report last
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      await this.lidoAccountingReportClient.submitLatestVaultReport(latestSubmitVaultReportParams);
      await this.yieldManagerContractClient.reportYield(this.yieldProvider, this.l2YieldRecipient);
    }
  }

  // Deficit
  private async _handleUnstakingRebalance(
    rebalanceAmount: bigint,
    isSimulateSubmitLatestVaultReportSuccessful: boolean,
    latestSubmitVaultReportParams: UpdateVaultDataParams,
  ): Promise<void> {
    // Submit report first
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      await this.lidoAccountingReportClient.submitLatestVaultReport(latestSubmitVaultReportParams);
      await this.yieldManagerContractClient.reportYield(this.yieldProvider, this.l2YieldRecipient);
    }
    // Then perform rebalance
    // TODO - Skip if only dust will be moved to LineaRollup
    await this.yieldManagerContractClient.safeAddToWithdrawalReserve(this.yieldProvider, rebalanceAmount);
  }
}
