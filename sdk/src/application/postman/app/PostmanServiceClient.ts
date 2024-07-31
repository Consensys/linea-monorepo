import { Wallet, JsonRpcProvider } from "ethers";
import { DataSource } from "typeorm";
import { ILogger } from "../../../core/utils/logging/ILogger";
import { DatabaseCleaner } from "../../../services/persistence/DatabaseCleaner";
import { TypeOrmMessageRepository } from "../persistence/repositories/TypeOrmMessageRepository";
import { LineaRollupClient } from "../../../clients/blockchain/ethereum/LineaRollupClient";
import { L2MessageServiceClient } from "../../../clients/blockchain/linea/L2MessageServiceClient";
import { EthersLineaRollupLogClient } from "../../../clients/blockchain/ethereum/EthersLineaRollupLogClient";
import { ChainQuerier } from "../../../clients/blockchain/ChainQuerier";
import { WinstonLogger } from "../../../utils/WinstonLogger";
import { EthersL2MessageServiceLogClient } from "../../../clients/blockchain/linea/EthersL2MessageServiceLogClient";
import { MessageSentEventPoller } from "../../../services/pollers/MessageSentEventPoller";
import { IPoller } from "../../../core/services/pollers/IPoller";
import { MessageAnchoringPoller } from "../../../services/pollers/MessageAnchoringPoller";
import { MessageAnchoringProcessor } from "../../../services/processors/MessageAnchoringProcessor";
import { PostmanOptions } from "./config/config";
import { DB } from "../persistence/dataSource";
import { Direction } from "../../../core/enums/MessageEnums";
import { MessageClaimingProcessor } from "../../../services/processors/MessageClaimingProcessor";
import { MessageClaimingPoller } from "../../../services/pollers/MessageClaimingPoller";
import { MessageClaimingPersister } from "../../../services/processors/MessageClaimingPersister";
import { MessagePersistingPoller } from "../../../services/pollers/MessagePersistingPoller";
import { MessageSentEventProcessor } from "../../../services/processors/MessageSentEventProcessor";
import { DatabaseCleaningPoller } from "../../../services/pollers/DatabaseCleaningPoller";
import { BaseError } from "../../../core/errors/Base";
import { LineaRollupMessageRetriever } from "../../../clients/blockchain/ethereum/LineaRollupMessageRetriever";
import { L2MessageServiceMessageRetriever } from "../../../clients/blockchain/linea/L2MessageServiceMessageRetriever";
import { MerkleTreeService } from "../../../clients/blockchain/ethereum/MerkleTreeService";
import { LineaMessageDBService } from "../../../services/persistence/LineaMessageDBService";
import { L2ChainQuerier } from "../../../clients/blockchain/linea/L2ChainQuerier";
import { EthereumMessageDBService } from "../../../services/persistence/EthereumMessageDBService";
import { L2ClaimMessageTransactionSizePoller } from "../../../services/pollers/L2ClaimMessageTransactionSizePoller";
import { L2ClaimMessageTransactionSizeProcessor } from "../../../services/processors/L2ClaimMessageTransactionSizeProcessor";
import { L2ClaimTransactionSizeCalculator } from "../../../services/L2ClaimTransactionSizeCalculator";
import { GasProvider } from "../../../clients/blockchain/gas/GasProvider";
import { LineaTransactionValidationService } from "../../../services/LineaTransactionValidationService";
import { EthereumTransactionValidationService } from "../../../services/EthereumTransactionValidationService";
import { getConfig } from "./config/utils";

export class PostmanServiceClient {
  // L1 -> L2 flow
  private l1MessageSentEventPoller: IPoller;
  private l2MessageAnchoringPoller: IPoller;
  private l2MessageClaimingPoller: IPoller;
  private l2MessagePersistingPoller: IPoller;
  private l2ClaimMessageTransactionSizePoller: IPoller;

  // L2 -> L1 flow
  private l2MessageSentEventPoller: IPoller;
  private l1MessageAnchoringPoller: IPoller;
  private l1MessageClaimingPoller: IPoller;
  private l1MessagePersistingPoller: IPoller;

  // Database Cleaner
  private databaseCleaningPoller: IPoller;

  private logger: ILogger;
  private db: DataSource;

