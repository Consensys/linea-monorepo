import { Address, TransactionReceipt } from "viem";
import { ILogger, attempt, msToSeconds } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { wait } from "@consensys/linea-sdk";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { OperationMode } from "../../core/enums/OperationModeEnums.js";
import { recordUnstakeRebalanceFromSafeWithdrawalResult } from "../../application/metrics/recordUnstakeRebalanceFromSafeWithdrawalResult.js";
import { OperationTrigger } from "../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";

export class OssificationCompleteProcessor implements IOperationModeProcessor {
  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly maxInactionMs: number,
    private readonly yieldProvider: Address,
  ) {}

  public async process(): Promise<void> {
    this.logger.info(`Waiting ${this.maxInactionMs}ms before executing actions`);
    await wait(this.maxInactionMs);

    this.metricsUpdater.incrementOperationModeTrigger(
      OperationMode.OSSIFICATION_COMPLETE_MODE,
      OperationTrigger.TIMEOUT,
    );
    const startedAt = performance.now();
    await this._process();
    const durationMs = performance.now() - startedAt;
    this.metricsUpdater.recordOperationModeDuration(OperationMode.OSSIFICATION_COMPLETE_MODE, msToSeconds(durationMs));
  }

  private async _process(): Promise<void> {
    // Max withdraw
    this.logger.info("_process - Performing max withdrawal from YieldProvider");
    const withdrawalResult = await attempt(
      this.logger,
      () => this.yieldManagerContractClient.safeMaxAddToWithdrawalReserve(this.yieldProvider),
      "_process - safeMaxAddToWithdrawalReserve failed (tolerated)",
    );
    recordUnstakeRebalanceFromSafeWithdrawalResult(
      withdrawalResult,
      this.yieldManagerContractClient,
      this.metricsUpdater,
    );

    // Max unstake
    this.logger.info("_process - Performing max unstake from beacon chain");
    await this.beaconChainStakingClient.submitMaxAvailableWithdrawalRequests();
  }
}
