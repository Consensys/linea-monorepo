import {
  LineaRollupClient,
  EthersLineaRollupLogClient,
  L1ClaimingService,
  LineaRollupMessageRetriever,
  MerkleTreeService,
} from "./clients/ethereum";
import {
  L2MessageServiceClient,
  EthersL2MessageServiceLogClient,
  L2MessageServiceMessageRetriever,
} from "./clients/linea";
import { DefaultGasProvider, GasProvider } from "./clients/gas";
import { Provider, LineaProvider, BrowserProvider, LineaBrowserProvider } from "./clients/providers";
import { Wallet } from "./clients/wallet";
import {
  DEFAULT_ENABLE_LINEA_ESTIMATE_GAS,
  DEFAULT_ENFORCE_MAX_GAS_FEE,
  DEFAULT_GAS_ESTIMATION_PERCENTILE,
  DEFAULT_L2_MESSAGE_TREE_DEPTH,
  DEFAULT_MAX_FEE_PER_GAS,
} from "./core/constants";
import { BaseError } from "./core/errors";
import { LineaSDKOptions, Network, SDKMode } from "./core/types";
import { NETWORKS } from "./core/constants";
import { isString } from "./core/utils";
import { Direction } from "./core/enums";

export class LineaSDK {
  private network: Network;
  private l1SignerPrivateKeyOrWallet: string | Wallet<Provider | BrowserProvider>;
  private l2SignerPrivateKeyOrWallet: string | Wallet<LineaProvider | LineaBrowserProvider>;
  private l1Provider: Provider | BrowserProvider;
  private l2Provider: LineaProvider | LineaBrowserProvider;
  private maxFeePerGas: bigint;
  private gasFeeEstimationPercentile: number;
  private enabledLineaEstimateGas: boolean;
  private enforceMaxGasFee: boolean;
  private mode: SDKMode;
  private l2MessageTreeDepth: number;

  /**
   * Initializes a new instance of the LineaSDK with the specified options.
   *
   * @param {LineaSDKOptions} options - Configuration options for the SDK, including network details, operational mode, and optional settings for L2 message tree depth and fee estimation.
   */
  constructor(options: LineaSDKOptions) {
    const {
      network,
      mode,
      l2MessageTreeDepth = DEFAULT_L2_MESSAGE_TREE_DEPTH,
      l1RpcUrlOrProvider,
      l2RpcUrlOrProvider,
      feeEstimatorOptions = {},
    } = options;

    this.network = network;
    this.mode = mode;
    this.l2MessageTreeDepth = l2MessageTreeDepth;

    if (mode === "read-write") {
      const { l1SignerPrivateKeyOrWallet, l2SignerPrivateKeyOrWallet } = options;

      if (!l1SignerPrivateKeyOrWallet || !l2SignerPrivateKeyOrWallet) {
        throw new BaseError("You need to provide both L1 and L2 signer private keys or wallets.");
      }

      this.l1SignerPrivateKeyOrWallet = l1SignerPrivateKeyOrWallet;
      this.l2SignerPrivateKeyOrWallet = l2SignerPrivateKeyOrWallet;
    }

    const {
      maxFeePerGas = DEFAULT_MAX_FEE_PER_GAS,
      gasFeeEstimationPercentile = DEFAULT_GAS_ESTIMATION_PERCENTILE,
      enableLineaEstimateGas = DEFAULT_ENABLE_LINEA_ESTIMATE_GAS,
      enforceMaxGasFee = DEFAULT_ENFORCE_MAX_GAS_FEE,
    } = feeEstimatorOptions;

    this.maxFeePerGas = maxFeePerGas;
    this.gasFeeEstimationPercentile = gasFeeEstimationPercentile;
    this.enabledLineaEstimateGas = enableLineaEstimateGas;
    this.enforceMaxGasFee = enforceMaxGasFee;

    this.l1Provider = isString(l1RpcUrlOrProvider)
      ? new Provider(l1RpcUrlOrProvider)
      : new BrowserProvider(l1RpcUrlOrProvider);

    this.l2Provider = isString(l2RpcUrlOrProvider)
      ? new LineaProvider(l2RpcUrlOrProvider)
      : new LineaBrowserProvider(l2RpcUrlOrProvider);
  }

  /**
   * Creates an instance of the `EthersL2MessageServiceLogClient` for interacting with L2 contract event logs.
   *
   * @param {string} [localL2ContractAddress] - Optional custom L2 contract address. Required if the network is set to 'custom'.
   * @returns {EthersL2MessageServiceLogClient} An instance of the L2 message service log client.
   */
  public getL2ContractEventLogClient(localL2ContractAddress?: string): EthersL2MessageServiceLogClient {
    const l2ContractAddress = this.getContractAddress("l2", localL2ContractAddress);
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
    const l1ContractAddress = this.getContractAddress("l1", localL1ContractAddress);
    const l2ContractAddress = this.getContractAddress("l2", localL2ContractAddress);

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
        enforceMaxGasFee: this.enforceMaxGasFee,
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
    const l2ContractAddress = this.getContractAddress("l2", localContractAddress);

    const signer =
      this.mode === "read-write"
        ? Wallet.getWallet(this.l2SignerPrivateKeyOrWallet).connect(this.l2Provider)
        : undefined;

    const gasProvider = new GasProvider(this.l2Provider, {
      maxFeePerGas: this.maxFeePerGas,
      enforceMaxGasFee: this.enforceMaxGasFee,
      gasEstimationPercentile: this.gasFeeEstimationPercentile,
      direction: Direction.L1_TO_L2,
      enableLineaEstimateGas: this.enabledLineaEstimateGas,
    });

    const l2MessageServiceContract = new L2MessageServiceClient(
      this.l2Provider,
      l2ContractAddress,
      new L2MessageServiceMessageRetriever(
        this.l2Provider,
        this.getL2ContractEventLogClient(l2ContractAddress),
        l2ContractAddress,
      ),
      gasProvider,
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

  /**
   * Retrieves the contract address for the specified contract type.
   * @param contractType The type of contract to retrieve the address for.
   * @param localContractAddress Optional custom contract address.
   * @returns The contract address for the specified contract type.
   */
  private getContractAddress(contractType: "l1" | "l2", localContractAddress?: string): string {
    if (this.network === "custom") {
      if (!localContractAddress) {
        throw new BaseError(`You need to provide a ${contractType.toUpperCase()} contract address.`);
      }
      return localContractAddress;
    } else {
      const contractAddress = NETWORKS[this.network][`${contractType}ContractAddress`];
      if (!contractAddress) {
        throw new BaseError(`Contract address for ${contractType.toUpperCase()} not found in network ${this.network}.`);
      }
      return contractAddress;
    }
  }
}
