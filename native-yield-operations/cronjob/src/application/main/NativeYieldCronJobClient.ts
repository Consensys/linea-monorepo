import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { NativeYieldCronJobClientConfig } from "./config/NativeYieldCronJobClientConfig";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector";
import { OperationModeSelector } from "../../services/operation-mode/OperationModeSelector";
import {
  IContractClientLibrary,
  ContractClientLibrary,
  Web3SignerClient,
  IContractSignerClient,
  IOAuth2TokenClient,
  IBeaconNodeAPIClient,
  BeaconNodeApiClient,
  OAuth2TokenClient,
} from "ts-libs/linea-shared-utils/src";
import {} from "ts-libs/linea-shared-utils/src";
import { Chain, PublicClient, TransactionReceipt } from "viem";
import { YieldManagerContractClient } from "../../clients/YieldManagerContractClient";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { YieldReportingOperationModeProcessor } from "../../services/operation-mode/YieldReportingOperationModeProcessor";
import { LazyOracleContractClient } from "../../clients/LazyOracleContractClient";
import { ILazyOracle } from "../../core/services/contracts/ILazyOracle";
import { ApolloClient } from "@apollo/client";
import { ILineaRollupYieldExtension } from "../../core/services/contracts/ILineaRollupYieldExtension";
import { LineaRollupYieldExtensionContractClient } from "../../clients/LineaRollupYieldExtensionContractClient";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient";
import { IValidatorDataClient } from "../../core/clients/IValidatorDataClient";
import { ConsensysStakingGraphQLClient } from "../../clients/ConsensysStakingGraphQLClient";
import { LidoAccountingReportClient } from "../../clients/LidoAccountingReportClient";
import { BeaconChainStakingClient } from "../../clients/BeaconChainStakingClient";
import { OssificationCompleteOperationModeProcessor } from "../../services/operation-mode/OssificationCompleteOperationModeProcessor";
import { OssificationPendingOperationModeProcessor } from "../../services/operation-mode/OssificationPendingOperationModeProcessor";
import { mainnet, hoodi } from "viem/chains";
import { createApolloClient } from "../../utils/createApolloClient";

export class NativeYieldCronJobClient {
  private readonly config: NativeYieldCronJobClientConfig;
  private readonly logger: ILogger;

  private ContractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>;
  private web3SignerClient: IContractSignerClient;
  private yieldManagerContractClient: IYieldManager<TransactionReceipt>;
  private lazyOracleContractClient: ILazyOracle<TransactionReceipt>;
  private lineaRollupYieldExtensionContractClient: ILineaRollupYieldExtension<TransactionReceipt>;

  private beaconNodeApiClient: IBeaconNodeAPIClient;
  private oAuth2TokenClient: IOAuth2TokenClient;
  private apolloClient: ApolloClient;
  private beaconChainStakingClient: IBeaconChainStakingClient;
  private lidoAccountingReportClient: ILidoAccountingReportClient;
  private consensysStakingGraphQLClient: IValidatorDataClient;

  private operationModeSelector: IOperationModeSelector;
  private yieldReportingOperationModeProcessor: IOperationModeProcessor;
  private ossificationPendingOperationModeProcessor: IOperationModeProcessor;
  private ossificationCompleteOperationModeProcessor: IOperationModeProcessor;

