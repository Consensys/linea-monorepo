import { create } from "zustand";
import { config, NetworkTokens, TokenType } from "@/config";
import { createJSONStorage, persist } from "zustand/middleware";

export const defaultTokensConfig: NetworkTokens = {
  MAINNET: [
    {
      name: "Ether",
      symbol: "ETH",
      decimals: 18,
      type: TokenType.ETH,
      L1: null,
      L2: null,
      UNKNOWN: null,
      image: "https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png",
      isDefault: true,
    },
  ],
  SEPOLIA: [
    {
      name: "Ether",
      symbol: "ETH",
      decimals: 18,
      type: TokenType.ETH,
      L1: null,
      L2: null,
      UNKNOWN: null,
      image: "https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png",
      isDefault: true,
    },
  ],
  UNKNOWN: [
    {
      name: "Ether",
      symbol: "ETH",
      decimals: 18,
      type: TokenType.ETH,
      L1: null,
      L2: null,
      UNKNOWN: null,
      image: "https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png",
      isDefault: true,
    },
  ],
};

export type TokenState = {
  tokensList: NetworkTokens;
};

export type TokenActions = {
  setTokensList: (tokensConfig: NetworkTokens) => void;
};

export type TokenStore = TokenState & TokenActions;

export const defaultInitState: TokenState = {
  tokensList: defaultTokensConfig,
};

export const useTokenStore = create<TokenStore>()(
  persist(
    (set) => ({
      ...defaultInitState,
      setTokensList: (tokensList: NetworkTokens) => set({ tokensList }),
    }),
    {
      name: "token-storage", // name of the item in the storage (must be unique)
      storage: createJSONStorage(() => localStorage),
      version: parseInt(config.storage.minVersion),
      migrate: () => {
        return defaultInitState;
      },
    },
  ),
);
