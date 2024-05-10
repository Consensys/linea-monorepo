import { Wallet } from "ethers";
import { EthersLineaRollupLogClient } from "../clients/blockchain/ethereum/EthersLineaRollupLogClient";
import { LineaRollupClient } from "../clients/blockchain/ethereum/LineaRollupClient";
import { EthersL2MessageServiceLogClient } from "../clients/blockchain/linea/EthersL2MessageServiceLogClient";
import { L2MessageServiceClient } from "../clients/blockchain/linea/L2MessageServiceClient";
import { DEFAULT_L2_MESSAGE_TREE_DEPTH } from "../core/constants";
import { BaseError } from "../core/errors/Base";
import { L1ClaimingService } from "./claiming/L1ClaimingService";
import { LineaSDKOptions, Network, SDKMode } from "./config";
import { NETWORKS } from "./networks";
import ProviderService from "./ProviderService";

export class LineaSDK {
  private network: Network;
  private l1SignerPrivateKeyOrWallet: string | Wallet;
  private l2SignerPrivateKeyOrWallet: string | Wallet;
  private l1Provider: ProviderService;
  private l2Provider: ProviderService;
  private maxFeePerGas?: bigint;
  private gasFeeEstimationPercentile?: number;
  private mode: SDKMode;
  private l2MessageTreeDepth: number;

  /**
   * Initializes a new instance of the LineaSDK with the specified options.
   *
   * @param {LineaSDKOptions} options - Configuration options for the SDK, including network details, operational mode, and optional settings for L2 message tree depth and fee estimation.
   */
  constructor(options: LineaSDKOptions) {
    this.network = options.network;
    this.mode = options.mode;
    this.l2MessageTreeDepth = options.l2MessageTreeDepth ?? DEFAULT_L2_MESSAGE_TREE_DEPTH;

    if (options.mode === "read-write") {
      this.l1SignerPrivateKeyOrWallet = options.l1SignerPrivateKeyOrWallet;
      this.l2SignerPrivateKeyOrWallet = options.l2SignerPrivateKeyOrWallet;
      this.maxFeePerGas = options.feeEstimatorOptions?.maxFeePerGas;
      this.gasFeeEstimationPercentile = options.feeEstimatorOptions?.gasFeeEstimationPercentile;
    }

    this.l1Provider = new ProviderService(options.l1RpcUrlOrProvider);
    this.l2Provider = new ProviderService(options.l2RpcUrlOrProvider);
  }

  /**
   * Creates an instance of the `EthersL2MessageServiceLogClient` for interacting with L2 contract event logs.
   *
   * @param {string} [localL2ContractAddress] - Optional custom L2 contract address. Required if the network is set to 'custom'.
   * @returns {EthersL2MessageServiceLogClient} An instance of the L2 message service log client.
   */
  public getL2ContractEventLogClient(localL2ContractAddress?: string): EthersL2MessageServiceLogClient {
    let l2ContractAddress: string;

    if (this.network === "custom") {
      if (!localL2ContractAddress) {
        throw new BaseError("You need to provide a L2 contract address.");
      }
      l2ContractAddress = localL2ContractAddress;
    } else {
      l2ContractAddress = NETWORKS[this.network].l2ContractAddress;
    }

    return new EthersL2MessageServiceLogClient(this.l2Provider.provider, l2ContractAddress);
  }

  /**
   * Retrieves an instance of the `LineaRollupClient` for interacting with the L1 contract.
   *
   * @param {string} [localL1ContractAddress] - Optional custom L1 contract address. Required if the network is set to 'custom'.
   * @param {string} [localL2ContractAddress] - Optional custom L2 contract address. Required if the network is set to 'custom'.
   * @returns {LineaRollupClient} An instance of the `LineaRollupClient` configured for the specified L1 contract.
   */
  public getL1Contract(localL1ContractAddress?: string, localL2ContractAddress?: string): LineaRollupClient {
    let l1ContractAddress: string;
    let l2ContractAddress: string;

    if (this.network === "custom") {
      if (!localL1ContractAddress || !localL2ContractAddress) {
        throw new BaseError("You need to provide both L1 and L2 contract addresses.");
      }
      l1ContractAddress = localL1ContractAddress;
      l2ContractAddress = localL2ContractAddress;
    } else {
      l1ContractAddress = NETWORKS[this.network].l1ContractAddress;
      l2ContractAddress = NETWORKS[this.network].l2ContractAddress;
    }

    const signer = this.mode === "read-write" ? this.l1Provider.getSigner(this.l1SignerPrivateKeyOrWallet) : undefined;

    const lineaRollupLogClient = new EthersLineaRollupLogClient(this.l1Provider.provider, l1ContractAddress);

    return new LineaRollupClient(
      this.l1Provider.provider,
      l1ContractAddress,
      lineaRollupLogClient,
      this.getL2ContractEventLogClient(l2ContractAddress),
      this.mode,
      signer,
      this.maxFeePerGas,
      this.gasFeeEstimationPercentile,
      false,
      this.l2MessageTreeDepth,
    );
  }

  /**
   * Retrieves an instance of the `L2MessageServiceClient` for interacting with the L2 contract.
   *
   * @param {string} [localContractAddress] - Optional custom L2 contract address. Required if the network is set to 'custom'.
   * @returns {L2MessageServiceClient} An instance of the `L2MessageServiceClient` configured for the specified L2 contract.
   */
  public getL2Contract(localContractAddress?: string): L2MessageServiceClient {
    let l2ContractAddress: string;

    if (this.network === "custom") {
      if (!localContractAddress) {
        throw new BaseError("You need to provide a L2 contract address.");
      }
      l2ContractAddress = localContractAddress;
    } else {
      l2ContractAddress = NETWORKS[this.network].l2ContractAddress;
    }

    const signer = this.mode === "read-write" ? this.l2Provider.getSigner(this.l2SignerPrivateKeyOrWallet) : undefined;

    const l2MessageServiceContract = new L2MessageServiceClient(
      this.l2Provider.provider,
      l2ContractAddress,
      this.getL2ContractEventLogClient(l2ContractAddress),
      this.mode,
      signer,
      this.maxFeePerGas,
      this.gasFeeEstimationPercentile,
      false,
    );

    return l2MessageServiceContract;
  }

  /**
   * Creates an instance of the `L1ClaimingService` for managing message claiming on L1.
   *
   * @param {string} [localL1ContractAddress] - Optional custom L1 contract address. Required if the network is set to 'custom'.
   * @param {string} [localL2ContractAddress] - Optional custom L2 contract address. Required if the network is set to 'custom'.
   * @returns {L1ClaimingService} An instance of the `L1ClaimingService` configured for the specified contract addresses.
   */
  public getL1ClaimingService(localL1ContractAddress?: string, localL2ContractAddress?: string) {
    return new L1ClaimingService(
      this.getL1Contract(localL1ContractAddress, localL2ContractAddress),
      this.getL2Contract(localL2ContractAddress),
      this.getL2ContractEventLogClient(localL2ContractAddress),
      this.network,
    );
  }
}
