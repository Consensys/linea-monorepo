"use client";

import { type ReactNode, createContext, useRef, useContext } from "react";
import { shallow } from "zustand/vanilla/shallow";
import { FormState, type FormStore, createFormStore } from "./formStore";
import { useStoreWithEqualityFn } from "zustand/traditional";
import { useAccount } from "wagmi";
import { Token } from "@/types";

export type FormStoreApi = ReturnType<typeof createFormStore>;

export const FormStoreContext = createContext<FormStoreApi | undefined>(undefined);

export interface FormStoreProviderProps {
  children: ReactNode;
  initialToken: Token;
}

export function FormStoreProvider({ children, initialToken }: FormStoreProviderProps) {
  const { address } = useAccount();

  const initialState: FormState = {
    token: initialToken,
    claim: "auto",
    amount: null,
    minimumFees: 0n,
    gasFees: 0n,
    bridgingFees: 0n,
    balance: 0n,
    recipient: address || "0x",
  };

  const storeRef = useRef<FormStoreApi>();
  if (!storeRef.current) {
    storeRef.current = createFormStore(initialState);
  }

  return <FormStoreContext.Provider value={storeRef.current}>{children}</FormStoreContext.Provider>;
}

export const useFormStore = <T,>(selector: (store: FormStore) => T): T => {
  const formStoreContext = useContext(FormStoreContext);

  if (!formStoreContext) {
    throw new Error(`useFormStore must be used within FormStoreProvider`);
  }

  return useStoreWithEqualityFn(formStoreContext, selector, shallow);
};
