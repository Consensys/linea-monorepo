import { ILogger } from "ts-libs/linea-shared-utils/src";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { NativeYieldCronJobClientConfig } from "../../application/main/config/NativeYieldCronJobClientConfig";
import { wait } from "@consensys/linea-sdk";

export class OperationModeSelector implements IOperationModeSelector {
  // Need IYieldManager instance...
  constructor(
    private readonly config: NativeYieldCronJobClientConfig,
    private readonly logger: ILogger,
  ) {}

  public async start(): Promise<void> {
    this.logger.info("Starting %s %s...", this.logger.name);
    this.selectOperationMode();
  }

  public stop() {
    this.logger.info("Stopped %s %s...", this.logger.name);
  }

  private async selectOperationMode(): Promise<void> {
    try {
      // Get ossification state from YieldManager
      // Choose mode based on ossification state
      // const { fromBlock, fromBlockLogIndex } = await this.getInitialFromBlock();
      // this.processEvents(fromBlock, fromBlockLogIndex);
    } catch (e) {
      this.logger.error(e);
      await wait(this.config.timing.contractReadRetryTimeSeconds);
    }
    // Main infinite loop
    this.selectOperationMode();
  }
}
