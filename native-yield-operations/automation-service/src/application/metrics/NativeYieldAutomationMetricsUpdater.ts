import { IMetricsService } from "@consensys/linea-shared-utils";
import {
  LineaNativeYieldAutomationServiceMetrics,
  OperationTrigger,
} from "../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";
import { RebalanceDirection } from "../../core/entities/RebalanceRequirement.js";
import { Address, Hex } from "viem";
import { OperationMode } from "../../core/enums/OperationModeEnums.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";

// Buckets range up to 20 minutes to account for long-running modes.
const OPERATION_MODE_DURATION_BUCKETS = [1, 5, 10, 30, 60, 120, 180, 300, 600, 900, 1200];

// Focused on defining the specific metrics, and methods for updating them
export class NativeYieldAutomationMetricsUpdater implements INativeYieldAutomationMetricsUpdater {
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
      LineaNativeYieldAutomationServiceMetrics.OperationModeTriggerTotal,
      "Operation mode triggers grouped by mode and triggers",
      ["mode", "trigger"],
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

  public recordRebalance(direction: RebalanceDirection.STAKE | RebalanceDirection.UNSTAKE, amountGwei: number): void {
    if (amountGwei <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.RebalanceAmountTotal,
      { direction },
      amountGwei,
    );
  }

  public addValidatorPartialUnstakeAmount(validatorPubkey: Hex, amountGwei: number): void {
    if (amountGwei <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.ValidatorPartialUnstakeAmountTotal,
      { validator_pubkey: validatorPubkey },
      amountGwei,
    );
  }

  public incrementValidatorExit(validatorPubkey: Hex, count: number = 1): void {
    if (count <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.ValidatorExitTotal,
      { validator_pubkey: validatorPubkey },
      count,
    );
  }

  public incrementLidoVaultAccountingReport(vaultAddress: Address): void {
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.LidoVaultAccountingReportSubmittedTotal,
      { vault_address: vaultAddress },
    );
  }

  public incrementReportYield(vaultAddress: Address): void {
    this.metricsService.incrementCounter(LineaNativeYieldAutomationServiceMetrics.ReportYieldTotal, {
      vault_address: vaultAddress,
    });
  }

  public addReportedYieldAmount(vaultAddress: Address, amountGwei: number): void {
    if (amountGwei <= 0) return;
    this.metricsService.incrementCounter(
      LineaNativeYieldAutomationServiceMetrics.ReportYieldAmountTotal,
      { vault_address: vaultAddress },
      amountGwei,
    );
  }

  public async setCurrentNegativeYieldLastReport(vaultAddress: Address, negativeYield: number): Promise<void> {
    this.metricsService.setGauge(
      LineaNativeYieldAutomationServiceMetrics.CurrentNegativeYieldLastReport,
      { vault_address: vaultAddress },
      negativeYield,
    );
  }

  public addNodeOperatorFeesPaid(vaultAddress: Address, amountGwei: number): void {
    this._incrementVaultAmountCounter(
      LineaNativeYieldAutomationServiceMetrics.NodeOperatorFeesPaidTotal,
      vaultAddress,
      amountGwei,
    );
  }

  public addLiabilitiesPaid(vaultAddress: Address, amountGwei: number): void {
    this._incrementVaultAmountCounter(
      LineaNativeYieldAutomationServiceMetrics.LiabilitiesPaidTotal,
      vaultAddress,
      amountGwei,
    );
  }

  public addLidoFeesPaid(vaultAddress: Address, amountGwei: number): void {
    this._incrementVaultAmountCounter(
      LineaNativeYieldAutomationServiceMetrics.LidoFeesPaidTotal,
      vaultAddress,
      amountGwei,
    );
  }

  public incrementOperationModeTrigger(mode: OperationMode, trigger: OperationTrigger): void {
    this.metricsService.incrementCounter(LineaNativeYieldAutomationServiceMetrics.OperationModeTriggerTotal, {
      mode,
      trigger,
    });
  }

  public incrementOperationModeExecution(mode: OperationMode): void {
    this.metricsService.incrementCounter(LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionTotal, {
      mode,
    });
  }

  public recordOperationModeDuration(mode: OperationMode, durationSeconds: number): void {
    if (durationSeconds < 0) return;
    this.metricsService.addValueToHistogram(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionDurationSeconds,
      durationSeconds,
      { mode },
    );
  }

  private _incrementVaultAmountCounter(
    metric: LineaNativeYieldAutomationServiceMetrics,
    vaultAddress: Address,
    amountGwei: number,
  ): void {
    if (amountGwei <= 0) return;
    this.metricsService.incrementCounter(metric, { vault_address: vaultAddress }, amountGwei);
  }
}
