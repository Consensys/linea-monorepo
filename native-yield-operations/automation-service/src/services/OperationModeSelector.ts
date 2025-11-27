import { ILogger, wait, attempt, weiToGweiNumber } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../core/clients/contracts/IYieldManager.js";
import { Address, TransactionReceipt } from "viem";
import { IOperationModeSelector } from "../core/services/operation-mode/IOperationModeSelector.js";
import { IOperationModeProcessor } from "../core/services/operation-mode/IOperationModeProcessor.js";
import { INativeYieldAutomationMetricsUpdater } from "../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { IValidatorDataClient } from "../core/clients/IValidatorDataClient.js";
import { OperationMode } from "../core/enums/OperationModeEnums.js";
import { OperationModeExecutionStatus } from "../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";

/**
 * Selects and executes the appropriate operation mode based on the yield provider's ossification state.
 * Continuously polls the YieldManager contract to determine the current state and routes execution
 * to the corresponding operation mode processor. Handles errors with retry logic.
 */
export class OperationModeSelector implements IOperationModeSelector {
  private isRunning = false;

  /**
   * Creates a new OperationModeSelector instance.
   *
   * @param {ILogger} logger - Logger instance for logging operation mode selection and execution.
   * @param {INativeYieldAutomationMetricsUpdater} metricsUpdater - Service for updating operation mode metrics.
   * @param {IYieldManager<TransactionReceipt>} yieldManagerContractClient - Client for reading yield provider state from YieldManager contract.
   * @param {IValidatorDataClient} validatorDataClient - Client for retrieving validator data.
   * @param {IOperationModeProcessor} yieldReportingOperationModeProcessor - Processor for YIELD_REPORTING_MODE operations.
   * @param {IOperationModeProcessor} ossificationPendingOperationModeProcessor - Processor for OSSIFICATION_PENDING_MODE operations.
   * @param {IOperationModeProcessor} ossificationCompleteOperationModeProcessor - Processor for OSSIFICATION_COMPLETE_MODE operations.
   * @param {Address} yieldProvider - The yield provider address to monitor and process.
   * @param {number} contractReadRetryTimeMs - Delay in milliseconds before retrying after a contract read error.
   */
  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly validatorDataClient: IValidatorDataClient,
    private readonly yieldReportingOperationModeProcessor: IOperationModeProcessor,
    private readonly ossificationPendingOperationModeProcessor: IOperationModeProcessor,
    private readonly ossificationCompleteOperationModeProcessor: IOperationModeProcessor,
    private readonly yieldProvider: Address,
    private readonly contractReadRetryTimeMs: number,
  ) {}

  /**
   * Starts the operation mode selection loop.
   * Sets the running flag and begins polling for operation mode selection.
   * If already running, returns immediately without starting a new loop.
   *
   * @returns {Promise<void>} A promise that resolves when the loop starts (but does not resolve until the loop stops).
   */
  public async start(): Promise<void> {
    if (this.isRunning) {
      this.logger.debug("OperationModeSelector.start() - already running, skipping");
      return;
    }

    this.isRunning = true;
    this.logger.info(`Starting selectOperationModeLoop`);
    await this.selectOperationModeLoop();
  }

  /**
   * Stops the operation mode selection loop.
   * Sets the running flag to false, which causes the loop to exit on its next iteration.
   * If not running, returns immediately.
   */
  public stop(): void {
    if (!this.isRunning) {
      this.logger.debug("OperationModeSelector.stop() - not running, skipping");
      return;
    }

    this.isRunning = false;
    this.logger.info(`Stopped selectOperationModeLoop`);
  }

  /**
   * Refreshes gauge metrics by querying validator data and updating the total pending partial withdrawals gauge.
   * Follows the same pattern as BeaconChainStakingClient.submitWithdrawalRequestsToFulfilAmount.
   *
   * @returns {Promise<void>} A promise that resolves when the gauge is updated (or silently fails if validator data is unavailable).
   */
  private async refreshGaugeMetrics(): Promise<void> {
    const sortedValidatorList = await this.validatorDataClient.getActiveValidatorsWithPendingWithdrawals();
    if (sortedValidatorList === undefined) {
      return;
    }
    const totalPendingPartialWithdrawalsWei =
      this.validatorDataClient.getTotalPendingPartialWithdrawalsWei(sortedValidatorList);
    this.metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei(weiToGweiNumber(totalPendingPartialWithdrawalsWei));
  }

  /**
   * Main loop that continuously selects and executes operation modes based on yield provider state.
   * Polls the YieldManager contract to check ossification status and routes execution accordingly:
   * - If ossified: executes OSSIFICATION_COMPLETE_MODE processor
   * - Else if ossification initiated: executes OSSIFICATION_PENDING_MODE processor
   * - Otherwise: executes YIELD_REPORTING_MODE processor
   * Records metrics for each execution and handles errors with retry logic using contractReadRetryTimeMs delay.
   *
   * @returns {Promise<void>} A promise that resolves when the loop exits (when isRunning becomes false).
   */
  private async selectOperationModeLoop(): Promise<void> {
    while (this.isRunning) {
      const refreshMetricsResult = await attempt(
        this.logger,
        () => this.refreshGaugeMetrics(),
        "Failed to refresh gauge metrics",
      );
      if (refreshMetricsResult.isErr()) {
        this.logger.error("Failed to refresh gauge metrics with details", {
          error: refreshMetricsResult.error,
          errorMessage: refreshMetricsResult.error.message,
          errorStack: refreshMetricsResult.error.stack,
        });
      }
      let currentMode: OperationMode = OperationMode.UNKNOWN;
      try {
        const [isOssificationInitiated, isOssified] = await Promise.all([
          this.yieldManagerContractClient.isOssificationInitiated(this.yieldProvider),
          this.yieldManagerContractClient.isOssified(this.yieldProvider),
        ]);

        if (isOssified) {
          currentMode = OperationMode.OSSIFICATION_COMPLETE_MODE;
          this.logger.info("Selected OSSIFICATION_COMPLETE_MODE");
          await this.ossificationCompleteOperationModeProcessor.process();
          this.logger.info("Completed OSSIFICATION_COMPLETE_MODE");
          this.metricsUpdater.incrementOperationModeExecution(
            OperationMode.OSSIFICATION_COMPLETE_MODE,
            OperationModeExecutionStatus.Success,
          );
        } else if (isOssificationInitiated) {
          currentMode = OperationMode.OSSIFICATION_PENDING_MODE;
          this.logger.info("Selected OSSIFICATION_PENDING_MODE");
          await this.ossificationPendingOperationModeProcessor.process();
          this.logger.info("Completed OSSIFICATION_PENDING_MODE");
          this.metricsUpdater.incrementOperationModeExecution(
            OperationMode.OSSIFICATION_PENDING_MODE,
            OperationModeExecutionStatus.Success,
          );
        } else {
          currentMode = OperationMode.YIELD_REPORTING_MODE;
          this.logger.info("Selected YIELD_REPORTING_MODE");
          await this.yieldReportingOperationModeProcessor.process();
          this.logger.info("Completed YIELD_REPORTING_MODE");
          this.metricsUpdater.incrementOperationModeExecution(
            OperationMode.YIELD_REPORTING_MODE,
            OperationModeExecutionStatus.Success,
          );
        }
      } catch (error) {
        this.logger.error(`selectOperationModeLoop error, retrying in ${this.contractReadRetryTimeMs}ms`, { error });
        this.metricsUpdater.incrementOperationModeExecution(currentMode, OperationModeExecutionStatus.Failure);
        await wait(this.contractReadRetryTimeMs);
      }
    }
  }
}
