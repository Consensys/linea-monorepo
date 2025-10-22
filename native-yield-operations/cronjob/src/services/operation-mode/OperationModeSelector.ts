import { ILogger } from "ts-libs/linea-shared-utils";
import { wait } from "@consensys/linea-sdk";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { TransactionReceipt } from "viem";
import { NativeYieldCronJobClientConfig } from "../../application/main/config/NativeYieldCronJobClientConfig";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";

export class OperationModeSelector implements IOperationModeSelector {
  private readonly yieldReportingOperationModeProcessor: IOperationModeProcessor;
  private isRunning = false;

  constructor(
    private readonly config: NativeYieldCronJobClientConfig,
    private readonly logger: ILogger,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    yieldReportingOperationModeProcessor: IOperationModeProcessor,
  ) {
    this.yieldReportingOperationModeProcessor = yieldReportingOperationModeProcessor;
  }

  public async start(): Promise<void> {
    if (this.isRunning) {
      return;
    }

    this.isRunning = true;
    this.logger.info("Starting %s...", this.logger.name);
    void this.selectOperationModeLoop();
  }

  public stop(): void {
    if (!this.isRunning) {
      return;
    }

    this.isRunning = false;
    this.logger.info("Stopped %s...", this.logger.name);
  }

  private async selectOperationModeLoop(): Promise<void> {
    while (this.isRunning) {
      try {
        const [isOssificationInitiated, isOssified] = await Promise.all([
          this.yieldManagerContractClient.isOssificationInitiated(
            this.config.contractAddresses.lidoYieldProviderAddress,
          ),
          this.yieldManagerContractClient.isOssified(this.config.contractAddresses.lidoYieldProviderAddress),
        ]);

        if (isOssified) {
          this.logger.info("Selected OSSIFICATION_COMPLETE_MODE");
        } else if (isOssificationInitiated) {
          this.logger.info("Selected OSSIFICATION_PENDING_MODE");
        } else {
          this.logger.info("Selected YIELD_REPORTING_MODE");
          await this.yieldReportingOperationModeProcessor.process();
        }
      } catch (error) {
        this.logger.error(error as Error);
        await wait(this.config.timing.contractReadRetryTimeSeconds);
      }
    }
  }
}
