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
  tokensConfig: NetworkTokens;
  defaultTokenList: NetworkTokens;
  usersTokens: NetworkTokens;
};

export type TokenActions = {
  setTokensConfig: (tokensConfig: NetworkTokens) => void;
  setDefaultTokenList: (defaultTokenList: NetworkTokens) => void;
  setUsersTokens: (usersTokens: NetworkTokens) => void;
};

export type TokenStore = TokenState & TokenActions;

export const defaultInitState: TokenState = {
  tokensConfig: defaultTokensConfig,
  defaultTokenList: { MAINNET: [], SEPOLIA: [], UNKNOWN: [] },
  usersTokens: { MAINNET: [], SEPOLIA: [], UNKNOWN: [] },
};

export const useTokenStore = create<TokenStore>()(
  persist(
    (set) => ({
      ...defaultInitState,
      setTokensConfig: (tokensConfig: NetworkTokens) => set({ tokensConfig }),
      setDefaultTokenList: (defaultTokenList: NetworkTokens) => set({ defaultTokenList }),
      setUsersTokens: (usersTokens: NetworkTokens) => set({ usersTokens }),
    }),
    {
      name: "token-storage", // name of the item in the storage (must be unique)
      storage: createJSONStorage(() => localStorage),
      version: parseInt(config.storage.minVersion),
      partialize: (state) => ({ usersTokens: state.usersTokens }),
      migrate: () => {
        return defaultInitState;
      },
    },
  ),
);
