import { ILogger } from "@consensys/linea-shared-utils";
import { createPublicClient, createWalletClient, defineChain, http, PublicClient } from "viem";

import { ChainContext } from "./ChainContext";
import { contractSignerToViemAccount } from "./contractSignerToViemAccount";
import { createSignerClient } from "./createSignerClient";
import { SignerConfig } from "./SignerConfig";

async function getChainId(client: PublicClient, logger: ILogger): Promise<number> {
  const deadlineMs = 60_000;
  const baseDelayMs = 1_000;
  const startTime = Date.now();
  let attempt = 0;
  let elapsed = 0;

  do {
    try {
      return await client.getChainId();
    } catch (error) {
      attempt++;
      elapsed = Date.now() - startTime;
      if (elapsed >= deadlineMs) {
        logger.error(`Failed to fetch chainId after ${elapsed}ms (${attempt} attempts)`, { error });
        throw error;
      }
      const delay = Math.min(baseDelayMs * 2 ** (attempt - 1), deadlineMs - elapsed);
      logger.warn(`Failed to fetch chainId, retrying in ${delay}ms (attempt ${attempt})`, { error });
      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  } while (elapsed < deadlineMs);

  throw new Error(`Failed to fetch chainId after ${deadlineMs}ms`);
}

export async function createChainContext(
  rpcUrl: string,
  signerConfig: SignerConfig,
  logger: ILogger,
): Promise<ChainContext> {
  const tempClient = createPublicClient({ transport: http(rpcUrl) });
  const chainId = await getChainId(tempClient, logger);

  const chain = defineChain({
    id: chainId,
    name: "custom",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: [rpcUrl] } },
  });

  const publicClient = createPublicClient({ chain, transport: http(rpcUrl) });

  const signer = createSignerClient(signerConfig, logger, rpcUrl, chain);
  const account = contractSignerToViemAccount(signer);

  const walletClient = createWalletClient({
    account,
    transport: http(rpcUrl),
    chain,
  });

  return { chainId, chain, publicClient, walletClient, account, signer };
}
