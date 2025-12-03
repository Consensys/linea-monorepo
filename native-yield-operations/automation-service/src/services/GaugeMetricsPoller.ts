import {
  ILogger,
  weiToGweiNumber,
  IBeaconNodeAPIClient,
  PendingPartialWithdrawal,
  PendingDeposit,
  get0x02WithdrawalCredentials,
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
    // Fetch validator data, pending withdrawals, pending deposits, vault address, and yield provider data in parallel
    // Use Promise.allSettled to handle failures gracefully
    const fetchResults = await Promise.allSettled([
      this.validatorDataClient.getActiveValidators(),
      this.validatorDataClient.getExitingValidators(),
      this.beaconNodeApiClient.getPendingPartialWithdrawals(),
      this.beaconNodeApiClient.getPendingDeposits(),
      this.yieldManagerContractClient.getLidoStakingVaultAddress(this.yieldProvider),
      this.yieldManagerContractClient.getYieldProviderData(this.yieldProvider),
    ]);

    const allValidators = fetchResults[0].status === "fulfilled" ? fetchResults[0].value : undefined;
    const exitingValidators = fetchResults[1].status === "fulfilled" ? fetchResults[1].value : undefined;
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    void exitingValidators; // TODO: Use exitingValidators for metrics in future update
    const pendingWithdrawalsQueue = fetchResults[2].status === "fulfilled" ? fetchResults[2].value : undefined;
    const pendingDeposits = fetchResults[3].status === "fulfilled" ? fetchResults[3].value : undefined;
    const vault = fetchResults[4].status === "fulfilled" ? fetchResults[4].value : undefined;
    const yieldProviderData = fetchResults[5].status === "fulfilled" ? fetchResults[5].value : undefined;

    // Log fetch failures if any
    if (fetchResults[0].status === "rejected") {
      this.logger.error("Failed to fetch active validators", { error: fetchResults[0].reason });
    }
    if (fetchResults[1].status === "rejected") {
      this.logger.error("Failed to fetch exiting validators", { error: fetchResults[1].reason });
    }
    if (fetchResults[2].status === "rejected") {
      this.logger.error("Failed to fetch pending partial withdrawals", { error: fetchResults[2].reason });
    }
    if (fetchResults[3].status === "rejected") {
      this.logger.error("Failed to fetch pending deposits", { error: fetchResults[3].reason });
    }
    if (fetchResults[4].status === "rejected") {
      this.logger.error("Failed to fetch vault address, skipping vault-dependent metrics", {
        error: fetchResults[4].reason,
      });
    }
    if (fetchResults[5].status === "rejected") {
      this.logger.error("Failed to fetch yield provider data", { error: fetchResults[5].reason });
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
      updatePromises.push(this._updatePendingDepositsQueueGauge(vault, pendingDeposits));
      updatePromises.push(this._updateTotalPendingDepositsGauge(vault, pendingDeposits));
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
          "pending deposits queue",
          "total pending deposits",
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
   * @returns {Promise<void>} A promise that resolves when the gauge is updated (or returns early with a warning if validator data is unavailable).
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
      this.logger.warn("Skipping total pending partial withdrawals gauge update: validator data unavailable");
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
   * @returns {Promise<void>} A promise that resolves when the gauges are updated (or returns early with a warning if validator data is unavailable).
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
      this.logger.warn("Skipping pending partial withdrawals queue gauge update: aggregated withdrawals unavailable");
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
   * @returns {Promise<void>} A promise that resolves when the gauge is updated (or returns early with a warning if validator data is unavailable).
   */
  private async _updateTotalValidatorBalanceGauge(allValidators: ValidatorBalance[] | undefined): Promise<void> {
    const totalValidatorBalanceGwei = this.validatorDataClient.getTotalValidatorBalanceGwei(allValidators);
    if (totalValidatorBalanceGwei === undefined) {
      this.logger.warn("Skipping total validator balance gauge update: validator balance unavailable");
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

  /**
   * Updates the pending deposits queue gauge metrics.
   * Filters pending deposits by withdrawal credentials matching the vault address,
   * then updates metrics for each matching deposit.
   *
   * @param {Address} vault - The vault address to use for filtering deposits.
   * @param {PendingDeposit[] | undefined} pendingDeposits - Array of pending deposits, or undefined.
   * @returns {Promise<void>} A promise that resolves when the gauges are updated (or returns early with a warning if deposit data is unavailable).
   */
  private async _updatePendingDepositsQueueGauge(
    vault: Address,
    pendingDeposits: PendingDeposit[] | undefined,
  ): Promise<void> {
    if (pendingDeposits === undefined) {
      this.logger.warn("Skipping pending deposits queue gauge update: pending deposits data unavailable");
      return;
    }

    const vaultWithdrawalCredentials = get0x02WithdrawalCredentials(vault);
    const matchingDeposits = pendingDeposits.filter(
      (deposit) => deposit.withdrawal_credentials.toLowerCase() === vaultWithdrawalCredentials.toLowerCase(),
    );

    for (const deposit of matchingDeposits) {
      // amount is already in gwei (as number)
      this.metricsUpdater.setPendingDepositQueueAmountGwei(deposit.pubkey as Hex, deposit.slot, deposit.amount);
    }
  }

  /**
   * Updates the total pending deposits gauge metric.
   * Filters pending deposits by withdrawal credentials matching the vault address,
   * then sums and updates the total metric.
   *
   * @param {Address} vault - The vault address to use for filtering deposits.
   * @param {PendingDeposit[] | undefined} pendingDeposits - Array of pending deposits, or undefined.
   * @returns {Promise<void>} A promise that resolves when the gauge is updated (or returns early with a warning if deposit data is unavailable).
   */
  private async _updateTotalPendingDepositsGauge(
    vault: Address,
    pendingDeposits: PendingDeposit[] | undefined,
  ): Promise<void> {
    if (pendingDeposits === undefined) {
      this.logger.warn("Skipping total pending deposits gauge update: pending deposits data unavailable");
      return;
    }

    const vaultWithdrawalCredentials = get0x02WithdrawalCredentials(vault);
    const matchingDeposits = pendingDeposits.filter(
      (deposit) => deposit.withdrawal_credentials.toLowerCase() === vaultWithdrawalCredentials.toLowerCase(),
    );

    const totalAmount = matchingDeposits.reduce((sum, deposit) => sum + deposit.amount, 0);
    // amount is already in gwei (as number)
    this.metricsUpdater.setLastTotalPendingDepositGwei(totalAmount);
  }
}
