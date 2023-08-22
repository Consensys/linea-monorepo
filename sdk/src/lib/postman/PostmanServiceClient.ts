import { JsonRpcProvider } from "@ethersproject/providers";
import { Wallet } from "ethers";
import { DataSource } from "typeorm";
import { PostmanConfig } from "./utils/types";
import { LineaLogger, getLogger } from "../logger";
import { L1MessageServiceContract, L2MessageServiceContract } from "../contracts";
import { DB } from "./dataSource";
import { L1SentEventListener, L2AnchoredEventListener } from "./listeners";
import { L2ClaimTxSender, L2ClaimStatusWatcher } from "./transactions";

export class PostmanServiceClient {
  private l1EventListener: L1SentEventListener;
  private l2AnchoredEventListener: L2AnchoredEventListener;
  private l2ClaimTxSender: L2ClaimTxSender;
  private l2ClaimStatusWatcher: L2ClaimStatusWatcher;

  private logger: LineaLogger;
  private db: DataSource;

  constructor(private readonly config: PostmanConfig) {
    this.logger = getLogger(PostmanServiceClient.name, config?.loggerOptions);

    const l1Provider = new JsonRpcProvider(config.l1Config.rpcUrl);
    const l2Provider = new JsonRpcProvider(config.l2Config.rpcUrl);

    const l1Signer = this.getSigner(config.l1Config.claiming.signerPrivateKey, l1Provider);
    const l2Signer = this.getSigner(config.l2Config.claiming.signerPrivateKey, l2Provider);

    const l1MessageServiceContract = new L1MessageServiceContract(
      l1Provider,
      config.l1Config.messageServiceContractAddress,
      "read-write",
      l1Signer,
      config.l1Config.claiming.maxFeePerGas,
      config.l1Config.claiming.gasEstimationPercentile,
    );

    const l2MessageServiceContract = new L2MessageServiceContract(
      l2Provider,
      config.l2Config.messageServiceContractAddress,
      "read-write",
      l2Signer,
      config.l2Config.claiming.maxFeePerGas,
      config.l2Config.claiming.gasEstimationPercentile,
    );

    this.db = DB.create(this.config.databaseOptions);

    this.l1EventListener = new L1SentEventListener(
      this.db,
      l1MessageServiceContract,
      config.l1Config,
      config?.loggerOptions,
    );

    this.l2AnchoredEventListener = new L2AnchoredEventListener(
      this.db,
      l2MessageServiceContract,
      config.l2Config,
      config.l1Config.messageServiceContractAddress,
      config?.loggerOptions,
    );

    this.l2ClaimTxSender = new L2ClaimTxSender(
      this.db,
      l2MessageServiceContract,
      config.l2Config,
      config.l1Config.messageServiceContractAddress,
      config?.loggerOptions,
    );

    this.l2ClaimStatusWatcher = new L2ClaimStatusWatcher(
      this.db,
      l2MessageServiceContract,
      config.l2Config,
      config?.loggerOptions,
    );
  }

  private getSigner(privateKey: string, provider: JsonRpcProvider) {
    try {
      return new Wallet(privateKey, provider);
    } catch (e) {
      throw new Error(
        "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
      );
    }
  }

  public async connectDatabase() {
    await this.db.initialize();
  }

  public startAllServices(): void {
    this.l1EventListener.start();
    this.l2AnchoredEventListener.start();
    this.l2ClaimTxSender.start();
    this.l2ClaimStatusWatcher.start();
    this.logger.info("All listeners and message deliverers have been started.");
  }

  public stopAllServices(): void {
    this.l1EventListener.stop();
    this.l2AnchoredEventListener.stop();
    this.l2ClaimTxSender.stop();
    this.l2ClaimStatusWatcher.stop();
    this.logger.info("All listeners and message deliverers have been stopped.");
  }
}
