import { PrivateKeyAccount } from "viem";
import { Config } from "../../config/config-schema";
import { ClientFactory } from "./client-factory";
import { createL1ReadContractsExtension } from "./extensions/l1-read-contracts";
import { createL1WriteContractsExtension } from "./extensions/l1-write-contracts";

export class L1Client {
  private readonly factory: ClientFactory;

  constructor(private readonly config: Config) {
    this.factory = new ClientFactory(this.config.L1.chainId, this.config.L1.rpcUrl);
  }

  public publicClient(params?: { chainId?: number; rpcUrl?: URL }) {
    return this.factory.getPublic(params).extend(createL1ReadContractsExtension(this.config.L1));
  }

  public walletClient(params?: { chainId?: number; rpcUrl?: URL; account: PrivateKeyAccount | undefined }) {
    return this.factory.getWallet(params).extend(createL1WriteContractsExtension(this.config.L1));
  }
}
