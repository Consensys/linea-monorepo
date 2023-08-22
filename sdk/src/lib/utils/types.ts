import { Event } from "@ethersproject/contracts";
import { BigNumber } from "ethers";

export type Message = {
  messageSender: string;
  destination: string;
  fee: BigNumber;
  value: BigNumber;
  messageNonce: BigNumber;
  calldata: string;
  messageHash: string;
};

export type Network = "linea-mainnet" | "linea-goerli" | "localhost";
export type SDKMode = "read-only" | "read-write";

interface BaseOptions {
  readonly network: Network;
  readonly l1RpcUrl: string;
  readonly l2RpcUrl: string;
  readonly mode: SDKMode;
}

export interface ReadOnlyModeOptions extends BaseOptions {
  readonly mode: "read-only";
}

export interface WriteModeOptions extends BaseOptions {
  readonly mode: "read-write";
  readonly l1SignerPrivateKey: string;
  readonly l2SignerPrivateKey: string;
  readonly feeEstimatorOptions?: FeeEstimatorOptions;
}

export type FeeEstimatorOptions = {
  maxFeePerGas?: number;
  gasFeeEstimationPercentile?: number;
};

export type LineaSDKOptions = WriteModeOptions | ReadOnlyModeOptions;

export type NetworkInfo = {
  [key in Exclude<Network, "localhost">]: {
    l1ContractAddress: string;
    l2ContractAddress: string;
  };
};

export type ParsedEvent<T extends Event> = {
  args: T["args"];
  blockNumber: number;
  logIndex: number;
  contractAddress: string;
  transactionHash: string;
};
