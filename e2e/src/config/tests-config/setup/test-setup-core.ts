import { AccountManager } from "../accounts/account-manager";
import { Config } from "../config/config-schema";
import { L1Client } from "./clients/l1-client";
import { L2Client, L2RpcEndpointType } from "./clients/l2-client";
import { PublicClientParams, WalletClientParams } from "./clients/client-factory";

export default abstract class TestSetupCore {
  public L1: {
    client: L1Client;
  };

  public L2: {
    client: L2Client;
  };

  constructor(protected readonly config: Config) {
    const l1Client = new L1Client(config);
    const l2Client = new L2Client(config.L2, undefined);

    this.L1 = {
      client: l1Client,
    };

    this.L2 = {
      client: l2Client,
    };
  }

  public getL1AccountManager(): AccountManager {
    return this.config.L1.accountManager;
  }

  public getL2AccountManager(): AccountManager {
    return this.config.L2.accountManager;
  }

  public getL2ChainId(): number {
    return this.config.L2.chainId;
  }

  public l1PublicClient(params?: PublicClientParams) {
    return this.L1.client.publicClient(params);
  }

  public l1WalletClient(params?: WalletClientParams) {
    return this.L1.client.walletClient(params);
  }

  public l2PublicClient(params?: { type?: L2RpcEndpointType; httpConfig?: PublicClientParams["httpConfig"] }) {
    return this.L2.client.publicClient(params);
  }

  public l2WalletClient(params?: {
    type?: L2RpcEndpointType;
    account: WalletClientParams["account"];
    httpConfig?: WalletClientParams["httpConfig"];
  }) {
    return this.L2.client.walletClient(params);
  }

  abstract isLocal(): boolean;
}
