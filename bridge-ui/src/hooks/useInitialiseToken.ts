import { useEffect } from "react";
import { NetworkTokens, TokenInfo, TokenType } from "@/config";
import { Token } from "@/models/token";
import { getTokens, USDC_TYPE } from "@/services";
import { defaultTokensConfig, useTokenStore } from "@/stores/tokenStore";

enum NetworkTypes {
  MAINNET = "MAINNET",
  SEPOLIA = "SEPOLIA",
}

export async function getConfig(): Promise<NetworkTokens> {
  const mainnetTokens = await getTokens(NetworkTypes.MAINNET);
  const sepoliaTokens = await getTokens(NetworkTypes.SEPOLIA);

  const updatedTokensConfig = { ...defaultTokensConfig };

  updatedTokensConfig.MAINNET = [
    ...defaultTokensConfig.MAINNET,
    ...(await Promise.all(
      mainnetTokens.map(async (token: Token): Promise<TokenInfo> => {
        const tokenType = token.symbol === USDC_TYPE ? TokenType.USDC : TokenType.ERC20;
        try {
          await fetch(token.logoURI);
        } catch (error) {
          token.logoURI = "/images/logo/noTokenLogo.svg";
        }

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
    )),
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
