import { Result } from "neverthrow";
import { Address, TransactionReceipt } from "viem";
import { weiToGweiNumber } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { IVaultHub } from "../../core/clients/contracts/IVaultHub.js";

export const updateMetricsForTransferFundsResult = async (
  result: Result<TransactionReceipt | undefined, Error>,
  metricsUpdater: INativeYieldAutomationMetricsUpdater,
  yieldManagerClient: IYieldManager<TransactionReceipt>,
  vaultHubClient: IVaultHub<TransactionReceipt>,
  yieldProvider: Address,
) => {
  if (result.isErr()) return;

  const receipt = result.value;
  if (!receipt) return;

  const vault = await yieldManagerClient.getLidoStakingVaultAddress(yieldProvider);
  const liabilityPayment = vaultHubClient.getLiabilityPaymentFromTxReceipt(receipt);
  if (liabilityPayment != 0n) {
    metricsUpdater.addLiabilitiesPaid(vault, weiToGweiNumber(liabilityPayment));
  }
};
