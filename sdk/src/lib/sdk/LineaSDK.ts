import { L1MessageServiceContract, L2MessageServiceContract } from "../contracts";
import { NETWORKS } from "../utils/networks";
import { LineaSDKOptions, SDKMode, Network } from "../utils/types";
import ProviderService from "./ProviderService";

export class LineaSDK {
  private network: Network;
  private l1SignerPrivateKey: string;
  private l2SignerPrivateKey: string;
  private l1Provider: ProviderService;
  private l2Provider: ProviderService;
  private maxFeePerGas?: number;
  private gasFeeEstimationPercentile?: number;
  private mode: SDKMode;

  constructor(options: LineaSDKOptions) {
    this.network = options.network;
    this.mode = options.mode;

    if (options.mode === "read-write") {
      this.l1SignerPrivateKey = options.l1SignerPrivateKey;
      this.l2SignerPrivateKey = options.l2SignerPrivateKey;
      this.maxFeePerGas = options.feeEstimatorOptions?.maxFeePerGas;
      this.gasFeeEstimationPercentile = options.feeEstimatorOptions?.gasFeeEstimationPercentile;
    }

    this.l1Provider = new ProviderService(options.l1RpcUrl);
    this.l2Provider = new ProviderService(options.l2RpcUrl);
  }

  public getL1Contract(localContractAddress?: string) {
    let l1ContractAddress: string;

    if (this.network === "localhost") {
      if (!localContractAddress) {
        throw new Error("You need to provide a contract address.");
      }
      l1ContractAddress = localContractAddress;
    } else {
      l1ContractAddress = NETWORKS[this.network].l1ContractAddress;
    }

    const signer = this.mode === "read-write" ? this.l1Provider.getSigner(this.l1SignerPrivateKey) : undefined;

    return new L1MessageServiceContract(
      this.l1Provider.provider,
      l1ContractAddress,
      this.mode,
      signer,
      this.maxFeePerGas,
      this.gasFeeEstimationPercentile,
    );
  }

  public getL2Contract(localContractAddress?: string) {
    let l2ContractAddress: string;

    if (this.network === "localhost") {
      if (!localContractAddress) {
        throw new Error("You need to provide a contract address.");
      }
      l2ContractAddress = localContractAddress;
    } else {
      l2ContractAddress = NETWORKS[this.network].l2ContractAddress;
    }

    const signer = this.mode === "read-write" ? this.l2Provider.getSigner(this.l2SignerPrivateKey) : undefined;

    return new L2MessageServiceContract(
      this.l2Provider.provider,
      l2ContractAddress,
      this.mode,
      signer,
      this.maxFeePerGas,
      this.gasFeeEstimationPercentile,
    );
  }
}