  private l1L2AutoClaimEnabled: boolean;
  private l2L1AutoClaimEnabled: boolean;

  /**
   * Initializes a new instance of the PostmanServiceClient.
   *
   * @param {PostmanOptions} options - Configuration options for the Postman service, including network details, database options, and logging configurations.
   */
  constructor(options: PostmanOptions) {
    const config = getConfig(options);

    this.logger = new WinstonLogger(PostmanServiceClient.name, config.loggerOptions);
    this.l1L2AutoClaimEnabled = config.l1L2AutoClaimEnabled;
    this.l2L1AutoClaimEnabled = config.l2L1AutoClaimEnabled;

    const l1Provider = new JsonRpcProvider(config.l1Config.rpcUrl);
    const l2Provider = new JsonRpcProvider(config.l2Config.rpcUrl);

    const l1Signer = this.getSigner(config.l1Config.claiming.signerPrivateKey, l1Provider);
    const l2Signer = this.getSigner(config.l2Config.claiming.signerPrivateKey, l2Provider);

    const l1Querier = new ChainQuerier(l1Provider, l1Signer);
    const l2Querier = new L2ChainQuerier(l2Provider, l2Signer);

    const lineaRollupLogClient = new EthersLineaRollupLogClient(
      l1Provider,
      config.l1Config.messageServiceContractAddress,
    );
    const l2MessageServiceLogClient = new EthersL2MessageServiceLogClient(
      l2Provider,
      config.l2Config.messageServiceContractAddress,
    );

    const l1GasProvider = new GasProvider(l1Querier, {
      maxFeePerGas: config.l1Config.claiming.maxFeePerGas,
      gasEstimationPercentile: config.l1Config.claiming.gasEstimationPercentile,
      enforceMaxGasFee: config.l1Config.claiming.isMaxGasFeeEnforced,
      enableLineaEstimateGas: false,
      direction: Direction.L2_TO_L1,
    });

    const l2GasProvider = new GasProvider(l2Querier, {
      maxFeePerGas: config.l2Config.claiming.maxFeePerGas,
      gasEstimationPercentile: config.l2Config.claiming.gasEstimationPercentile,
      enforceMaxGasFee: config.l2Config.claiming.isMaxGasFeeEnforced,
      enableLineaEstimateGas: config.l2Config.enableLineaEstimateGas,
      direction: Direction.L1_TO_L2,
    });

    const lineaRollupMessageRetriever = new LineaRollupMessageRetriever(
      l1Querier,
      lineaRollupLogClient,
      config.l1Config.messageServiceContractAddress,
    );

    const l1MerkleTreeService = new MerkleTreeService(
      l1Querier,
      config.l1Config.messageServiceContractAddress,
      lineaRollupLogClient,
      l2MessageServiceLogClient,
      config.l2Config.l2MessageTreeDepth,
    );

    const l2MessageServiceMessageRetriever = new L2MessageServiceMessageRetriever(
      l2Querier,
      l2MessageServiceLogClient,
      config.l2Config.messageServiceContractAddress,
    );

    const l1MessageServiceContract = new LineaRollupClient(
      l1Querier,
      config.l1Config.messageServiceContractAddress,
      lineaRollupLogClient,
      l2MessageServiceLogClient,
      l1GasProvider,
      lineaRollupMessageRetriever,
      l1MerkleTreeService,
      "read-write",
      l1Signer,
    );

    const l2MessageServiceContract = new L2MessageServiceClient(
      l2Querier,
      config.l2Config.messageServiceContractAddress,
      l2MessageServiceMessageRetriever,
      l2GasProvider,
      "read-write",
      l2Signer,
    );

    this.db = DB.create(config.databaseOptions);

    const messageRepository = new TypeOrmMessageRepository(this.db);
    const lineaMessageDBService = new LineaMessageDBService(l2Querier, messageRepository);
    const ethereumMessageDBService = new EthereumMessageDBService(l1GasProvider, messageRepository);

    // L1 -> L2 flow

    const l1MessageSentEventProcessor = new MessageSentEventProcessor(
      lineaMessageDBService,
      lineaRollupLogClient,
      l1Querier,
      {
        direction: Direction.L1_TO_L2,
        maxBlocksToFetchLogs: config.l1Config.listener.maxBlocksToFetchLogs,
        blockConfirmation: config.l1Config.listener.blockConfirmation,
        isEOAEnabled: config.l1Config.isEOAEnabled,
        isCalldataEnabled: config.l1Config.isCalldataEnabled,
      },
      new WinstonLogger(`L1${MessageSentEventProcessor.name}`, config.loggerOptions),
    );

    this.l1MessageSentEventPoller = new MessageSentEventPoller(
      l1MessageSentEventProcessor,
      l1Querier,
      lineaMessageDBService,
      {
        direction: Direction.L1_TO_L2,
        pollingInterval: config.l1Config.listener.pollingInterval,
        initialFromBlock: config.l1Config.listener.initialFromBlock,
        originContractAddress: config.l1Config.messageServiceContractAddress,
      },
      new WinstonLogger(`L1${MessageSentEventPoller.name}`, config.loggerOptions),
    );

    const l2MessageAnchoringProcessor = new MessageAnchoringProcessor(
      l2MessageServiceContract,
      l2Querier,
      lineaMessageDBService,
      {
        maxFetchMessagesFromDb: config.l1Config.listener.maxFetchMessagesFromDb,
        originContractAddress: config.l1Config.messageServiceContractAddress,
      },
      new WinstonLogger(`L2${MessageAnchoringProcessor.name}`, config.loggerOptions),
    );

    this.l2MessageAnchoringPoller = new MessageAnchoringPoller(
      l2MessageAnchoringProcessor,
      {
        direction: Direction.L1_TO_L2,
        pollingInterval: config.l2Config.listener.pollingInterval,
      },
      new WinstonLogger(`L2${MessageAnchoringPoller.name}`, config.loggerOptions),
    );

    const l2TransactionValidationService = new LineaTransactionValidationService(
      {
        profitMargin: config.l2Config.claiming.profitMargin,
        maxClaimGasLimit: BigInt(config.l2Config.claiming.maxClaimGasLimit),
      },
      l2Querier,
      l2MessageServiceContract,
    );

    const l2MessageClaimingProcessor = new MessageClaimingProcessor(
      l2MessageServiceContract,
      l2Querier,
      lineaMessageDBService,
      l2TransactionValidationService,
      {
        direction: Direction.L1_TO_L2,
        originContractAddress: config.l1Config.messageServiceContractAddress,
        maxNonceDiff: config.l2Config.claiming.maxNonceDiff,
        feeRecipientAddress: config.l2Config.claiming.feeRecipientAddress,
        profitMargin: config.l2Config.claiming.profitMargin,
        maxNumberOfRetries: config.l2Config.claiming.maxNumberOfRetries,
        retryDelayInSeconds: config.l2Config.claiming.retryDelayInSeconds,
        maxClaimGasLimit: BigInt(config.l2Config.claiming.maxClaimGasLimit),
      },
      new WinstonLogger(`L2${MessageClaimingProcessor.name}`, config.loggerOptions),
    );

    this.l2MessageClaimingPoller = new MessageClaimingPoller(
      l2MessageClaimingProcessor,
      {
        direction: Direction.L1_TO_L2,
        pollingInterval: config.l2Config.listener.pollingInterval,
      },
      new WinstonLogger(`L2${MessageClaimingPoller.name}`, config.loggerOptions),
    );

    const l2MessageClaimingPersister = new MessageClaimingPersister(
      lineaMessageDBService,
      l2MessageServiceContract,
      l2Querier,
      {
        direction: Direction.L1_TO_L2,
        messageSubmissionTimeout: config.l2Config.claiming.messageSubmissionTimeout,
        maxTxRetries: config.l2Config.claiming.maxTxRetries,
      },
      new WinstonLogger(`L2${MessageClaimingPersister.name}`, config.loggerOptions),
    );

    this.l2MessagePersistingPoller = new MessagePersistingPoller(
      l2MessageClaimingPersister,
      {
        direction: Direction.L1_TO_L2,
        pollingInterval: config.l2Config.listener.pollingInterval,
      },
      new WinstonLogger(`L2${MessagePersistingPoller.name}`, config.loggerOptions),
    );

    const transactionSizeCalculator = new L2ClaimTransactionSizeCalculator(l2MessageServiceContract);
    const transactionSizeCompressor = new L2ClaimMessageTransactionSizeProcessor(
      lineaMessageDBService,
      l2MessageServiceContract,
      transactionSizeCalculator,
      {
        direction: Direction.L1_TO_L2,
        originContractAddress: config.l1Config.messageServiceContractAddress,
      },
      new WinstonLogger(`${L2ClaimMessageTransactionSizeProcessor.name}`, config.loggerOptions),
    );

    this.l2ClaimMessageTransactionSizePoller = new L2ClaimMessageTransactionSizePoller(
      transactionSizeCompressor,
      {
        pollingInterval: config.l2Config.listener.pollingInterval,
      },
      new WinstonLogger(`${L2ClaimMessageTransactionSizePoller.name}`, config.loggerOptions),
    );

    // L2 -> L1 flow
    const l2MessageSentEventProcessor = new MessageSentEventProcessor(
      ethereumMessageDBService,
      l2MessageServiceLogClient,
      l2Querier,
      {
        direction: Direction.L2_TO_L1,
        maxBlocksToFetchLogs: config.l2Config.listener.maxBlocksToFetchLogs,
        blockConfirmation: config.l2Config.listener.blockConfirmation,
        isEOAEnabled: config.l2Config.isEOAEnabled,
        isCalldataEnabled: config.l2Config.isCalldataEnabled,
      },
      new WinstonLogger(`L2${MessageSentEventProcessor.name}`, config.loggerOptions),
    );

    this.l2MessageSentEventPoller = new MessageSentEventPoller(
      l2MessageSentEventProcessor,
      l2Querier,
      ethereumMessageDBService,
      {
        direction: Direction.L2_TO_L1,
        pollingInterval: config.l2Config.listener.pollingInterval,
        initialFromBlock: config.l2Config.listener.initialFromBlock,
        originContractAddress: config.l2Config.messageServiceContractAddress,
      },
      new WinstonLogger(`L2${MessageSentEventPoller.name}`, config.loggerOptions),
    );

    const l1MessageAnchoringProcessor = new MessageAnchoringProcessor(
      l1MessageServiceContract,
      l1Querier,
      ethereumMessageDBService,
      {
        maxFetchMessagesFromDb: config.l1Config.listener.maxFetchMessagesFromDb,
        originContractAddress: config.l2Config.messageServiceContractAddress,
      },
      new WinstonLogger(`L1${MessageAnchoringProcessor.name}`, config.loggerOptions),
    );

    this.l1MessageAnchoringPoller = new MessageAnchoringPoller(
      l1MessageAnchoringProcessor,
      {
        direction: Direction.L2_TO_L1,
        pollingInterval: config.l1Config.listener.pollingInterval,
      },
      new WinstonLogger(`L1${MessageAnchoringPoller.name}`, config.loggerOptions),
    );

    const l1TransactionValidationService = new EthereumTransactionValidationService(
      l1MessageServiceContract,
      l1GasProvider,
      {
        profitMargin: config.l1Config.claiming.profitMargin,
        maxClaimGasLimit: BigInt(config.l1Config.claiming.maxClaimGasLimit),
      },
    );

    const l1MessageClaimingProcessor = new MessageClaimingProcessor(
      l1MessageServiceContract,
      l1Querier,
      ethereumMessageDBService,
      l1TransactionValidationService,
      {
        direction: Direction.L2_TO_L1,
        maxNonceDiff: config.l1Config.claiming.maxNonceDiff,
        feeRecipientAddress: config.l1Config.claiming.feeRecipientAddress,
        profitMargin: config.l1Config.claiming.profitMargin,
        maxNumberOfRetries: config.l1Config.claiming.maxNumberOfRetries,
        retryDelayInSeconds: config.l1Config.claiming.retryDelayInSeconds,
        maxClaimGasLimit: BigInt(config.l1Config.claiming.maxClaimGasLimit),
        originContractAddress: config.l2Config.messageServiceContractAddress,
      },
      new WinstonLogger(`L1${MessageClaimingProcessor.name}`, config.loggerOptions),
    );

    this.l1MessageClaimingPoller = new MessageClaimingPoller(
      l1MessageClaimingProcessor,
      {
        direction: Direction.L2_TO_L1,
        pollingInterval: config.l1Config.listener.pollingInterval,
      },
      new WinstonLogger(`L1${MessageClaimingPoller.name}`, config.loggerOptions),
    );

    const l1MessageClaimingPersister = new MessageClaimingPersister(
      ethereumMessageDBService,
      l1MessageServiceContract,
      l1Querier,
      {
        direction: Direction.L2_TO_L1,
        messageSubmissionTimeout: config.l1Config.claiming.messageSubmissionTimeout,
        maxTxRetries: config.l1Config.claiming.maxTxRetries,
      },
      new WinstonLogger(`L1${MessageClaimingPersister.name}`, config.loggerOptions),
    );

    this.l1MessagePersistingPoller = new MessagePersistingPoller(
      l1MessageClaimingPersister,
      {
        direction: Direction.L2_TO_L1,
        pollingInterval: config.l1Config.listener.pollingInterval,
      },
      new WinstonLogger(`L1${MessagePersistingPoller.name}`, config.loggerOptions),
    );

    // Database Cleaner
    const databaseCleaner = new DatabaseCleaner(
      ethereumMessageDBService,
      new WinstonLogger(`${DatabaseCleaner.name}`, config.loggerOptions),
    );

    this.databaseCleaningPoller = new DatabaseCleaningPoller(
      databaseCleaner,
      new WinstonLogger(`${DatabaseCleaningPoller.name}`, config.loggerOptions),
      {
        enabled: config.databaseCleanerConfig.enabled,
        daysBeforeNowToDelete: config.databaseCleanerConfig.daysBeforeNowToDelete,
        cleaningInterval: config.databaseCleanerConfig.cleaningInterval,
      },
    );
  }