  constructor(config: NativeYieldCronJobClientConfig) {
    this.config = config;
    this.logger = new WinstonLogger(NativeYieldCronJobClient.name, config.loggerOptions);

    this.web3SignerClient = new Web3SignerClient(
      config.web3signer.url,
      config.web3signer.publicKey,
      config.web3signer.keystore.path,
      config.web3signer.keystore.passphrase,
      config.web3signer.truststore.path,
      config.web3signer.truststore.passphrase,
    );

    const getChain = (chainId: number): Chain => {
      switch (chainId) {
        case mainnet.id:
          return mainnet;
        case hoodi.id:
          return hoodi;
        default:
          throw new Error(`Unsupported chain ID: ${chainId}`);
      }
    };
    this.ContractClientLibrary = new ContractClientLibrary(
      config.dataSources.l1RpcUrl,
      getChain(config.dataSources.chainId),
      this.web3SignerClient,
    );
    this.yieldManagerContractClient = new YieldManagerContractClient(
      this.ContractClientLibrary,
      config.contractAddresses.yieldManagerAddress,
      config.rebalanceToleranceBps,
      config.minWithdrawalThresholdEth,
    );
    this.lazyOracleContractClient = new LazyOracleContractClient(
      this.ContractClientLibrary,
      config.contractAddresses.lazyOracleAddress,
      new WinstonLogger(LazyOracleContractClient.name, config.loggerOptions),
      config.timing.trigger.maxInactionMs,
    );
    this.lineaRollupYieldExtensionContractClient = new LineaRollupYieldExtensionContractClient(
      this.ContractClientLibrary,
      config.contractAddresses.lineaRollupContractAddress,
    );

    this.beaconNodeApiClient = new BeaconNodeApiClient(config.dataSources.beaconChainRpcUrl);
    this.oAuth2TokenClient = new OAuth2TokenClient(
      new WinstonLogger(OAuth2TokenClient.name, config.loggerOptions),
      config.consensysStakingOAuth2.tokenEndpoint,
      config.consensysStakingOAuth2.clientId,
      config.consensysStakingOAuth2.clientSecret,
      config.consensysStakingOAuth2.audience,
    );
    this.apolloClient = createApolloClient(this.oAuth2TokenClient, config.dataSources.stakingGraphQLUrl);
    this.consensysStakingGraphQLClient = new ConsensysStakingGraphQLClient(
      this.apolloClient,
      this.beaconNodeApiClient,
      new WinstonLogger(ConsensysStakingGraphQLClient.name, config.loggerOptions),
    );
    this.lidoAccountingReportClient = new LidoAccountingReportClient(
      this.lazyOracleContractClient,
      config.dataSources.ipfsBaseUrl,
      new WinstonLogger(LidoAccountingReportClient.name, config.loggerOptions),
      this.config.contractAddresses.lidoYieldProviderAddress, // TODO - Wrong address because can't get vault sync
    );
    this.beaconChainStakingClient = new BeaconChainStakingClient(
      this.consensysStakingGraphQLClient,
      config.maxValidatorWithdrawalRequestsPerTransaction,
      this.yieldManagerContractClient,
      this.config.contractAddresses.lidoYieldProviderAddress,
    );

    this.yieldReportingOperationModeProcessor = new YieldReportingOperationModeProcessor(
      this.yieldManagerContractClient,
      this.lazyOracleContractClient,
      this.lineaRollupYieldExtensionContractClient,
      this.lidoAccountingReportClient,
      this.beaconChainStakingClient,
      new WinstonLogger(YieldReportingOperationModeProcessor.name, config.loggerOptions),
      config.timing.trigger.maxInactionMs,
      config.contractAddresses.lidoYieldProviderAddress,
      config.contractAddresses.l2YieldRecipientAddress,
    );

    this.ossificationPendingOperationModeProcessor = new OssificationPendingOperationModeProcessor(
      this.yieldManagerContractClient,
      this.lazyOracleContractClient,
      this.lidoAccountingReportClient,
      this.beaconChainStakingClient,
      new WinstonLogger(OssificationPendingOperationModeProcessor.name, config.loggerOptions),
      config.timing.trigger.maxInactionMs,
      config.contractAddresses.lidoYieldProviderAddress,
    );

    this.ossificationCompleteOperationModeProcessor = new OssificationCompleteOperationModeProcessor(
      this.yieldManagerContractClient,
      this.beaconChainStakingClient,
      config.timing.trigger.maxInactionMs,
      config.contractAddresses.lidoYieldProviderAddress,
    );

    this.operationModeSelector = new OperationModeSelector(
      new WinstonLogger(OperationModeSelector.name, config.loggerOptions),
      this.yieldManagerContractClient,
      this.yieldReportingOperationModeProcessor,
      this.ossificationPendingOperationModeProcessor,
      this.ossificationCompleteOperationModeProcessor,
      config.contractAddresses.lidoYieldProviderAddress,
      config.timing.contractReadRetryTimeMs,
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
