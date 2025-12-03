import {
  ILogger,
  weiToGweiNumber,
  IBeaconNodeAPIClient,
  PendingPartialWithdrawal,
} from "@consensys/linea-shared-utils";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient.js";
import { INativeYieldAutomationMetricsUpdater } from "../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { IYieldManager } from "../core/clients/contracts/IYieldManager.js";
import { IVaultHub } from "../core/clients/contracts/IVaultHub.js";
import { ValidatorBalance } from "../core/entities/ValidatorBalance.js";
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
   * @param {IBeaconNodeAPIClient} beaconNodeApiClient - Client for retrieving pending partial withdrawals from beacon chain.
   */
  constructor(
    private readonly logger: ILogger,
    private readonly validatorDataClient: IValidatorDataClient,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly vaultHubContractClient: IVaultHub<TransactionReceipt>,
    private readonly yieldProvider: Address,
    private readonly beaconNodeApiClient: IBeaconNodeAPIClient,
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
    // Fetch validator data and pending withdrawals in parallel at the start
    // Use Promise.allSettled to handle failures gracefully
    const fetchResults = await Promise.allSettled([
      this.validatorDataClient.getActiveValidators(),
      this.beaconNodeApiClient.getPendingPartialWithdrawals(),
    ]);

    const allValidators = fetchResults[0].status === "fulfilled" ? fetchResults[0].value : undefined;
    const pendingWithdrawalsQueue = fetchResults[1].status === "fulfilled" ? fetchResults[1].value : undefined;

    // Log fetch failures if any
    if (fetchResults[0].status === "rejected") {
      this.logger.error("Failed to fetch active validators", { error: fetchResults[0].reason });
    }
    if (fetchResults[1].status === "rejected") {
      this.logger.error("Failed to fetch pending partial withdrawals", { error: fetchResults[1].reason });
    }

    // Fetch vault address once before parallel execution
    // If vault fetch fails, we'll skip vault-dependent metrics but still update others
    let vault: Address | undefined;
    try {
      vault = await this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider);
    } catch (error) {
      this.logger.error("Failed to fetch vault address, skipping vault-dependent metrics", { error });
    }

    // Update metrics in parallel for efficiency
    // Use Promise.allSettled to ensure all updates are attempted even if one fails
    const updatePromises: Promise<void>[] = [
      this._updatePendingPartialWithdrawalsGauge(allValidators, pendingWithdrawalsQueue),
    ];

    // Only add vault-dependent metrics if we successfully fetched the vault address
    if (vault !== undefined) {
      updatePromises.push(
        this._updateYieldReportedCumulativeGauge(vault),
        this._updateLastVaultReportTimestampGauge(vault),
      );
    }

    const results = await Promise.allSettled(updatePromises);

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
   * @param {ValidatorBalance[] | undefined} allValidators - Array of active validators, or undefined.
   * @param {PendingPartialWithdrawal[] | undefined} pendingWithdrawalsQueue - Array of pending partial withdrawals, or undefined.
   * @returns {Promise<void>} A promise that resolves when the gauge is updated (or silently returns if validator data is unavailable).
   */
  private async _updatePendingPartialWithdrawalsGauge(
    allValidators: ValidatorBalance[] | undefined,
    pendingWithdrawalsQueue: PendingPartialWithdrawal[] | undefined,
  ): Promise<void> {
    const joinedValidatorList = this.validatorDataClient.joinValidatorsWithPendingWithdrawals(
      allValidators,
      pendingWithdrawalsQueue,
    );
    if (joinedValidatorList === undefined) {
      return;
    }
    const totalPendingPartialWithdrawalsWei =
      this.validatorDataClient.getTotalPendingPartialWithdrawalsWei(joinedValidatorList);
    this.metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei(weiToGweiNumber(totalPendingPartialWithdrawalsWei));
  }

  /**
   * Updates the cumulative yield reported gauge metric from the YieldManager contract.
   *
   * @param {Address} vault - The vault address to use for the metric.
   * @returns {Promise<void>} A promise that resolves when the gauge is updated.
   */
  private async _updateYieldReportedCumulativeGauge(vault: Address): Promise<void> {
    const yieldProviderData = await this.yieldManagerContractClient.getYieldProviderData(this.yieldProvider);
    this.metricsUpdater.setYieldReportedCumulative(vault, weiToGweiNumber(yieldProviderData.yieldReportedCumulative));
  }

  /**
   * Updates the last vault report timestamp gauge metric from the VaultHub contract.
   *
   * @param {Address} vault - The vault address to use for the metric.
   * @returns {Promise<void>} A promise that resolves when the gauge is updated.
   */
  private async _updateLastVaultReportTimestampGauge(vault: Address): Promise<void> {
    const timestamp = await this.vaultHubContractClient.getLatestVaultReportTimestamp(vault);
    this.metricsUpdater.setLastVaultReportTimestamp(vault, Number(timestamp));
  }
}
