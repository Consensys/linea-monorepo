import { IMetricsService } from "@consensys/linea-shared-utils";
import { LineaNativeYieldAutomationServiceMetrics, RebalanceTypeLabel } from "../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";
import { RebalanceDirection } from "../../core/entities/RebalanceRequirement.js";

const OPERATION_MODE_DURATION_BUCKETS = [1, 5, 10, 30, 60, 120, 300, 600];

export type YieldReportingTriggerLabel = "VaultsReportDataUpdated_event" | "timeout";

export class NativeYieldAutomationMetricsUpdater {
  constructor(private readonly metricsService: IMetricsService<LineaNativeYieldAutomationServiceMetrics>) {
    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.RebalanceAmountTotal,
      "Total rebalance amount between L1MessageService and YieldProvider",
      ["direction", "type"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.ValidatorPartialUnstakeAmountTotal,
      "Total amount partially unstaked per validator",
      ["validator_pubkey"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.ValidatorExitTotal,
      "Total validator exits initiated by automation",
      ["validator_pubkey"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.YieldReportingModeProcessorTriggerTotal,
      "Yield reporting processor activations grouped by trigger",
      ["trigger"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.LidoVaultAccountingReportSubmittedTotal,
      "Accounting reports submitted to Lido per vault",
      ["vault_address"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.ReportYieldTotal,
      "Yield reports submitted to YieldManager per vault",
      ["vault_address"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.ReportYieldAmountTotal,
      "Total yield amount reported per vault",
      ["vault_address"],
    );

    this.metricsService.createGauge(
      LineaNativeYieldAutomationServiceMetrics.CurrentNegativeYieldLastReport,
      "Outstanding negative yield as of the latest report",
      ["vault_address"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.NodeOperatorFeesPaidTotal,
      "Node operator fees paid by automation per vault",
      ["vault_address"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.LiabilitiesPaidTotal,
      "Liabilities paid by automation per vault",
      ["vault_address"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.LidoFeesPaidTotal,
      "Lido fees paid by automation per vault",
      ["vault_address"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.TransactionFeesGwei,
      "Transaction fees paid (gwei) by automation per vault",
      ["vault_address"],
    );

    this.metricsService.createCounter(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionTotal,
      "Operation mode executions grouped by mode",
      ["mode"],
    );

    this.metricsService.createHistogram(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionDurationSeconds,
      OPERATION_MODE_DURATION_BUCKETS,
      "Operation mode execution duration in seconds",
      ["mode"],
    );
  }

  public recordRebalance(
    direction: RebalanceDirection.STAKE | RebalanceDirection.UNSTAKE,
    type: RebalanceTypeLabel,
    amountGwei: number,
  ): void {
    if (amountGwei <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.RebalanceAmountTotal,
      { direction, type },
      amountGwei,
    );
  }

  public addValidatorPartialUnstakeAmount(validatorPubkey: string, amount: number): void {
    if (amount <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.ValidatorPartialUnstakeAmountTotal,
      { validator_pubkey: validatorPubkey },
      amount,
    );
  }

  public incrementValidatorExit(validatorPubkey: string, count: number = 1): void {
    if (count <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.ValidatorExitTotal,
      { validator_pubkey: validatorPubkey },
      count,
    );
  }

  public incrementYieldReportingTrigger(trigger: YieldReportingTriggerLabel): void {
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.YieldReportingModeProcessorTriggerTotal,
      { trigger },
    );
  }

  public incrementLidoVaultAccountingReport(vaultAddress: string): void {
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.LidoVaultAccountingReportSubmittedTotal,
      { vault_address: vaultAddress },
    );
  }

  public incrementReportYield(vaultAddress: string): void {
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.ReportYieldTotal,
      { vault_address: vaultAddress },
    );
  }

  public addReportedYieldAmount(vaultAddress: string, amount: number): void {
    if (amount <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.ReportYieldAmountTotal,
      { vault_address: vaultAddress },
      amount,
    );
  }

  public async setCurrentNegativeYieldLastReport(vaultAddress: string, negativeYield: number): Promise<void> {
    const target = negativeYield;
    const current = (await this.metricsService.getGaugeValue(
      LineaNativeYieldAutomationServiceMetrics.CurrentNegativeYieldLastReport,
      { vault_address: vaultAddress },
    )) ?? 0;

    const delta = target - current;
    if (delta === 0) {
      return;
    }

    if (delta > 0) {
      this.metricsService.incrementGauge(
        LineaNativeYieldAutomationServiceMetrics.CurrentNegativeYieldLastReport,
        { vault_address: vaultAddress },
        delta,
      );
      return;
    }

    this.metricsService.decrementGauge(
      LineaNativeYieldAutomationServiceMetrics.CurrentNegativeYieldLastReport,
      { vault_address: vaultAddress },
      Math.abs(delta),
    );
  }

  public addNodeOperatorFeesPaid(vaultAddress: string, amount: number): void {
    this.incrementVaultAmountCounter(
      LineaNativeYieldAutomationServiceMetrics.NodeOperatorFeesPaidTotal,
      vaultAddress,
      amount,
    );
  }

  public addLiabilitiesPaid(vaultAddress: string, amount: number): void {
    this.incrementVaultAmountCounter(
      LineaNativeYieldAutomationServiceMetrics.LiabilitiesPaidTotal,
      vaultAddress,
      amount,
    );
  }

  public addLidoFeesPaid(vaultAddress: string, amount: number): void {
    this.incrementVaultAmountCounter(
      LineaNativeYieldAutomationServiceMetrics.LidoFeesPaidTotal,
      vaultAddress,
      amount,
    );
  }

  public addTransactionFeesGwei(vaultAddress: string, amount: number): void {
    this.incrementVaultAmountCounter(
      LineaNativeYieldAutomationServiceMetrics.TransactionFeesGwei,
      vaultAddress,
      amount,
    );
  }

  public incrementOperationModeExecution(mode: string): void {
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionTotal,
      { mode },
    );
  }

  public recordOperationModeDuration(mode: string, durationSeconds: number): void {
    if (durationSeconds < 0) return;
    this.metricsService.addValueToHistogram(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionDurationSeconds,
      durationSeconds,
      { mode },
    );
  }

  private incrementVaultAmountCounter(
    metric: LineaNativeYieldAutomationServiceMetrics,
    vaultAddress: string,
    amount: number,
  ): void {
    if (amount <= 0) return;
    this.metricsService.incrementCounter(metric, { vault_address: vaultAddress }, amount);
  }
}
