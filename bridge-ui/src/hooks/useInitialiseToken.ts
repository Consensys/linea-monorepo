import { NetworkTokens, TokenInfo, TokenType } from "@/config";
import { Token } from "@/models/token";
import { defaultTokensConfig, useTokenStore } from "@/stores/tokenStore";
import log from "loglevel";
import { useEffect } from "react";

enum NetworkTypes {
  MAINNET = "MAINNET",
  SEPOLIA = "SEPOLIA",
}
const CANONICAL_BRIDGED_TYPE = "canonical-bridge";
const USDC_TYPE = "USDC";

async function getTokens(networkTypes: NetworkTypes): Promise<Token[]> {
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

export async function getConfig(): Promise<NetworkTokens> {
  const mainnetTokens = await getTokens(NetworkTypes.MAINNET);
  const sepoliaTokens = await getTokens(NetworkTypes.SEPOLIA);

  const updatedTokensConfig = { ...defaultTokensConfig };

  updatedTokensConfig.MAINNET = [
    ...defaultTokensConfig.MAINNET,
    ...mainnetTokens.map((token: Token): TokenInfo => {
      const tokenType = token.symbol === USDC_TYPE ? TokenType.USDC : TokenType.ERC20;
      return {
        name: token.name,
        symbol: token.symbol,
        decimals: token.decimals,
        type: tokenType,
        L1: token?.extension?.rootAddress ?? null,
        L2: token.address,
        UNKNOWN: null,
        image: token.logoURI,
        isDefault: true,
      };
    }),
  ];

  updatedTokensConfig.SEPOLIA = [
    ...defaultTokensConfig.SEPOLIA,
    ...sepoliaTokens.map((token: Token): TokenInfo => {
      const tokenType = token.symbol === USDC_TYPE ? TokenType.USDC : TokenType.ERC20;
      return {
        name: token.name,
        symbol: token.symbol,
        decimals: token.decimals,
        type: tokenType,
        L1: token?.extension?.rootAddress ?? null,
        L2: token.address,
        UNKNOWN: null,
        image: token.logoURI,
        isDefault: true,
      };
    }),
  ];

  return updatedTokensConfig;
}

const useInitialiseToken = () => {
  const { setTokensConfig, setDefaultTokenList, usersTokens, defaultTokenList } = useTokenStore((state) => ({
    setTokensConfig: state.setTokensConfig,
    setDefaultTokenList: state.setDefaultTokenList,
    usersTokens: state.usersTokens,
    defaultTokenList: state.defaultTokenList,
  }));

  useEffect(() => {
    const updateDefaultTokenList = async () => {
      // Get the latest default tokens if they have not been loaded yet
      const _tokenList = await getConfig();
      setDefaultTokenList(_tokenList);
    };

    // Update the context every time the users's token storage is updated
    updateDefaultTokenList();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const updateTokenConfig = async () => {
      if (defaultTokenList) {
        const _newTokensConfig: NetworkTokens = {
          MAINNET: [],
          SEPOLIA: [],
          UNKNOWN: [],
        };

        _newTokensConfig.MAINNET = [...defaultTokenList.MAINNET, ...usersTokens.MAINNET];
        _newTokensConfig.SEPOLIA = [...defaultTokenList.SEPOLIA, ...usersTokens.SEPOLIA];

        setTokensConfig(_newTokensConfig);
      }
    };

    // Update the context every time the users's token storage is updated
    updateTokenConfig();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [usersTokens, defaultTokenList]);
};

export default useInitialiseToken;
