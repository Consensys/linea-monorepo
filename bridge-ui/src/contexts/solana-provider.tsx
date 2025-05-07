"use client";

import type { PropsWithChildren } from "react";
import { DynamicSolanaProvider } from "./dynamic-solana-provider";
import { clusterApiUrl } from "@solana/web3.js";
import { Adapter, WalletAdapterNetwork } from "@solana/wallet-adapter-base";
import { ConnectionProvider, WalletProvider } from "@solana/wallet-adapter-react";

const endpoint = clusterApiUrl(WalletAdapterNetwork.Mainnet);

const wallets: Adapter[] = [];

type SolanaWalletProviderProps = PropsWithChildren;

export function SolanaWalletProvider({ children }: SolanaWalletProviderProps) {
  return (
    <ConnectionProvider endpoint={endpoint}>
      <WalletProvider wallets={wallets} autoConnect>
        <DynamicSolanaProvider>{children}</DynamicSolanaProvider>
      </WalletProvider>
    </ConnectionProvider>
  );
}
