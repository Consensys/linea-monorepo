import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { NativeYieldCronJobClientConfig } from "./config/NativeYieldCronJobClientConfig";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { OperationModeSelector } from "../../services/operation-mode/OperationModeSelector";
import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import { EthereumMainnetClientLibrary } from "ts-libs/linea-shared-utils/src/clients/ethereum/EthereumMainnetClientLibrary";
import { BaseError, PublicClient, TransactionReceipt } from "viem";
import { YieldManagerContractClient } from "../../clients/YieldManagerContractClient";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
export class NativeYieldCronJobClient {
  private readonly config: NativeYieldCronJobClientConfig;
  private readonly logger: ILogger;

  private ethereumMainnetClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt, BaseError>;
  private yieldManagerContractClient: IYieldManager<TransactionReceipt>;

  private operationModeSelector: IOperationModeSelector;

  constructor(config: NativeYieldCronJobClientConfig) {
    this.config = config;
    this.logger = new WinstonLogger(NativeYieldCronJobClient.name, config.loggerOptions);

    this.ethereumMainnetClientLibrary = new EthereumMainnetClientLibrary(config.dataSources.l1RpcUrl);
    this.yieldManagerContractClient = new YieldManagerContractClient(
      this.ethereumMainnetClientLibrary,
      config.contractAddresses.yieldManagerAddress,
    );

    this.operationModeSelector = new OperationModeSelector(
      config,
      new WinstonLogger(OperationModeSelector.name, config.loggerOptions),
      this.yieldManagerContractClient,
    );
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
