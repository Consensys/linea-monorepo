"use client";

import { State } from "wagmi";
import { Web3Provider } from "@/contexts/web3.context";
import { ModalProvider } from "@/contexts/modal.context";

type ProvidersProps = {
  children: JSX.Element;
  initialState?: State;
};

export function Providers({ children, initialState }: ProvidersProps) {
  return (
    <Web3Provider initialState={initialState}>
      <ModalProvider>{children}</ModalProvider>
    </Web3Provider>
  );
}
