import { Result } from "neverthrow";
import { TransactionReceipt } from "viem";
import { weiToGweiNumber } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { RebalanceDirection } from "../../core/entities/RebalanceRequirement.js";

export const recordUnstakeRebalanceFromSafeWithdrawalResult = (
  result: Result<TransactionReceipt | undefined, Error>,
  yieldManagerClient: IYieldManager<TransactionReceipt>,
  metricsUpdater: INativeYieldAutomationMetricsUpdater,
) => {
  if (result.isErr()) return;

  const receipt = result.value;
  if (!receipt) return;

  const amountWei = yieldManagerClient.getWithdrawalAmountFromTxReceipt(receipt);
  if (amountWei <= 0n) return;

  metricsUpdater.recordRebalance(RebalanceDirection.UNSTAKE, weiToGweiNumber(amountWei));
};
