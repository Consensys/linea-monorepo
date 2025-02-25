import { Address, Chain } from "viem";
import { create } from "zustand";
import { createSelectorHooks, ZustandHookSelectors } from "auto-zustand-selectors-hook";
import { config, NetworkLayer, NetworkType, TokenInfo, TokenType } from "@/config";
import { defaultTokensConfig } from "./tokenStore";
import { linea, lineaSepolia, mainnet, sepolia } from "viem/chains";

export type ChainState = {
  chains: Chain[];
  networkType: NetworkType;
  networkLayer: NetworkLayer;
  messageServiceAddress: Address | null;
  tokenBridgeAddress: Address | null;
  l1Chain: Chain | undefined;
  l2Chain: Chain | undefined;
  fromChain: Chain | undefined;
  toChain: Chain | undefined;
  token: TokenInfo | null;
};

export type ChainActions = {
  setToken: (token: TokenInfo | null) => void;
  resetToken: () => void;
  setTokenBridgeAddress: (address: Address | null) => void;
  switchChain: () => void;
  setNetworkLayer: (layer: NetworkLayer) => void;
  setNetworkType: (type: NetworkType) => void;
  setMessageServiceAddress: (address: Address | null) => void;
  setL1Chain: (chain: Chain) => void;
  setL2Chain: (chain: Chain) => void;
  setFromChain: (chain: Chain | undefined) => void;
  setToChain: (chain: Chain | undefined) => void;
};

export type ChainStore = ChainState & ChainActions;

export const defaultInitState: ChainState = {
  chains: [mainnet, sepolia, linea, lineaSepolia],
  networkType: NetworkType.UNKNOWN,
  networkLayer: NetworkLayer.UNKNOWN,
  messageServiceAddress: null,
  tokenBridgeAddress: null,
  l1Chain: undefined,
  l2Chain: undefined,
  fromChain: undefined,
  toChain: undefined,
  token: defaultTokensConfig.UNKNOWN[0],
};

const useChainStoreBase = create<ChainStore>()((set, get) => ({
  ...defaultInitState,
  setToken: (token) => set({ token }),
  resetToken: () => {
    const { networkLayer, networkType, token } = get();
    const networkLayerTo = networkLayer === NetworkLayer.L1 ? NetworkLayer.L2 : NetworkLayer.L1;
    if (networkType !== NetworkType.WRONG_NETWORK) {
      token && !token[networkLayerTo] && set({ token: defaultTokensConfig[networkType][0] });
    }
  },
  setTokenBridgeAddress: (address) => set({ tokenBridgeAddress: address }),
  switchChain: () => {
    const { fromChain, toChain, networkLayer, networkType, token } = get();
    const tempFromChain = fromChain;
    set({ fromChain: toChain, toChain: tempFromChain });

    const newNetworkLayer = networkLayer === NetworkLayer.L1 ? NetworkLayer.L2 : NetworkLayer.L1;
    set({ networkLayer: newNetworkLayer });

    set({ messageServiceAddress: config.networks[networkType][newNetworkLayer].messageServiceAddress });

    if (token?.type === TokenType.ERC20) {
      set({ tokenBridgeAddress: config.networks[networkType][newNetworkLayer].tokenBridgeAddress });
    } else if (token?.type === TokenType.USDC) {
      set({ tokenBridgeAddress: config.networks[networkType][newNetworkLayer].usdcBridgeAddress });
    }
  },
  setNetworkLayer: (layer) => set({ networkLayer: layer }),
  setNetworkType: (type) => set({ networkType: type }),
  setMessageServiceAddress: (address) => set({ messageServiceAddress: address }),
  setL1Chain: (chain) => set({ l1Chain: chain }),
  setL2Chain: (chain) => set({ l2Chain: chain }),
  setFromChain: (chain) => set({ fromChain: chain }),
  setToChain: (chain) => set({ toChain: chain }),
}));

export const useChainStore = createSelectorHooks(useChainStoreBase) as unknown as typeof useChainStoreBase &
  ZustandHookSelectors<ChainStore>;
