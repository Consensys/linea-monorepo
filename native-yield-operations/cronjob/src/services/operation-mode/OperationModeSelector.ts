import { ILogger } from "ts-libs/linea-shared-utils/src";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { NativeYieldCronJobClientConfig } from "../../application/main/config/NativeYieldCronJobClientConfig";
import { wait } from "@consensys/linea-sdk";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { TransactionReceipt } from "viem";

export class OperationModeSelector implements IOperationModeSelector {
  constructor(
    private readonly config: NativeYieldCronJobClientConfig,
    private readonly logger: ILogger,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
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
      const [isOssificationInitiated, isOssified] = await Promise.all([
        this.yieldManagerContractClient.isOssificationInitiated(this.config.contractAddresses.lidoYieldProviderAddress),
        this.yieldManagerContractClient.isOssified(this.config.contractAddresses.lidoYieldProviderAddress),
      ]);
      if (isOssified) {
        // Run ossification complete
      } else if (isOssificationInitiated) {
        // Run ossification pending
      } else {
        // Run Yield reporter
      }
    } catch (e) {
      this.logger.error(e);
      await wait(this.config.timing.contractReadRetryTimeSeconds);
    }
    // Main infinite loop
    this.selectOperationMode();
  }
}
