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
import { PostmanConfig } from "./config/config";
import { DB } from "../persistence/dataSource";
import { Direction } from "../../../core/enums/MessageEnums";
import { MessageClaimingProcessor } from "../../../services/processors/MessageClaimingProcessor";
import { MessageClaimingPoller } from "../../../services/pollers/MessageClaimingPoller";
import { MessageClaimingPersister } from "../../../services/processors/MessageClaimingPersister";
import { MessagePersistingPoller } from "../../../services/pollers/MessagePersistingPoller";
import { MessageSentEventProcessor } from "../../../services/processors/MessageSentEventProcessor";
import { DatabaseCleaningPoller } from "../../../services/pollers/DatabaseCleaningPoller";
import { BaseError } from "../../../core/errors/Base";

export class PostmanServiceClient {
  // L1 -> L2 flow
  private l1MessageSentEventPoller: IPoller;
  private l2MessageAnchoringPoller: IPoller;
  private l2MessageClaimingPoller: IPoller;
  private l2MessagePersistingPoller: IPoller;

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
   * @param {PostmanConfig} config - Configuration settings for the Postman service, including network details, database options, and logging configurations.
   */
  constructor(private readonly config: PostmanConfig) {
    this.logger = new WinstonLogger(PostmanServiceClient.name, config.loggerOptions);
    this.l1L2AutoClaimEnabled = config.l1L2AutoClaimEnabled;
    this.l2L1AutoClaimEnabled = config.l2L1AutoClaimEnabled;

    const l1Provider = new JsonRpcProvider(config.l1Config.rpcUrl);
    const l2Provider = new JsonRpcProvider(config.l2Config.rpcUrl);

    const l1Signer = this.getSigner(config.l1Config.claiming.signerPrivateKey, l1Provider);
    const l2Signer = this.getSigner(config.l2Config.claiming.signerPrivateKey, l2Provider);

    const l1Querier = new ChainQuerier(l1Provider, l1Signer);
    const l2Querier = new ChainQuerier(l2Provider, l2Signer);

    const lineaRollupLogClient = new EthersLineaRollupLogClient(
      l1Provider,
      config.l1Config.messageServiceContractAddress,
    );
    const l2MessageServiceLogClient = new EthersL2MessageServiceLogClient(
      l2Provider,
      config.l2Config.messageServiceContractAddress,
    );

    const l1MessageServiceContract = new LineaRollupClient(
      l1Provider,
      config.l1Config.messageServiceContractAddress,
      lineaRollupLogClient,
      l2MessageServiceLogClient,
      "read-write",
      l1Signer,
      config.l1Config.claiming.maxFeePerGas,
      config.l1Config.claiming.gasEstimationPercentile,
      config.l1Config.claiming.isMaxGasFeeEnforced,
      config.l2Config.l2MessageTreeDepth,
    );

    const l2MessageServiceContract = new L2MessageServiceClient(
      l2Provider,
      config.l2Config.messageServiceContractAddress,
      l2MessageServiceLogClient,
      "read-write",
      l2Signer,
      config.l2Config.claiming.maxFeePerGas,
      config.l2Config.claiming.gasEstimationPercentile,
      config.l2Config.claiming.isMaxGasFeeEnforced,
    );

    this.db = DB.create(this.config.databaseOptions);

    const messageRepository = new TypeOrmMessageRepository(this.db);

    // L1 -> L2 flow
    const l1MessageSentEventProcessor = new MessageSentEventProcessor(
      messageRepository,
      lineaRollupLogClient,
      l1Querier,
      config.l1Config,
      Direction.L1_TO_L2,
      new WinstonLogger(`L1${MessageSentEventProcessor.name}`, config.loggerOptions),
    );

    this.l1MessageSentEventPoller = new MessageSentEventPoller(
      l1MessageSentEventProcessor,
      l1Querier,
      messageRepository,
      Direction.L1_TO_L2,
      config.l1Config,
      new WinstonLogger(`L1${MessageSentEventPoller.name}`, config.loggerOptions),
    );

    const l2MessageAnchoringProcessor = new MessageAnchoringProcessor(
      messageRepository,
      l2MessageServiceContract,
      l2Querier,
      config.l2Config,
      Direction.L1_TO_L2,
      config.l1Config.messageServiceContractAddress,
      new WinstonLogger(`L2${MessageAnchoringProcessor.name}`, config.loggerOptions),
    );

    this.l2MessageAnchoringPoller = new MessageAnchoringPoller(
      l2MessageAnchoringProcessor,
      Direction.L1_TO_L2,
      config.l2Config,
      new WinstonLogger(`L2${MessageAnchoringPoller.name}`, config.loggerOptions),
    );

    const l2MessageClaimingProcessor = new MessageClaimingProcessor(
      messageRepository,
      l2MessageServiceContract,
      l2Querier,
      config.l2Config,
      Direction.L1_TO_L2,
      config.l1Config.messageServiceContractAddress,
      new WinstonLogger(`L2${MessageClaimingProcessor.name}`, config.loggerOptions),
    );

    this.l2MessageClaimingPoller = new MessageClaimingPoller(
      l2MessageClaimingProcessor,
      Direction.L1_TO_L2,
      config.l2Config,
      new WinstonLogger(`L2${MessageClaimingPoller.name}`, config.loggerOptions),
    );

    const l2MessageClaimingPersister = new MessageClaimingPersister(
      messageRepository,
      l2MessageServiceContract,
      l2Querier,
      config.l2Config,
      Direction.L1_TO_L2,
      new WinstonLogger(`L2${MessageClaimingPersister.name}`, config.loggerOptions),
    );

    this.l2MessagePersistingPoller = new MessagePersistingPoller(
      l2MessageClaimingPersister,
      Direction.L1_TO_L2,
      config.l2Config,
      new WinstonLogger(`L2${MessagePersistingPoller.name}`, config.loggerOptions),
    );

    // L2 -> L1 flow
    const l2MessageSentEventProcessor = new MessageSentEventProcessor(
      messageRepository,
      l2MessageServiceLogClient,
      l2Querier,
      config.l2Config,
      Direction.L2_TO_L1,
      new WinstonLogger(`L2${MessageSentEventProcessor.name}`, config.loggerOptions),
    );

    this.l2MessageSentEventPoller = new MessageSentEventPoller(
      l2MessageSentEventProcessor,
      l2Querier,
      messageRepository,
      Direction.L2_TO_L1,
      config.l2Config,
      new WinstonLogger(`L2${MessageSentEventPoller.name}`, config.loggerOptions),
    );

    const l1MessageAnchoringProcessor = new MessageAnchoringProcessor(
      messageRepository,
      l1MessageServiceContract,
      l1Querier,
      config.l1Config,
      Direction.L2_TO_L1,
      config.l2Config.messageServiceContractAddress,
      new WinstonLogger(`L1${MessageAnchoringProcessor.name}`, config.loggerOptions),
    );

    this.l1MessageAnchoringPoller = new MessageAnchoringPoller(
      l1MessageAnchoringProcessor,
      Direction.L2_TO_L1,
      config.l1Config,
      new WinstonLogger(`L1${MessageAnchoringPoller.name}`, config.loggerOptions),
    );

    const l1MessageClaimingProcessor = new MessageClaimingProcessor(
      messageRepository,
      l1MessageServiceContract,
      l1Querier,
      config.l1Config,
      Direction.L2_TO_L1,
      config.l2Config.messageServiceContractAddress,
      new WinstonLogger(`L1${MessageClaimingProcessor.name}`, config.loggerOptions),
    );

    this.l1MessageClaimingPoller = new MessageClaimingPoller(
      l1MessageClaimingProcessor,
      Direction.L2_TO_L1,
      config.l1Config,
      new WinstonLogger(`L1${MessageClaimingPoller.name}`, config.loggerOptions),
    );

    const l1MessageClaimingPersister = new MessageClaimingPersister(
      messageRepository,
      l1MessageServiceContract,
      l1Querier,
      config.l1Config,
      Direction.L2_TO_L1,
      new WinstonLogger(`L1${MessageClaimingPersister.name}`, config.loggerOptions),
    );

    this.l1MessagePersistingPoller = new MessagePersistingPoller(
      l1MessageClaimingPersister,
      Direction.L2_TO_L1,
      config.l1Config,
      new WinstonLogger(`L1${MessagePersistingPoller.name}`, config.loggerOptions),
    );

    // Database Cleaner
    const databaseCleaner = new DatabaseCleaner(
      messageRepository,
      new WinstonLogger(`${DatabaseCleaner.name}`, config.loggerOptions),
    );

    this.databaseCleaningPoller = new DatabaseCleaningPoller(
      databaseCleaner,
      new WinstonLogger(`${DatabaseCleaningPoller.name}`, config.loggerOptions),
      config?.databaseCleanerConfig,
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
