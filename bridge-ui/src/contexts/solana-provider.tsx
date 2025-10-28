"use client";

import type { PropsWithChildren } from "react";
import { type Adapter, WalletAdapterNetwork } from "@solana/wallet-adapter-base";
import { clusterApiUrl } from "@solana/web3.js";
import { ConnectionProvider, WalletProvider } from "@solana/wallet-adapter-react";

const endpoint = clusterApiUrl(WalletAdapterNetwork.Mainnet);

const wallets: Adapter[] = [];

type SolanaWalletProviderProps = PropsWithChildren;

export function SolanaWalletProvider({ children }: SolanaWalletProviderProps) {
  return (
    <ConnectionProvider endpoint={endpoint}>
      <WalletProvider wallets={wallets} autoConnect>
        {children}
      </WalletProvider>
    </ConnectionProvider>
  );
}
