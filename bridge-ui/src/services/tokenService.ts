import log from "loglevel";
import { Address } from "viem";
import { NetworkTokens, TokenInfo } from "@/config";
import { Token } from "@/models/token";
import { defaultTokensConfig } from "@/stores/tokenStore";
import { SupportedCurrencies } from "@/stores/configStore";
import { BridgeProvider } from "@/config/config";

enum NetworkTypes {
  MAINNET = "MAINNET",
  SEPOLIA = "SEPOLIA",
}

export async function getTokens(networkTypes: NetworkTypes): Promise<Token[]> {
  try {
    // Fetch the JSON data from the URL.
    let url = process.env.MAINNET_TOKEN_LIST ? (process.env.MAINNET_TOKEN_LIST as string) : "";
    if (networkTypes === NetworkTypes.SEPOLIA) {
      url = process.env.SEPOLIA_TOKEN_LIST ? (process.env.SEPOLIA_TOKEN_LIST as string) : "";
    }

    const response = await fetch(url);
    const data = await response.json();
    const tokens = data.tokens as Token[];
    const bridgedTokens = tokens.filter(
      (token: Token) =>
        token.tokenType.includes("canonical-bridge") || token.tokenType.includes("native") || token.symbol === "USDC",
    );
    return bridgedTokens;
  } catch (error) {
    log.error("Error getTokens", { error });
    return [];
  }
}

export async function fetchTokenPrices(
  tokenAddresses: Address[],
  currency: SupportedCurrencies,
  chainId?: number,
): Promise<Record<string, number>> {
  if (!chainId) {
    return {};
  }

  const response = await fetch(
    `https://price.api.cx.metamask.io/v2/chains/${chainId}/spot-prices?tokenAddresses=${tokenAddresses.join(",")}&vsCurrency=${currency}`,
  );
  if (!response.ok) {
    throw new Error("Error in getTokenPrices");
  }

  const data: Record<string, Record<SupportedCurrencies, number>> = await response.json();

  return Object.fromEntries(Object.entries(data).map(([address, value]) => [address, value[currency]]));
}

export async function validateTokenURI(url: string): Promise<string> {
  try {
    await fetch(url);
    return url;
  } catch (error) {
    return "/images/logo/noTokenLogo.svg";
  }
}

export async function formatToken(token: Token): Promise<TokenInfo> {
  const bridgeProvider = token.symbol === "USDC" ? BridgeProvider.CCTP : BridgeProvider.NATIVE;

  const logoURI = await validateTokenURI(token.logoURI);

  return {
    type: token.tokenType,
    name: token.name,
    symbol: token.symbol,
    decimals: token.decimals,
    L1: token?.extension?.rootAddress ?? null,
    L2: token.address,
    image: logoURI,
    isDefault: true,
    bridgeProvider,
  };
}

export async function getTokenConfig(): Promise<NetworkTokens> {
  const [mainnetTokens, sepoliaTokens] = await Promise.all([
    getTokens(NetworkTypes.MAINNET),
    getTokens(NetworkTypes.SEPOLIA),
  ]);

  const updatedTokensConfig = { ...defaultTokensConfig };

  updatedTokensConfig.MAINNET = [
    ...defaultTokensConfig.MAINNET,
    ...(await Promise.all(mainnetTokens.map(async (token: Token): Promise<TokenInfo> => formatToken(token)))),
  ];

  updatedTokensConfig.SEPOLIA = [
    ...defaultTokensConfig.SEPOLIA,
    ...(await Promise.all(sepoliaTokens.map((token: Token): Promise<TokenInfo> => formatToken(token)))),
  ];

  return updatedTokensConfig;
}
