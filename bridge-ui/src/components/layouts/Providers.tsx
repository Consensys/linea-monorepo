"use client";

import { State } from "wagmi";
import { UIProvider } from "@/contexts/ui.context";
import { Web3Provider } from "@/contexts/web3.context";

type ProvidersProps = {
  children: JSX.Element;
  initialState?: State;
};

export function Providers({ children, initialState }: ProvidersProps) {
  return (
    <UIProvider>
      <Web3Provider initialState={initialState}>{children}</Web3Provider>
    </UIProvider>
  );
}
