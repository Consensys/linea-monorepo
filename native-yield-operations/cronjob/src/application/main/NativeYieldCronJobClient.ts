import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { NativeYieldCronJobClientConfig } from "./config/NativeYieldCronJobClientConfig";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { OperationModeSelector } from "../../services/operation-mode/OperationModeSelector";
import {
  createPublicClient,
  http,
} from "viem";
import { mainnet } from "viem/chains";
export class NativeYieldCronJobClient {
  private readonly config: NativeYieldCronJobClientConfig;
  private readonly logger: ILogger;

  private operationModeSelector: IOperationModeSelector

  constructor(config: NativeYieldCronJobClientConfig) {
    this.config = config;
    this.logger = new WinstonLogger(NativeYieldCronJobClient.name, config.loggerOptions);

    this.operationModeSelector = new OperationModeSelector(config, new WinstonLogger(OperationModeSelector.name, config.loggerOptions))
  }

  public async connectServices(): Promise<void> {
    // TO-DO - startup Prom metrics API endpoint
  }

  public startAllServices(): void {
    this.operationModeSelector.start();
    this.logger.info("Native yield cron job started");
  }

  public stopAllServices(): void {
    this.operationModeSelector.stop();
    this.logger.info("Native yield cron job stopped");
  }

  public getConfig(): NativeYieldCronJobClientConfig {
    return this.config;
  }
}
