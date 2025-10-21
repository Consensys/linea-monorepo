import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { NativeYieldCronJobClientConfig } from "./config/NativeYieldCronJobClientConfig";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { OperationModeSelector } from "../../services/operation-mode/OperationModeSelector";
import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import { EthereumMainnetClientLibrary } from "ts-libs/linea-shared-utils/src/clients/ethereum/EthereumMainnetClientLibrary";
import { PublicClient, TransactionReceipt } from "viem";
import { YieldManagerContractClient } from "../../clients/YieldManagerContractClient";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { IContractSignerService } from "ts-libs/linea-shared-utils/src/core/services/IContractSignerService";
import { Web3SignerService } from "ts-libs/linea-shared-utils/src/services/signers/Web3SignerService";
import { YieldReportingOperationModeProcessor } from "../../services/operation-mode/YieldReportingOperationModeProcessor";
import { IBaseContractClient } from "../../core/clients/IBaseContractClient";

export class NativeYieldCronJobClient {
  private readonly config: NativeYieldCronJobClientConfig;
  private readonly logger: ILogger;

  private ethereumMainnetClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>;
  private yieldManagerContractClient: IYieldManager<TransactionReceipt> & IBaseContractClient;
  private web3SignerService: IContractSignerService;

  private operationModeSelector: IOperationModeSelector;

  constructor(config: NativeYieldCronJobClientConfig) {
    this.config = config;
    this.logger = new WinstonLogger(NativeYieldCronJobClient.name, config.loggerOptions);

    this.web3SignerService = new Web3SignerService(
      config.web3signer.url,
      config.web3signer.publicKey,
      config.web3signer.keystore.path,
      config.web3signer.keystore.passphrase,
      config.web3signer.truststore.path,
      config.web3signer.truststore.passphrase,
    );
    this.ethereumMainnetClientLibrary = new EthereumMainnetClientLibrary(
      config.dataSources.l1RpcUrl,
      this.web3SignerService,
    );
    this.yieldManagerContractClient = new YieldManagerContractClient(
      this.ethereumMainnetClientLibrary,
      config.contractAddresses.yieldManagerAddress,
    );

    const yieldReportingProcessor = new YieldReportingOperationModeProcessor(
      this.yieldManagerContractClient,
      this.ethereumMainnetClientLibrary,
      new WinstonLogger(YieldReportingOperationModeProcessor.name, config.loggerOptions),
      config.timing.trigger,
    );

    this.operationModeSelector = new OperationModeSelector(
      config,
      new WinstonLogger(OperationModeSelector.name, config.loggerOptions),
      this.yieldManagerContractClient,
      yieldReportingProcessor,
    );
  }

  public async connectServices(): Promise<void> {
    // TO-DO - startup Prom metrics API endpoint
  }

  public startAllServices(): void {
    void this.operationModeSelector.start();
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
