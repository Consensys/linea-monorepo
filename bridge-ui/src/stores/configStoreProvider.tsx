"use client";

import { createContext, useRef, useContext, PropsWithChildren } from "react";
import { useStoreWithEqualityFn } from "zustand/traditional";
import { shallow } from "zustand/vanilla/shallow";
import { ConfigStore, createConfigStore } from "./configStore";

export type ConfigStoreApi = ReturnType<typeof createConfigStore>;

export const ConfigStoreContext = createContext<ConfigStoreApi | undefined>(undefined);

export interface ConfigStoreProviderProps {}

export function ConfigStoreProvider({ children }: PropsWithChildren<ConfigStoreProviderProps>) {
  const storeRef = useRef<ConfigStoreApi>();
  if (!storeRef.current) {
    storeRef.current = createConfigStore();
  }

  return <ConfigStoreContext.Provider value={storeRef.current}>{children}</ConfigStoreContext.Provider>;
}

export const useConfigStore = <T,>(selector: (store: ConfigStore) => T): T => {
  const configStoreContext = useContext(ConfigStoreContext);

  if (!configStoreContext) {
    throw new Error(`useConfigStore must be used within ConfigStoreProvider`);
  }

  return useStoreWithEqualityFn(configStoreContext, selector, shallow);
};
