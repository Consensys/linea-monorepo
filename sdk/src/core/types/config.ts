import { Eip1193Provider, Wallet } from "ethers";

/**
 * Represents the supported Linea blockchain networks or a custom network configuration.
 */
export type Network = "linea-mainnet" | "linea-sepolia" | "custom";

/**
 * Defines the operational mode of the SDK, either `read-only` or `read-write`.
 */
export type SDKMode = "read-only" | "read-write";

/**
 * Base configuration options common to both `read-only` and `read-write` modes.
 */
interface BaseOptions {
  readonly network: Network;
  readonly l1RpcUrlOrProvider: string | Eip1193Provider;
  readonly l2RpcUrlOrProvider: string | Eip1193Provider;
  readonly mode: SDKMode;
  readonly l2MessageTreeDepth?: number;
  readonly feeEstimatorOptions?: FeeEstimatorOptions;
}

/**
 * Configuration options for initializing the SDK in `read-only` mode.
 */
export interface ReadOnlyModeOptions extends BaseOptions {
  readonly mode: "read-only";
}

/**
 * Configuration options for initializing the SDK in `read-write` mode.
 */
export interface WriteModeOptions extends BaseOptions {
  readonly mode: "read-write";
  readonly l1SignerPrivateKeyOrWallet: string | Wallet;
  readonly l2SignerPrivateKeyOrWallet: string | Wallet;
}

/**
 * Options for configuring gas fee estimation in `read-write` mode.
 */
export type FeeEstimatorOptions = {
  maxFeePerGas?: bigint;
  gasFeeEstimationPercentile?: number;
  enforceMaxGasFee?: boolean;
  enableLineaEstimateGas?: boolean;
};

/**
 * Union type representing the possible SDK configuration options, either for `read-only` or `read-write` mode.
 */
export type LineaSDKOptions = WriteModeOptions | ReadOnlyModeOptions;

/**
 * Defines the contract addresses for the supported Linea networks, excluding custom configurations.
 */
export type NetworkInfo = {
  [key in Exclude<Network, "custom">]: {
    l1ContractAddress: string;
    l2ContractAddress: string;
  };
};
