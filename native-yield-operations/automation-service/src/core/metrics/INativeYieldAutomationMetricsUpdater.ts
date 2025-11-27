import { Address, Hex } from "viem";
import { RebalanceDirection } from "../entities/RebalanceRequirement.js";
import { OperationMode } from "../enums/OperationModeEnums.js";
import { OperationModeExecutionStatus } from "./LineaNativeYieldAutomationServiceMetrics.js";

export interface INativeYieldAutomationMetricsUpdater {
  recordRebalance(direction: RebalanceDirection, amountGwei: number): void;

  addValidatorPartialUnstakeAmount(validatorPubkey: Hex, amountGwei: number): void;

  incrementValidatorExit(validatorPubkey: Hex, count?: number): void;

  incrementLidoVaultAccountingReport(vaultAddress: Address): void;

  incrementReportYield(vaultAddress: Address): void;

  addReportedYieldAmount(vaultAddress: Address, amountGwei: number): void;

  setLastPeekedNegativeYieldReport(vaultAddress: Address, negativeYield: number): void;

  setLastPeekedPositiveYieldReport(vaultAddress: Address, yieldAmount: number): void;

  setLastSettleableLidoFees(vaultAddress: Address, feesAmount: number): void;

  setLastTotalPendingPartialWithdrawalsGwei(totalPendingPartialWithdrawalsGwei: number): void;

  addNodeOperatorFeesPaid(vaultAddress: Address, amountGwei: number): void;

  addLiabilitiesPaid(vaultAddress: Address, amountGwei: number): void;

  addLidoFeesPaid(vaultAddress: Address, amountGwei: number): void;

  incrementOperationModeExecution(mode: OperationMode, status?: OperationModeExecutionStatus): void;

  recordOperationModeDuration(mode: OperationMode, durationSeconds: number): void;
}
