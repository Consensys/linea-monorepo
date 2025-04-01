import { createStore } from "zustand";
import { linea, lineaSepolia, mainnet, sepolia } from "viem/chains";
import { Chain } from "@/types";
import { generateChain, generateChains } from "@/utils";

export type ChainState = {
  chains: Chain[];
  fromChain: Chain;
  toChain: Chain;
  availableChains: Chain[];
};

export type ChainActions = {
  switchChain: () => void;
  setFromChain: (chain: Chain | undefined) => void;
  setToChain: (chain: Chain | undefined) => void;
  setAvailableChains: (isTestnet: boolean) => void;
};

export type ChainStore = ChainState & ChainActions;

export const defaultInitState: ChainState = {
  chains: generateChains([mainnet, sepolia, linea, lineaSepolia]),
  fromChain: generateChain(mainnet),
  toChain: generateChain(linea),
  availableChains: generateChains([mainnet, linea]),
};

export const createChainStore = () =>
  createStore<ChainStore>()((set, get) => ({
    ...defaultInitState,
    switchChain: () => {
      const { fromChain, toChain } = get();
      const tempFromChain = fromChain;
      set({ fromChain: toChain, toChain: tempFromChain });
    },
    setFromChain: (chain) => set({ fromChain: chain }),
    setToChain: (chain) => set({ toChain: chain }),
    setAvailableChains: (isTestnet) => {
      const { chains } = get();
      const availableChains = chains.filter((chain) => chain.testnet === isTestnet);
      set({ availableChains });
    },
  }));
