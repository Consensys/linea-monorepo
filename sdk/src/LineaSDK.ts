import { EthersLineaRollupLogClient } from "./clients/ethereum/EthersLineaRollupLogClient";
import { LineaRollupClient } from "./clients/ethereum/LineaRollupClient";
import { EthersL2MessageServiceLogClient } from "./clients/linea/EthersL2MessageServiceLogClient";
import { L2MessageServiceClient } from "./clients/linea/L2MessageServiceClient";
import {
  DEFAULT_GAS_ESTIMATION_PERCENTILE,
  DEFAULT_L2_MESSAGE_TREE_DEPTH,
  DEFAULT_MAX_FEE_PER_GAS,
} from "./core/constants";
import { BaseError } from "./core/errors/Base";
import { L1ClaimingService } from "./clients/ethereum/L1ClaimingService";
import { LineaSDKOptions, Network, SDKMode } from "./core/types/config";
import { NETWORKS } from "./core/constants/networks";
import { DefaultGasProvider } from "./clients/gas/DefaultGasProvider";
import { LineaGasProvider } from "./clients/gas/LineaGasProvider";
import { LineaRollupMessageRetriever } from "./clients/ethereum/LineaRollupMessageRetriever";
import { MerkleTreeService } from "./clients/ethereum/MerkleTreeService";
import { L2MessageServiceMessageRetriever } from "./clients/linea/L2MessageServiceMessageRetriever";
import { Provider, LineaProvider, BrowserProvider, LineaBrowserProvider } from "./clients/providers";
import { isString } from "./core/utils/shared";
import { Wallet } from "./clients/wallet";

export class LineaSDK {
  private network: Network;
  private l1SignerPrivateKeyOrWallet: string | Wallet<Provider | BrowserProvider>;
  private l2SignerPrivateKeyOrWallet: string | Wallet<LineaProvider | LineaBrowserProvider>;
  private l1Provider: Provider | BrowserProvider;
  private l2Provider: LineaProvider | LineaBrowserProvider;
  private maxFeePerGas: bigint;
  private gasFeeEstimationPercentile: number;
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
      this.maxFeePerGas = options.feeEstimatorOptions?.maxFeePerGas ?? DEFAULT_MAX_FEE_PER_GAS;
      this.gasFeeEstimationPercentile =
        options.feeEstimatorOptions?.gasFeeEstimationPercentile ?? DEFAULT_GAS_ESTIMATION_PERCENTILE;
    } else {
      this.maxFeePerGas = DEFAULT_MAX_FEE_PER_GAS;
      this.gasFeeEstimationPercentile = DEFAULT_GAS_ESTIMATION_PERCENTILE;
    }

    if (isString(options.l1RpcUrlOrProvider)) {
      this.l1Provider = new Provider(options.l1RpcUrlOrProvider);
      console.log(JSON.stringify(this.l1Provider));
    } else {
      this.l1Provider = new BrowserProvider(options.l1RpcUrlOrProvider);
    }

    if (isString(options.l2RpcUrlOrProvider)) {
      this.l2Provider = new LineaProvider(options.l2RpcUrlOrProvider);
    } else {
      this.l2Provider = new LineaBrowserProvider(options.l2RpcUrlOrProvider);
    }
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

    return new EthersL2MessageServiceLogClient(this.l2Provider, l2ContractAddress);
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

    const signer =
      this.mode === "read-write"
        ? Wallet.getWallet(this.l1SignerPrivateKeyOrWallet).connect(this.l1Provider)
        : undefined;

    const lineaRollupLogClient = new EthersLineaRollupLogClient(this.l1Provider, l1ContractAddress);
    const l2MessageServiceLogClient = this.getL2ContractEventLogClient(l2ContractAddress);

    return new LineaRollupClient(
      this.l1Provider,
      l1ContractAddress,
      lineaRollupLogClient,
      l2MessageServiceLogClient,
      new DefaultGasProvider(this.l1Provider, {
        maxFeePerGas: this.maxFeePerGas,
        gasEstimationPercentile: this.gasFeeEstimationPercentile,
        enforceMaxGasFee: false,
      }),
      new LineaRollupMessageRetriever(this.l1Provider, lineaRollupLogClient, l1ContractAddress),
      new MerkleTreeService(
        this.l1Provider,
        l1ContractAddress,
        lineaRollupLogClient,
        l2MessageServiceLogClient,
        this.l2MessageTreeDepth,
      ),
      this.mode,
      signer,
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

    const signer =
      this.mode === "read-write"
        ? Wallet.getWallet(this.l2SignerPrivateKeyOrWallet).connect(this.l2Provider)
        : undefined;

    const l2MessageServiceContract = new L2MessageServiceClient(
      this.l2Provider,
      l2ContractAddress,
      new L2MessageServiceMessageRetriever(
        this.l2Provider,
        this.getL2ContractEventLogClient(l2ContractAddress),
        l2ContractAddress,
      ),
      new LineaGasProvider(this.l2Provider, {
        maxFeePerGas: this.maxFeePerGas,
        enforceMaxGasFee: false,
      }),
      this.mode,
      signer,
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
  public getL1ClaimingService(localL1ContractAddress?: string, localL2ContractAddress?: string): L1ClaimingService {
    return new L1ClaimingService(
      this.getL1Contract(localL1ContractAddress, localL2ContractAddress),
      this.getL2Contract(localL2ContractAddress),
      this.getL2ContractEventLogClient(localL2ContractAddress),
      this.network,
    );
  }
}
