"use client";

import { ReactNode } from "react";
import { State, WagmiProvider } from "wagmi";
import { createWeb3Modal } from "@web3modal/wagmi/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { config, wagmiConfig } from "@/config";

const queryClient = new QueryClient();

if (!config.walletConnectId) throw new Error("Project ID is not defined");

createWeb3Modal({ wagmiConfig, projectId: config.walletConnectId });

type Web3ProviderProps = {
  children: ReactNode;
  initialState?: State;
};

export function Web3Provider({ children, initialState }: Web3ProviderProps) {
  return (
    <WagmiProvider config={wagmiConfig} initialState={initialState}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </WagmiProvider>
  );
}
