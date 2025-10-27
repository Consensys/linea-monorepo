import { Result } from "neverthrow";
import { Address, TransactionReceipt } from "viem";
import { weiToGweiNumber } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { getNodeOperatorFeesPaidFromTxReceipt } from "../../clients/contracts/getNodeOperatorFeesPaidFromTxReceipt.js";
import { IVaultHub } from "../../core/clients/contracts/IVaultHub.js";

export const updateMetricsForProgressOssification = async (
  result: Result<TransactionReceipt | undefined, Error>,
  metricsUpdater: INativeYieldAutomationMetricsUpdater,
  yieldManagerClient: IYieldManager<TransactionReceipt>,
  vaultHubClient: IVaultHub<TransactionReceipt>,
  yieldProvider: Address,
) => {
  if (result.isErr()) return;

  const receipt = result.value;
  if (!receipt) return;

  const [vault, dashboard] = await Promise.all([
    yieldManagerClient.getLidoStakingVaultAddress(yieldProvider),
    yieldManagerClient.getLidoDashboardAddress(yieldProvider),
  ]);

  const nodeOperatorFeesDisbursed = getNodeOperatorFeesPaidFromTxReceipt(receipt, dashboard);
  if (nodeOperatorFeesDisbursed != 0n) {
    metricsUpdater.addNodeOperatorFeesPaid(vault, weiToGweiNumber(nodeOperatorFeesDisbursed));
  }

  const lidoFeePayment = vaultHubClient.getLidoFeePaymentFromTxReceipt(receipt);
  if (lidoFeePayment != 0n) {
    metricsUpdater.addLidoFeesPaid(vault, weiToGweiNumber(lidoFeePayment));
  }

  const liabilityPayment = vaultHubClient.getLiabilityPaymentFromTxReceipt(receipt);
  if (liabilityPayment != 0n) {
    metricsUpdater.addLiabilitiesPaid(vault, weiToGweiNumber(liabilityPayment));
  }
};
