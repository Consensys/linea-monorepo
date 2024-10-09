import { AccountManager } from "./accounts/account-manager";

export type BaseConfig = {
  rpcUrl: URL;
  chainId: number;
  accountManager: AccountManager;
  dummyContractAddress?: string;
};

export type L1Config = BaseConfig & {
  lineaRollupAddress: string;
};

export type L2Config = BaseConfig & {
  l2MessageServiceAddress: string;
  shomeiEndpoint?: URL;
  shomeiFrontendEndpoint?: URL;
  sequencerEndpoint?: URL;
};

export type Config = {
  L1: L1Config;
  L2: L2Config;
};
