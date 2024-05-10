import { createContext, useEffect, useState, useContext } from 'react';
import { useLocalStorage } from 'usehooks-ts';

import { TokenType, NetworkTokens, TokenInfo } from '@/config/config';
import { StorageKeys } from '@/contexts/storage';
import { Token } from '@/models/token';
import log from 'loglevel';

const CANONICAL_BRIDGED_TYPE = 'canonical-bridge';
const USDC_TYPE = 'USDC';

type Props = {
  children: JSX.Element;
};

interface ConfigContextData {
  tokensConfig: NetworkTokens;
}

enum NetworkTypes {
  MAINNET = 'MAINNET',
  SEPOLIA = 'SEPOLIA',
}

export const ConfigContext = createContext<ConfigContextData>({} as ConfigContextData);

async function getTokens(networkTypes: NetworkTypes): Promise<Token[]> {
  try {
    // Fetch the JSON data from the URL.
    let url = process.env.NEXT_PUBLIC_MAINNET_TOKEN_LIST ? (process.env.NEXT_PUBLIC_MAINNET_TOKEN_LIST as string) : '';
    if (networkTypes === NetworkTypes.SEPOLIA) {
      url = process.env.NEXT_PUBLIC_SEPOLIA_TOKEN_LIST ? (process.env.NEXT_PUBLIC_SEPOLIA_TOKEN_LIST as string) : '';
    }

    const response = await fetch(url);
    const data = await response.json();
    const tokens = data.tokens;
    const bridgedTokens = tokens.filter(
      (token: Token) => token.tokenType.includes(CANONICAL_BRIDGED_TYPE) || token.symbol === USDC_TYPE,
    );
    return bridgedTokens;
  } catch (error) {
    log.error('Error getTokens', { error });
    return [];
  }
}

export const defaultTokensConfig: NetworkTokens = {
  MAINNET: [
    {
      name: 'Ether',
      symbol: 'ETH',
      decimals: 18,
      type: TokenType.ETH,
      L1: null,
      L2: null,
      UNKNOWN: null,
      image: 'https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png',
      isDefault: true,
    },
  ],
  SEPOLIA: [
    {
      name: 'Ether',
      symbol: 'ETH',
      decimals: 18,
      type: TokenType.ETH,
      L1: null,
      L2: null,
      UNKNOWN: null,
      image: 'https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png',
      isDefault: true,
    },
  ],
  UNKNOWN: [
    {
      name: 'Ether',
      symbol: 'ETH',
      decimals: 18,
      type: TokenType.ETH,
      L1: null,
      L2: null,
      UNKNOWN: null,
      image: 'https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png',
      isDefault: true,
    },
  ],
};

export const INITIAL_TOKEN_STORAGE: NetworkTokens = {
  MAINNET: [],
  SEPOLIA: [],
  UNKNOWN: [],
};

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

export const useConfigContext = () => {
  return useContext(ConfigContext);
};

export const TokenProvider = ({ children }: Props) => {
  // State used for the context, contains all the tokens (default + customs)
  const [tokensConfig, setTokensConfig] = useState<NetworkTokens>(defaultTokensConfig);

  // Token list from official Linea's token list
  const [defaultTokenList, setDefaultTokenList] = useState<NetworkTokens | undefined>(undefined);

  // Hooks
  // Only store the custom tokens in the user's storage, the default tokens
  // come from the official Linea's token list
  const [storedTokens] = useLocalStorage(StorageKeys.USER_TOKENS, INITIAL_TOKEN_STORAGE);

  useEffect(() => {
    const updateDefaultTokenList = async () => {
      // Get the latest default tokens if they have not been loaded yet
      const _tokenList = await getConfig();
      setDefaultTokenList(_tokenList);
    };

    // Update the context every time the users's token storage is updated
    updateDefaultTokenList();
  }, []);

  useEffect(() => {
    const updateTokenConfig = async () => {
      if (defaultTokenList) {
        const _newTokensConfig: NetworkTokens = {
          MAINNET: [],
          SEPOLIA: [],
          UNKNOWN: [],
        };

        _newTokensConfig.MAINNET = [...defaultTokenList.MAINNET, ...storedTokens.MAINNET];
        _newTokensConfig.SEPOLIA = [...defaultTokenList.SEPOLIA, ...storedTokens.SEPOLIA];

        setTokensConfig(_newTokensConfig);
      }
    };

    // Update the context every time the users's token storage is updated
    updateTokenConfig();
  }, [storedTokens, defaultTokenList]);

  return (
    <ConfigContext.Provider
      value={{
        tokensConfig,
      }}
    >
      {children}
    </ConfigContext.Provider>
  );
};
