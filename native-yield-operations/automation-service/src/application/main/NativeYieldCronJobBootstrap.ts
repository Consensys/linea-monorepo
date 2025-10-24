import { ILogger, WinstonLogger } from "@consensys/linea-shared-utils";
import { NativeYieldCronJobBootstrapConfig } from "./config/config.js";
import { IOperationModeSelector } from "../../core/services/operation-mode/IOperationModeSelector.js";
import { OperationModeSelector } from "../../services/OperationModeSelector.js";
import {
  IBlockchainClient,
  ViemBlockchainClientAdapter,
  Web3SignerClient,
  IContractSignerClient,
  IOAuth2TokenClient,
  IBeaconNodeAPIClient,
  BeaconNodeApiClient,
  OAuth2TokenClient,
} from "@consensys/linea-shared-utils";
import {} from "@consensys/linea-shared-utils";
import { Chain, PublicClient, TransactionReceipt } from "viem";
import { YieldManagerContractClient } from "../../clients/contracts/YieldManagerContractClient.js";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { YieldReportingProcessor } from "../../services/operation-mode-processors/YieldReportingProcessor.js";
import { LazyOracleContractClient } from "../../clients/contracts/LazyOracleContractClient.js";
import { ILazyOracle } from "../../core/clients/contracts/ILazyOracle.js";
import { ApolloClient } from "@apollo/client";
import { ILineaRollupYieldExtension } from "../../core/clients/contracts/ILineaRollupYieldExtension.js";
import { LineaRollupYieldExtensionContractClient } from "../../clients/contracts/LineaRollupYieldExtensionContractClient.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient.js";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";
import { IValidatorDataClient } from "../../core/clients/IValidatorDataClient.js";
import { ConsensysStakingApiClient } from "../../clients/ConsensysStakingApiClient.js";
import { LidoAccountingReportClient } from "../../clients/LidoAccountingReportClient.js";
import { BeaconChainStakingClient } from "../../clients/BeaconChainStakingClient.js";
import { OssificationCompleteProcessor } from "../../services/operation-mode-processors/OssificationCompleteProcessor.js";
import { OssificationPendingProcessor } from "../../services/operation-mode-processors/OssificationPendingProcessor.js";
import { mainnet, hoodi } from "viem/chains";
import { createApolloClient } from "../../utils/createApolloClient.js";

export class NativeYieldCronJobBootstrap {
  private readonly config: NativeYieldCronJobBootstrapConfig;
  private readonly logger: ILogger;

  private ViemBlockchainClientAdapter: IBlockchainClient<PublicClient, TransactionReceipt>;
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

  constructor(config: NativeYieldCronJobBootstrapConfig) {
    this.config = config;
    this.logger = new WinstonLogger(NativeYieldCronJobBootstrap.name, config.loggerOptions);

    this.web3SignerClient = new Web3SignerClient(
      new WinstonLogger(Web3SignerClient.name, config.loggerOptions),
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
    this.ViemBlockchainClientAdapter = new ViemBlockchainClientAdapter(
      new WinstonLogger(ViemBlockchainClientAdapter.name, config.loggerOptions),
      config.dataSources.l1RpcUrl,
      getChain(config.dataSources.chainId),
      this.web3SignerClient,
    );
    this.yieldManagerContractClient = new YieldManagerContractClient(
      new WinstonLogger(YieldManagerContractClient.name, config.loggerOptions),
      this.ViemBlockchainClientAdapter,
      config.contractAddresses.yieldManagerAddress,
      config.rebalanceToleranceBps,
      config.minWithdrawalThresholdEth,
    );
    this.lazyOracleContractClient = new LazyOracleContractClient(
      new WinstonLogger(LazyOracleContractClient.name, config.loggerOptions),
      this.ViemBlockchainClientAdapter,
      config.contractAddresses.lazyOracleAddress,
      config.timing.trigger.maxInactionMs,
    );
    this.lineaRollupYieldExtensionContractClient = new LineaRollupYieldExtensionContractClient(
      new WinstonLogger(LineaRollupYieldExtensionContractClient.name, config.loggerOptions),
      this.ViemBlockchainClientAdapter,
      config.contractAddresses.lineaRollupContractAddress,
    );

    this.beaconNodeApiClient = new BeaconNodeApiClient(
      new WinstonLogger(BeaconNodeApiClient.name, config.loggerOptions),
      config.dataSources.beaconChainRpcUrl,
    );
    this.oAuth2TokenClient = new OAuth2TokenClient(
      new WinstonLogger(OAuth2TokenClient.name, config.loggerOptions),
      config.consensysStakingOAuth2.tokenEndpoint,
      config.consensysStakingOAuth2.clientId,
      config.consensysStakingOAuth2.clientSecret,
      config.consensysStakingOAuth2.audience,
    );
    this.apolloClient = createApolloClient(this.oAuth2TokenClient, config.dataSources.stakingGraphQLUrl);
    this.consensysStakingGraphQLClient = new ConsensysStakingApiClient(
      new WinstonLogger(ConsensysStakingApiClient.name, config.loggerOptions),
      this.apolloClient,
      this.beaconNodeApiClient,
    );
    this.lidoAccountingReportClient = new LidoAccountingReportClient(
      new WinstonLogger(LidoAccountingReportClient.name, config.loggerOptions),
      this.lazyOracleContractClient,
      config.dataSources.ipfsBaseUrl,
      this.config.contractAddresses.lidoYieldProviderAddress, // TODO - Wrong address because can't get vault sync
    );
    this.beaconChainStakingClient = new BeaconChainStakingClient(
      new WinstonLogger(BeaconChainStakingClient.name, config.loggerOptions),
      this.consensysStakingGraphQLClient,
      config.maxValidatorWithdrawalRequestsPerTransaction,
      this.yieldManagerContractClient,
      this.config.contractAddresses.lidoYieldProviderAddress,
    );

    this.yieldReportingOperationModeProcessor = new YieldReportingProcessor(
      new WinstonLogger(YieldReportingProcessor.name, config.loggerOptions),
      this.yieldManagerContractClient,
      this.lazyOracleContractClient,
      this.lineaRollupYieldExtensionContractClient,
      this.lidoAccountingReportClient,
      this.beaconChainStakingClient,
      config.timing.trigger.maxInactionMs,
      config.contractAddresses.lidoYieldProviderAddress,
      config.contractAddresses.l2YieldRecipientAddress,
    );

    this.ossificationPendingOperationModeProcessor = new OssificationPendingProcessor(
      new WinstonLogger(OssificationPendingProcessor.name, config.loggerOptions),
      this.yieldManagerContractClient,
      this.lazyOracleContractClient,
      this.lidoAccountingReportClient,
      this.beaconChainStakingClient,
      config.timing.trigger.maxInactionMs,
      config.contractAddresses.lidoYieldProviderAddress,
    );

    this.ossificationCompleteOperationModeProcessor = new OssificationCompleteProcessor(
      new WinstonLogger(OssificationCompleteProcessor.name, config.loggerOptions),
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
    this.logger.info("Native yield automation service started");
  }

  public stopAllServices(): void {
    this.operationModeSelector.stop();
    this.logger.info("Native yield automation service stopped");
  }

  public getConfig(): NativeYieldCronJobBootstrapConfig {
    return this.config;
  }
}
