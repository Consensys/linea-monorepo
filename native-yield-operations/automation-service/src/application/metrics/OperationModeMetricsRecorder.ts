// Take operation results and record the relevant figures into metrics

import { Address, TransactionReceipt } from "viem";
import { Result } from "neverthrow";
import { IOperationModeMetricsRecorder } from "../../core/metrics/IOperationModeMetricsRecorder.js";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { ILogger, weiToGweiNumber } from "@consensys/linea-shared-utils";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { IVaultHub } from "../../core/clients/contracts/IVaultHub.js";
import { getNodeOperatorFeesPaidFromTxReceipt } from "../../clients/contracts/getNodeOperatorFeesPaidFromTxReceipt.js";
import { RebalanceDirection } from "../../core/entities/RebalanceRequirement.js";

export class OperationModeMetricsRecorder implements IOperationModeMetricsRecorder {
  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly yieldManagerClient: IYieldManager<TransactionReceipt>,
    private readonly vaultHubClient: IVaultHub<TransactionReceipt>,
  ) {
    void this.logger;
  }

  async recordProgressOssificationMetrics(
    yieldProvider: Address,
    txReceiptResult: Result<TransactionReceipt | undefined, Error>,
  ): Promise<void> {
    if (txReceiptResult.isErr()) return;
    const receipt = txReceiptResult.value;
    if (!receipt) return;

    const [vault, dashboard] = await Promise.all([
      this.yieldManagerClient.getLidoStakingVaultAddress(yieldProvider),
      this.yieldManagerClient.getLidoDashboardAddress(yieldProvider),
    ]);

    const nodeOperatorFeesDisbursed = getNodeOperatorFeesPaidFromTxReceipt(receipt, dashboard);
    if (nodeOperatorFeesDisbursed != 0n) {
      this.metricsUpdater.addNodeOperatorFeesPaid(vault, weiToGweiNumber(nodeOperatorFeesDisbursed));
    }

    const lidoFeePayment = this.vaultHubClient.getLidoFeePaymentFromTxReceipt(receipt);
    if (lidoFeePayment != 0n) {
      this.metricsUpdater.addLidoFeesPaid(vault, weiToGweiNumber(lidoFeePayment));
    }

    const liabilityPayment = this.vaultHubClient.getLiabilityPaymentFromTxReceipt(receipt);
    if (liabilityPayment != 0n) {
      this.metricsUpdater.addLiabilitiesPaid(vault, weiToGweiNumber(liabilityPayment));
    }
  }

  async recordReportYieldMetrics(
    yieldProvider: Address,
    txReceiptResult: Result<TransactionReceipt | undefined, Error>,
  ): Promise<void> {
    if (txReceiptResult.isErr()) return;
    const receipt = txReceiptResult.value;
    if (!receipt) return;

    const yieldReport = this.yieldManagerClient.getYieldReportFromTxReceipt(receipt);
    if (yieldReport === undefined) return;

    const [vault, dashboard] = await Promise.all([
      this.yieldManagerClient.getLidoStakingVaultAddress(yieldReport.yieldProvider),
      this.yieldManagerClient.getLidoDashboardAddress(yieldReport.yieldProvider),
    ]);

    this.metricsUpdater.incrementReportYield(vault);
    this.metricsUpdater.addReportedYieldAmount(vault, weiToGweiNumber(yieldReport.yieldAmount));
    this.metricsUpdater.setCurrentNegativeYieldLastReport(vault, weiToGweiNumber(yieldReport.outstandingNegativeYield));

    const nodeOperatorFeesDisbursed = getNodeOperatorFeesPaidFromTxReceipt(receipt, dashboard);
    if (nodeOperatorFeesDisbursed != 0n) {
      this.metricsUpdater.addNodeOperatorFeesPaid(vault, weiToGweiNumber(nodeOperatorFeesDisbursed));
    }

    const lidoFeePayment = this.vaultHubClient.getLidoFeePaymentFromTxReceipt(receipt);
    if (lidoFeePayment != 0n) {
      this.metricsUpdater.addLidoFeesPaid(vault, weiToGweiNumber(lidoFeePayment));
    }

    const liabilityPayment = this.vaultHubClient.getLiabilityPaymentFromTxReceipt(receipt);
    if (liabilityPayment != 0n) {
      this.metricsUpdater.addLiabilitiesPaid(vault, weiToGweiNumber(liabilityPayment));
    }
  }

  async recordSafeWithdrawalMetrics(
    yieldProvider: Address,
    txReceiptResult: Result<TransactionReceipt | undefined, Error>,
  ): Promise<void> {
    if (txReceiptResult.isErr()) return;
    const receipt = txReceiptResult.value;
    if (!receipt) return;

    const withdrawalReport = this.yieldManagerClient.getWithdrawalEventFromTxReceipt(receipt);
    if (!withdrawalReport) return;
    const { reserveIncrementAmount } = withdrawalReport;

    this.metricsUpdater.recordRebalance(RebalanceDirection.UNSTAKE, weiToGweiNumber(reserveIncrementAmount));

    const vault = await this.yieldManagerClient.getLidoStakingVaultAddress(yieldProvider);
    const liabilityPayment = this.vaultHubClient.getLiabilityPaymentFromTxReceipt(receipt);
    if (liabilityPayment != 0n) {
      this.metricsUpdater.addLiabilitiesPaid(vault, weiToGweiNumber(liabilityPayment));
    }
  }

  async recordTransferFundsMetrics(
    yieldProvider: Address,
    txReceiptResult: Result<TransactionReceipt | undefined, Error>,
  ): Promise<void> {
    if (txReceiptResult.isErr()) return;
    const receipt = txReceiptResult.value;
    if (!receipt) return;

    const vault = await this.yieldManagerClient.getLidoStakingVaultAddress(yieldProvider);
    const liabilityPayment = this.vaultHubClient.getLiabilityPaymentFromTxReceipt(receipt);
    if (liabilityPayment != 0n) {
      this.metricsUpdater.addLiabilitiesPaid(vault, weiToGweiNumber(liabilityPayment));
    }
  }
}
