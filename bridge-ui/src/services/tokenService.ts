import axios, { AxiosResponse } from "axios";
import log from "loglevel";
import { Address } from "viem";
import { GetTokenReturnType, getToken } from "@wagmi/core";
import { sepolia, linea, mainnet, lineaSepolia, Chain } from "viem/chains";
import { NetworkType, TokenInfo, TokenType, wagmiConfig } from "@/config";
import { Token } from "@/models/token";

interface CoinGeckoToken {
  id: string;
  symbol: string;
  name: string;
}

interface CoinGeckoTokenDetail {
  image: {
    small: string;
  };
}

enum NetworkTypes {
  MAINNET = "MAINNET",
  SEPOLIA = "SEPOLIA",
}

export const CANONICAL_BRIDGED_TYPE = "canonical-bridge";
export const USDC_TYPE = "USDC";

export async function fetchERC20Image(name: string) {
  try {
    if (!name) {
      throw new Error("Name is required");
    }

    const coinsResponse: AxiosResponse<CoinGeckoToken[]> = await axios.get(
      "https://api.coingecko.com/api/v3/coins/list",
    );
    const coin = coinsResponse.data.find((coin: CoinGeckoToken) => coin.name === name);

    if (!coin) {
      throw new Error("Coin not found");
    }

    const coinId = coin.id;
    const coinDataResponse: AxiosResponse<CoinGeckoTokenDetail> = await axios.get(
      `https://api.coingecko.com/api/v3/coins/${coinId}`,
    );

    if (!coinDataResponse.data.image.small) {
      throw new Error("Image not found");
    }

    const imageUrl = coinDataResponse.data.image.small.split("?")[0];
    // Test image URL
    const response = await axios.get(imageUrl, { timeout: 5000 });

    if (response.status !== 200) {
      return "/images/logo/noTokenLogo.svg";
    }

    return imageUrl;
  } catch (error) {
    log.warn(error);
    return "/images/logo/noTokenLogo.svg";
  }
}

export async function fetchTokenInfo(
  tokenAddress: Address,
  networkType: NetworkType,
  fromChain?: Chain,
): Promise<TokenInfo | undefined> {
  let erc20: GetTokenReturnType | undefined;
  let chainFound;

  if (!chainFound) {
    const chains: Chain[] = networkType === NetworkType.SEPOLIA ? [lineaSepolia, sepolia] : [linea, mainnet];

    // Put the fromChain arg at the begining to take it as priority
    if (fromChain) chains.unshift(fromChain);

    for (const chain of chains) {
      try {
        erc20 = await getToken(wagmiConfig, {
          address: tokenAddress,
          chainId: chain.id,
        });
        if (erc20.name) {
          // Found the token if no errors with fetchToken
          chainFound = chain;
          break;
        }
      } catch (err) {
        continue;
      }
    }
  }

  if (!erc20 || !chainFound || !erc20.name) {
    return;
  }

  const L1Token = chainFound.id === mainnet.id || chainFound.id === sepolia.id;

  // Fetch image
  const name = erc20.name;
  const image = await fetchERC20Image(name);

  try {
    return {
      name,
      symbol: erc20.symbol!,
      decimals: erc20.decimals,
      L1: L1Token ? tokenAddress : null,
      L2: !L1Token ? tokenAddress : null,
      image,
      type: TokenType.ERC20,
      UNKNOWN: null,
      isDefault: false,
    };
  } catch (err) {
    log.error(err);
    return;
  }
}

export async function getTokens(networkTypes: NetworkTypes): Promise<Token[]> {
  try {
    // Fetch the JSON data from the URL.
    let url = process.env.NEXT_PUBLIC_MAINNET_TOKEN_LIST ? (process.env.NEXT_PUBLIC_MAINNET_TOKEN_LIST as string) : "";
    if (networkTypes === NetworkTypes.SEPOLIA) {
      url = process.env.NEXT_PUBLIC_SEPOLIA_TOKEN_LIST ? (process.env.NEXT_PUBLIC_SEPOLIA_TOKEN_LIST as string) : "";
    }

    const response = await fetch(url);
    const data = await response.json();
    const tokens = data.tokens;
    const bridgedTokens = tokens.filter(
      (token: Token) => token.tokenType.includes(CANONICAL_BRIDGED_TYPE) || token.symbol === USDC_TYPE,
    );
    return bridgedTokens;
  } catch (error) {
    log.error("Error getTokens", { error });
    return [];
  }
}

export async function fetchTokenPrices(
  tokenAddresses: Address[],
  chainId?: number,
): Promise<Record<string, { usd: number }>> {
  if (!chainId) {
    return {};
  }

  const response = await fetch(
    `https://price.api.cx.metamask.io/v2/chains/${chainId}/spot-prices?tokenAddresses=${tokenAddresses.join(",")}&vsCurrency=usd`,
  );
  if (!response.ok) {
    throw new Error("Error in getTokenPrices");
  }

  const data = await response.json();
  return data;
}
