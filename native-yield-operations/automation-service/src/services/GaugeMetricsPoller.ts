import { ILogger, weiToGweiNumber } from "@consensys/linea-shared-utils";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient.js";
import { INativeYieldAutomationMetricsUpdater } from "../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { IYieldManager } from "../core/clients/contracts/IYieldManager.js";
import { IVaultHub } from "../core/clients/contracts/IVaultHub.js";
import { Address, TransactionReceipt } from "viem";
import { IGaugeMetricsPoller } from "../core/services/IGaugeMetricsPoller.js";

/**
 * Polls various data sources and updates gauge metrics.
 * Handles updating metrics like total pending partial withdrawals and cumulative yield reported.
 */
export class GaugeMetricsPoller implements IGaugeMetricsPoller {
  /**
   * Creates a new GaugeMetricsPoller instance.
   *
   * @param {ILogger} logger - Logger instance for logging errors.
   * @param {IValidatorDataClient} validatorDataClient - Client for retrieving validator data.
   * @param {INativeYieldAutomationMetricsUpdater} metricsUpdater - Service for updating metrics.
   * @param {IYieldManager<TransactionReceipt>} yieldManagerContractClient - Client for reading yield provider data from YieldManager contract.
   * @param {IVaultHub<TransactionReceipt>} vaultHubContractClient - Client for reading vault data from VaultHub contract.
   * @param {Address} yieldProvider - The yield provider address.
   */
  constructor(
    private readonly logger: ILogger,
    private readonly validatorDataClient: IValidatorDataClient,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly vaultHubContractClient: IVaultHub<TransactionReceipt>,
    private readonly yieldProvider: Address,
  ) {}

  /**
   * Polls data sources and updates gauge metrics.
   * Updates both validator-related metrics and contract-based metrics in parallel.
   * Gracefully handles failures - if one metric update fails, others will still proceed.
   * Errors are logged but do not propagate to prevent breaking the polling loop.
   *
   * @returns {Promise<void>} A promise that resolves when gauge metrics are updated.
   */
  async poll(): Promise<void> {
    // Update metrics in parallel for efficiency
    // Use Promise.allSettled to ensure all updates are attempted even if one fails
    const results = await Promise.allSettled([
      this._updatePendingPartialWithdrawalsGauge(),
      this._updateYieldReportedCumulativeGauge(),
      this._updateLastVaultReportTimestampGauge(),
    ]);

    // Log any failures
    results.forEach((result, index) => {
      if (result.status === "rejected") {
        const metricNames = ["pending partial withdrawals", "yield reported cumulative", "last vault report timestamp"];
        const metricName = metricNames[index] || "unknown";
        this.logger.error(`Failed to update ${metricName} gauge metric`, { error: result.reason });
      }
    });
  }

  /**
   * Updates the total pending partial withdrawals gauge metric.
   * Follows the same pattern as BeaconChainStakingClient.submitWithdrawalRequestsToFulfilAmount.
   *
   * @returns {Promise<void>} A promise that resolves when the gauge is updated (or silently returns if validator data is unavailable).
   */
  private async _updatePendingPartialWithdrawalsGauge(): Promise<void> {
    const sortedValidatorList = await this.validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending();
    if (sortedValidatorList === undefined) {
      return;
    }
    const totalPendingPartialWithdrawalsWei =
      this.validatorDataClient.getTotalPendingPartialWithdrawalsWei(sortedValidatorList);
    this.metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei(weiToGweiNumber(totalPendingPartialWithdrawalsWei));
  }

  /**
   * Updates the cumulative yield reported gauge metric from the YieldManager contract.
   *
   * @returns {Promise<void>} A promise that resolves when the gauge is updated.
   */
  private async _updateYieldReportedCumulativeGauge(): Promise<void> {
    const yieldProviderData = await this.yieldManagerContractClient.getYieldProviderData(this.yieldProvider);
    const vault = await this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider);
    this.metricsUpdater.setYieldReportedCumulative(vault, weiToGweiNumber(yieldProviderData.yieldReportedCumulative));
  }

  /**
   * Updates the last vault report timestamp gauge metric from the VaultHub contract.
   *
   * @returns {Promise<void>} A promise that resolves when the gauge is updated.
   */
  private async _updateLastVaultReportTimestampGauge(): Promise<void> {
    const vault = await this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider);
    const timestamp = await this.vaultHubContractClient.getLatestVaultReportTimestamp(vault);
    this.metricsUpdater.setLastVaultReportTimestamp(vault, Number(timestamp));
  }
}
