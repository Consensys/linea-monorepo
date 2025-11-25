import { createPublicClient, createWalletClient, http, PrivateKeyAccount } from "viem";
import { resolveChain } from "../chains/chain-registry";

export class ClientFactory {
  constructor(
    private readonly chainId: number,
    private readonly rpcUrl: URL,
  ) {}

  getPublic(params?: { chainId?: number; rpcUrl?: URL }) {
    const { chainId = this.chainId, rpcUrl = this.rpcUrl } = params ?? {};
    return createPublicClient({
      chain: resolveChain(chainId),
      transport: http(rpcUrl.toString()),
    });
  }

  getWallet(params?: { chainId?: number; rpcUrl?: URL; account: PrivateKeyAccount | undefined }) {
    const { chainId = this.chainId, rpcUrl = this.rpcUrl, account } = params ?? {};
    return createWalletClient({
      chain: resolveChain(chainId),
      transport: http(rpcUrl.toString()),
      ...(account ? { account } : {}),
    });
  }
}
