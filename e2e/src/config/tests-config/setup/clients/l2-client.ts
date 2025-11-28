import { ClientFactory } from "./client-factory";

import { PrivateKeyAccount } from "viem";
import { L2Config, LocalL2Config } from "../../config/config-schema";
import { createL2WriteContractsExtension } from "./extensions/l2-write-contracts";
import { createL2ReadContractsExtension } from "./extensions/l2-read-contracts";

export type L2RpcEndpointType =
  | "default"
  | "sequencer"
  | "besuNode"
  | "besuFollower"
  | "shomei"
  | "shomeiFrontend"
  | "transactionExclusion";

export enum L2RpcEndpoint {
  Default = "default",
  Sequencer = "sequencer",
  BesuNode = "besuNode",
  BesuFollower = "besuFollower",
  Shomei = "shomei",
  ShomeiFrontend = "shomeiFrontend",
  TransactionExclusion = "transactionExclusion",
}

export class L2Client {
  private readonly factory: ClientFactory;

  constructor(
    private readonly config: L2Config,
    private readonly local?: LocalL2Config,
  ) {
    this.factory = new ClientFactory(config.chainId, config.rpcUrl);
  }

  public publicClient(params?: { type?: L2RpcEndpointType }) {
    const { type = L2RpcEndpoint.Default } = params ?? {};

    const url = this.resolveRpcUrl(type);
    return this.factory
      .getPublic({ chainId: this.config.chainId, rpcUrl: url })
      .extend(createL2ReadContractsExtension(this.config));
  }

  public walletClient(params?: { type?: L2RpcEndpointType; account: PrivateKeyAccount | undefined }) {
    const { type = L2RpcEndpoint.Default, account } = params ?? {};
    const url = this.resolveRpcUrl(type);
    return this.factory
      .getWallet({ chainId: this.config.chainId, rpcUrl: url, account })
      .extend(createL2WriteContractsExtension(this.config));
  }

  private resolveRpcUrl(type: L2RpcEndpointType): URL {
    if (type === L2RpcEndpoint.Default) {
      return this.config.rpcUrl;
    }

    if (!this.local) {
      throw new Error(`RPC type "${type}" is only available on local L2 setups.`);
    }

    switch (type) {
      case L2RpcEndpoint.Sequencer:
        return this.local.sequencerEndpoint;

      case L2RpcEndpoint.BesuNode:
        return this.local.besuNodeRpcUrl;

      case L2RpcEndpoint.BesuFollower:
        if (!this.local.besuFollowerNodeRpcUrl) {
          throw new Error("besuFollowerNodeRpcUrl not configured.");
        }
        return this.local.besuFollowerNodeRpcUrl;
      case L2RpcEndpoint.Shomei:
        if (!this.local.shomeiEndpoint) {
          throw new Error("shomeiEndpoint not configured.");
        }
        return this.local.shomeiEndpoint;

      case L2RpcEndpoint.ShomeiFrontend:
        if (!this.local.shomeiFrontendEndpoint) {
          throw new Error("shomeiFrontendEndpoint not configured.");
        }
        return this.local.shomeiFrontendEndpoint;

      case L2RpcEndpoint.TransactionExclusion:
        if (!this.local.transactionExclusionEndpoint) {
          throw new Error("transactionExclusionEndpoint not configured.");
        }
        return this.local.transactionExclusionEndpoint;

      default:
        throw new Error(`Unknown L2 RPC endpoint type: ${type}`);
    }
  }
}
