"use client";

import { createContext, useRef, useContext, PropsWithChildren } from "react";
import { useStoreWithEqualityFn } from "zustand/traditional";
import { shallow } from "zustand/vanilla/shallow";
import { ChainStore, createChainStore } from "./chainStore";
import { ChainUpdater } from "./chainUpdater";

export type ChainStoreApi = ReturnType<typeof createChainStore>;

export const ChainStoreContext = createContext<ChainStoreApi | undefined>(undefined);

export interface ChainStoreProviderProps {}

export function ChainStoreProvider({ children }: PropsWithChildren<ChainStoreProviderProps>) {
  const storeRef = useRef<ChainStoreApi>();
  if (!storeRef.current) {
    storeRef.current = createChainStore();
  }

  return (
    <ChainStoreContext.Provider value={storeRef.current}>
      <ChainUpdater />
      {children}
    </ChainStoreContext.Provider>
  );
}

export const useChainStore = <T,>(selector: (store: ChainStore) => T): T => {
  const chainStoreContext = useContext(ChainStoreContext);

  if (!chainStoreContext) {
    throw new Error(`useChainStore must be used within ChainStoreProvider`);
  }

  return useStoreWithEqualityFn(chainStoreContext, selector, shallow);
};
