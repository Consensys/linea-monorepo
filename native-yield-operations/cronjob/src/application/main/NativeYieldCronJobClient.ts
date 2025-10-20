import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { NativeYieldCronJobClientOptions } from "./config/NativeYieldCronJobClientOptions";

export class NativeYieldCronJobClient {
  private readonly config: NativeYieldCronJobClientOptions;
  private readonly logger: ILogger;

  constructor(private readonly options: NativeYieldCronJobClientOptions) {
    this.config = options;
    this.logger = new WinstonLogger(NativeYieldCronJobClient.name, options.loggerOptions);
  }

  public async connectServices(): Promise<void> {
    // TO-DO - startup Prom metrics API endpoint
  }

  /**
   * Starts all cron job processors.
   */
  public startAllServices(): void {
    this.logger.info("Native yield cron job started");
  }

  /**
   * Stops the cron job processors.
   */
  public stopAllServices(): void {
    this.logger.info("Native yield cron job stopped");
  }

  public getConfig(): NativeYieldCronJobClientOptions {
    return this.config;
  }
}
