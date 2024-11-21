"use client";

import { ReactNode } from "react";
import { WagmiProvider } from "wagmi";
import { createWeb3Modal } from "@web3modal/wagmi/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { config, wagmiConfig } from "@/config";

const queryClient = new QueryClient();

if (!config.walletConnectId) throw new Error("Project ID is not defined");

createWeb3Modal({ wagmiConfig, projectId: config.walletConnectId });

type Web3ProviderProps = {
  children: ReactNode;
};

export function Web3Provider({ children }: Web3ProviderProps) {
  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </WagmiProvider>
  );
}
