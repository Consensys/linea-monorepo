import { ILogger } from "@consensys/linea-shared-utils";
import { wait } from "@consensys/linea-sdk";
import { IYieldManager } from "../core/clients/contracts/IYieldManager.js";
import { Address, TransactionReceipt } from "viem";
import { IOperationModeSelector } from "../core/services/operation-mode/IOperationModeSelector.js";
import { IOperationModeProcessor } from "../core/services/operation-mode/IOperationModeProcessor.js";
import { INativeYieldAutomationMetricsUpdater } from "../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { OperationMode } from "../core/enums/OperationModeEnums.js";

export class OperationModeSelector implements IOperationModeSelector {
  private isRunning = false;

  constructor(
    private readonly logger: ILogger,
    private readonly metricsUpdater: INativeYieldAutomationMetricsUpdater,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly yieldReportingOperationModeProcessor: IOperationModeProcessor,
    private readonly ossificationPendingOperationModeProcessor: IOperationModeProcessor,
    private readonly ossificationCompleteOperationModeProcessor: IOperationModeProcessor,
    private readonly yieldProvider: Address,
    private readonly contractReadRetryTimeMs: number,
  ) {
    this.yieldReportingOperationModeProcessor = yieldReportingOperationModeProcessor;
  }

  public async start(): Promise<void> {
    if (this.isRunning) {
      return;
    }

    this.isRunning = true;
    this.logger.info(`Starting ${this.logger.name}...`, { loggerName: this.logger.name });
    void this.selectOperationModeLoop();
  }

  public stop(): void {
    if (!this.isRunning) {
      return;
    }

    this.isRunning = false;
    this.logger.info(`Stopped ${this.logger.name}...`, { loggerName: this.logger.name });
  }

  private async selectOperationModeLoop(): Promise<void> {
    while (this.isRunning) {
      try {
        const [isOssificationInitiated, isOssified] = await Promise.all([
          this.yieldManagerContractClient.isOssificationInitiated(this.yieldProvider),
          this.yieldManagerContractClient.isOssified(this.yieldProvider),
        ]);

        if (isOssified) {
          this.logger.info("Selected OSSIFICATION_COMPLETE_MODE");
          await this.ossificationCompleteOperationModeProcessor.process();
          this.logger.info("Completed OSSIFICATION_COMPLETE_MODE");
          this.metricsUpdater.incrementOperationModeExecution(OperationMode.OSSIFICATION_COMPLETE_MODE);
        } else if (isOssificationInitiated) {
          this.logger.info("Selected OSSIFICATION_PENDING_MODE");
          await this.ossificationPendingOperationModeProcessor.process();
          this.logger.info("Completed OSSIFICATION_PENDING_MODE");
          this.metricsUpdater.incrementOperationModeExecution(OperationMode.OSSIFICATION_PENDING_MODE);
        } else {
          this.logger.info("Selected YIELD_REPORTING_MODE");
          await this.yieldReportingOperationModeProcessor.process();
          this.logger.info("Completed YIELD_REPORTING_MODE");
          this.metricsUpdater.incrementOperationModeExecution(OperationMode.YIELD_REPORTING_MODE);
        }
      } catch (error) {
        this.logger.error(
          `selectOperationModeLoop error, retrying in ${this.contractReadRetryTimeMs}ms`,
          { error },
        );
        await wait(this.contractReadRetryTimeMs);
      }
    }
  }
}
