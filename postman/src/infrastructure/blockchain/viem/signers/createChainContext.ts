import { ILogger } from "@consensys/linea-shared-utils";
import { createPublicClient, createWalletClient, defineChain, http } from "viem";

import { ChainContext } from "./ChainContext";
import { contractSignerToViemAccount } from "./contractSignerToViemAccount";
import { createSignerClient } from "./createSignerClient";
import { SignerConfig } from "./SignerConfig";

export async function createChainContext(
  rpcUrl: string,
  signerConfig: SignerConfig,
  logger: ILogger,
): Promise<ChainContext> {
  const publicClient = createPublicClient({ transport: http(rpcUrl) });
  const chainId = await publicClient.getChainId();

  const chain = defineChain({
    id: chainId,
    name: "custom",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: [rpcUrl] } },
  });

  const signer = createSignerClient(signerConfig, logger, rpcUrl, chain);
  const account = contractSignerToViemAccount(signer);

  const walletClient = createWalletClient({
    account,
    transport: http(rpcUrl),
    chain,
  });

  return { chainId, chain, publicClient, walletClient, account, signer };
}
