import {
  ILogger,
  weiToGweiNumber,
  IBeaconNodeAPIClient,
  PendingPartialWithdrawal,
} from "@consensys/linea-shared-utils";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient.js";
import { INativeYieldAutomationMetricsUpdater } from "../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { IYieldManager, YieldProviderData } from "../core/clients/contracts/IYieldManager.js";
import { IVaultHub } from "../core/clients/contracts/IVaultHub.js";
import { ValidatorBalance } from "../core/entities/ValidatorBalance.js";
import { Address, Hex, TransactionReceipt } from "viem";
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
    // Fetch validator data, pending withdrawals, vault address, and yield provider data in parallel
    // Use Promise.allSettled to handle failures gracefully
    const fetchResults = await Promise.allSettled([
      this.validatorDataClient.getActiveValidators(),
      this.beaconNodeApiClient.getPendingPartialWithdrawals(),
      this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider),
      this.yieldManagerContractClient.getYieldProviderData(this.yieldProvider),
    ]);

    const allValidators = fetchResults[0].status === "fulfilled" ? fetchResults[0].value : undefined;
    const pendingWithdrawalsQueue = fetchResults[1].status === "fulfilled" ? fetchResults[1].value : undefined;
    const vault = fetchResults[2].status === "fulfilled" ? fetchResults[2].value : undefined;
    const yieldProviderData = fetchResults[3].status === "fulfilled" ? fetchResults[3].value : undefined;

    // Log fetch failures if any
    if (fetchResults[0].status === "rejected") {
      this.logger.error("Failed to fetch active validators", { error: fetchResults[0].reason });
    }
    if (fetchResults[1].status === "rejected") {
      this.logger.error("Failed to fetch pending partial withdrawals", { error: fetchResults[1].reason });
    }
    if (fetchResults[2].status === "rejected") {
      this.logger.error("Failed to fetch vault address, skipping vault-dependent metrics", {
        error: fetchResults[2].reason,
      });
    }
    if (fetchResults[3].status === "rejected") {
      this.logger.error("Failed to fetch yield provider data", { error: fetchResults[3].reason });
    }

    // Update metrics in parallel for efficiency
    // Use Promise.allSettled to ensure all updates are attempted even if one fails
    const updatePromises: Promise<void>[] = [
      this._updateTotalPendingPartialWithdrawalsGauge(allValidators, pendingWithdrawalsQueue),
      this._updatePendingPartialWithdrawalsQueueGauge(allValidators, pendingWithdrawalsQueue),
      this._updateTotalValidatorBalanceGauge(allValidators),
    ];

    // Only add vault-dependent metrics if we successfully fetched the vault address
    if (vault !== undefined) {
      updatePromises.push(this._updateLastVaultReportTimestampGauge(vault));
      // Only add yield provider data metrics if we successfully fetched the data
      if (yieldProviderData !== undefined) {
        updatePromises.push(
          this._updateYieldReportedCumulativeGauge(vault, yieldProviderData),
          this._updateLstLiabilityPrincipalGauge(vault, yieldProviderData),
        );
      }
    }

    const results = await Promise.allSettled(updatePromises);

    // Log any failures
    results.forEach((result, index) => {
      if (result.status === "rejected") {
        const metricNames = [
          "total pending partial withdrawals",
          "pending partial withdrawals queue",
          "total validator balance",
          "last vault report timestamp",
          "yield reported cumulative",
          "lst liability principal",
        ];
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
  private async _updateTotalPendingPartialWithdrawalsGauge(
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
   * Updates the per-validator pending partial withdrawal queue gauge metrics.
   * Filters and aggregates pending withdrawals by validator_index and withdrawable_epoch,
   * then updates metrics for each aggregated withdrawal.
   *
   * @param {ValidatorBalance[] | undefined} allValidators - Array of active validators, or undefined.
   * @param {PendingPartialWithdrawal[] | undefined} pendingWithdrawalsQueue - Array of pending partial withdrawals, or undefined.
   * @returns {Promise<void>} A promise that resolves when the gauges are updated (or silently returns if validator data is unavailable).
   */
  private async _updatePendingPartialWithdrawalsQueueGauge(
    allValidators: ValidatorBalance[] | undefined,
    pendingWithdrawalsQueue: PendingPartialWithdrawal[] | undefined,
  ): Promise<void> {
    const aggregatedWithdrawals = this.validatorDataClient.getFilteredAndAggregatedPendingWithdrawals(
      allValidators,
      pendingWithdrawalsQueue,
    );
    if (aggregatedWithdrawals === undefined) {
      return;
    }
    for (const withdrawal of aggregatedWithdrawals) {
      const amountGwei = Number(withdrawal.amount);
      this.metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei(
        withdrawal.pubkey as Hex,
        withdrawal.withdrawable_epoch,
        amountGwei,
      );
    }
  }

  /**
   * Updates the total validator balance gauge metric.
   *
   * @param {ValidatorBalance[] | undefined} allValidators - Array of active validators, or undefined.
   * @returns {Promise<void>} A promise that resolves when the gauge is updated (or silently returns if validator data is unavailable).
   */
  private async _updateTotalValidatorBalanceGauge(allValidators: ValidatorBalance[] | undefined): Promise<void> {
    const totalValidatorBalanceGwei = this.validatorDataClient.getTotalValidatorBalanceGwei(allValidators);
    if (totalValidatorBalanceGwei === undefined) {
      return;
    }
    this.metricsUpdater.setLastTotalValidatorBalanceGwei(Number(totalValidatorBalanceGwei));
  }

  /**
   * Updates the cumulative yield reported gauge metric from the YieldManager contract.
   *
   * @param {Address} vault - The vault address to use for the metric.
   * @param {YieldProviderData} yieldProviderData - The yield provider data from the YieldManager contract.
   * @returns {Promise<void>} A promise that resolves when the gauge is updated.
   */
  private async _updateYieldReportedCumulativeGauge(
    vault: Address,
    yieldProviderData: YieldProviderData,
  ): Promise<void> {
    this.metricsUpdater.setYieldReportedCumulative(vault, weiToGweiNumber(yieldProviderData.yieldReportedCumulative));
  }

  /**
   * Updates the LST liability principal gauge metric from the YieldManager contract.
   *
   * @param {Address} vault - The vault address to use for the metric.
   * @param {YieldProviderData} yieldProviderData - The yield provider data from the YieldManager contract.
   * @returns {Promise<void>} A promise that resolves when the gauge is updated.
   */
  private async _updateLstLiabilityPrincipalGauge(vault: Address, yieldProviderData: YieldProviderData): Promise<void> {
    this.metricsUpdater.setLstLiabilityPrincipalGwei(vault, weiToGweiNumber(yieldProviderData.lstLiabilityPrincipal));
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
