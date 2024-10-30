import { AccountManager } from "./accounts/account-manager";

export type BaseConfig = {
  rpcUrl: URL;
  chainId: number;
  accountManager: AccountManager;
  dummyContractAddress: string;
};

export type L1Config = BaseConfig & {
  lineaRollupAddress: string;
};

export type BaseL2Config = BaseConfig & {
  l2MessageServiceAddress: string;
  l2TestContractAddress: string;
};

export type LocalL2Config = BaseL2Config & {
  besuNodeRpcUrl: URL;
  shomeiEndpoint: URL;
  shomeiFrontendEndpoint: URL;
  sequencerEndpoint: URL;
  transactionExclusionEndpoint: URL;
};

export type DevL2Config = BaseL2Config;

export type SepoliaL2Config = BaseL2Config;

export type L2Config = LocalL2Config | DevL2Config | SepoliaL2Config;

export type Config = {
  L1: L1Config;
  L2: L2Config;
};
