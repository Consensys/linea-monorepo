import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { NativeYieldCronJobClientConfig } from "./config/NativeYieldCronJobClientConfig";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { OperationModeSelector } from "../../services/operation-mode/OperationModeSelector";
import { IContractClientLibrary } from "ts-libs/linea-shared-utils/core/client/IContractClientLibrary";
import { EthereumMainnetClientLibrary } from "ts-libs/linea-shared-utils/clients/ethereum/EthereumMainnetClientLibrary";
import { PublicClient, TransactionReceipt } from "viem";
import { YieldManagerContractClient } from "../../clients/YieldManagerContractClient";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { Web3SignerService } from "ts-libs/linea-shared-utils/services/signers/Web3SignerService";
import { YieldReportingOperationModeProcessor } from "../../services/operation-mode/YieldReportingOperationModeProcessor";
import { LazyOracleContractClient } from "../../clients/LazyOracleContractClient";
import { ILazyOracle } from "../../core/services/contracts/ILazyOracle";
import { ApolloClient, InMemoryCache, HttpLink, from } from "@apollo/client";
import { SetContextLink } from "@apollo/client/link/context";
import { IContractSignerClient, IOAuth2TokenClient } from "ts-libs/linea-shared-utils/src";

export class NativeYieldCronJobClient {
  private readonly config: NativeYieldCronJobClientConfig;
  private readonly logger: ILogger;

  private ethereumMainnetClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>;
  private yieldManagerContractClient: IYieldManager<TransactionReceipt>;
  private lazyOracleContractClient: ILazyOracle<TransactionReceipt>;

  private web3SignerService: IContractSignerClient;

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
    this.lazyOracleContractClient = new LazyOracleContractClient(
      this.ethereumMainnetClientLibrary,
      config.contractAddresses.lazyOracleAddress,
      new WinstonLogger(LazyOracleContractClient.name, config.loggerOptions),
      config.timing.trigger.maxInactionMs,
    );

    const yieldReportingProcessor = new YieldReportingOperationModeProcessor(
      this.yieldManagerContractClient,
      this.lazyOracleContractClient,
      new WinstonLogger(YieldReportingOperationModeProcessor.name, config.loggerOptions),
      config.timing.trigger.pollIntervalMs,
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

  private _createApolloClient(oAuth2TokenClient: IOAuth2TokenClient): ApolloClient {
    // --- create the base HTTP transport
    const httpLink = new HttpLink({
      uri: this.config.dataSources.stakingGraphQLUrl,
    });

    const asyncAuthLink = new SetContextLink(async (prevContext, operation) => {
      const token = await oAuth2TokenClient.getBearerToken();
      return {
        headers: {
          ...prevContext.headers,
          authorization: `Bearer ${token}`,
        },
      };
    });
    // --- combine links so authLink runs before httpLink
    const client = new ApolloClient({
      link: from([asyncAuthLink, httpLink]),
      cache: new InMemoryCache(),
    });
    return client;
  }
}