  /**
   * Creates a Wallet instance as a signer using the provided private key and JSON RPC provider.
   *
   * @param {string} privateKey - The private key to use for the signer.
   * @param {JsonRpcProvider} provider - The JSON RPC provider associated with the network.
   * @returns {Wallet} A Wallet instance configured with the provided private key and provider.
   */
  private getSigner(privateKey: string, provider: JsonRpcProvider): Wallet {
    try {
      return new Wallet(privateKey, provider);
    } catch (e) {
      throw new BaseError(
        "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
      );
    }
  }

  /**
   * Initializes the database connection using the configuration provided.
   */
  public async connectDatabase() {
    await this.db.initialize();
  }

  /**
   * Starts all configured services and pollers. This includes message event pollers for both L1 to L2 and L2 to L1 flows, message anchoring, claiming, persisting pollers, and the database cleaning poller.
   */
  public startAllServices(): void {
    if (this.l1L2AutoClaimEnabled) {
      // L1 -> L2 flow
      this.l1MessageSentEventPoller.start();
      this.l2MessageAnchoringPoller.start();
      this.l2MessageClaimingPoller.start();
      this.l2MessagePersistingPoller.start();
      this.l2ClaimMessageTransactionSizePoller.start();
    }

    if (this.l2L1AutoClaimEnabled) {
      // L2 -> L1 flow
      this.l2MessageSentEventPoller.start();
      this.l1MessageAnchoringPoller.start();
      this.l1MessageClaimingPoller.start();
      this.l1MessagePersistingPoller.start();
    }

    // Database Cleaner
    this.databaseCleaningPoller.start();

    this.logger.info("All listeners and message deliverers have been started.");
  }

  /**
   * Stops all running services and pollers to gracefully shut down the Postman service.
   */
  public stopAllServices(): void {
    if (this.l1L2AutoClaimEnabled) {
      // L1 -> L2 flow
      this.l1MessageSentEventPoller.stop();
      this.l2MessageAnchoringPoller.stop();
      this.l2MessageClaimingPoller.stop();
      this.l2MessagePersistingPoller.stop();
      this.l2ClaimMessageTransactionSizePoller.stop();
    }

    if (this.l2L1AutoClaimEnabled) {
      // L2 -> L1 flow
      this.l2MessageSentEventPoller.stop();
      this.l1MessageAnchoringPoller.stop();
      this.l1MessageClaimingPoller.stop();
      this.l1MessagePersistingPoller.stop();
    }

    // Database Cleaner
    this.databaseCleaningPoller.stop();

    this.logger.info("All listeners and message deliverers have been stopped.");
  }
}
