import { Result } from "neverthrow";
import { TransactionReceipt } from "viem";
import { weiToGweiNumber } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { RebalanceDirection } from "../../core/entities/RebalanceRequirement.js";
import { IVaultHub } from "../../core/clients/contracts/IVaultHub.js";

export const updateMetricsForSafeWithdrawalResult = async (
  result: Result<TransactionReceipt | undefined, Error>,
  metricsUpdater: INativeYieldAutomationMetricsUpdater,
  yieldManagerClient: IYieldManager<TransactionReceipt>,
  vaultHubClient: IVaultHub<TransactionReceipt>,
) => {
  if (result.isErr()) return;

  const receipt = result.value;
  if (!receipt) return;

  const withdrawalReport = yieldManagerClient.getWithdrawalEventFromTxReceipt(receipt);
  if (!withdrawalReport) return;
  const { yieldProvider, reserveIncrementAmount } = withdrawalReport;

  metricsUpdater.recordRebalance(RebalanceDirection.UNSTAKE, weiToGweiNumber(reserveIncrementAmount));

  const vault = await yieldManagerClient.getLidoStakingVaultAddress(yieldProvider);
  const liabilityPayment = vaultHubClient.getLiabilityPaymentFromTxReceipt(receipt);
  if (liabilityPayment != 0n) {
    metricsUpdater.addLiabilitiesPaid(vault, weiToGweiNumber(liabilityPayment));
  }
};
