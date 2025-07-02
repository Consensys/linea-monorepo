"use client";

import { PropsWithChildren } from "react";
import { WagmiProvider } from "wagmi";
import {
  DynamicWagmiConnector,
  EthereumWalletConnectors,
  DynamicContextProvider,
  SolanaWalletConnectors,
} from "@/lib/dynamic";
import { config as wagmiConfig } from "@/lib/wagmi";
import { config } from "@/config";
import { SolanaWalletProvider } from "./solana-provider";

type Web3ProviderProps = PropsWithChildren;

export const cssOverrides = `
  .connect-button {
    font-size: 0.875rem;
    border-radius: 1.875rem;
    padding: 0.75rem 1.5rem;
    text-align: center;
    line-height: 1;
    cursor: pointer;
  }

  .connect-button .typography {
    font-size: 0.875rem;
  }

  .dynamic-widget-inline-controls {
    background-color: transparent;
    border: 1px solid var(--color-indigo);
    border-radius: 1.875rem;
  }

  .dynamic-widget-inline-controls__account-control-container {
    display: flex;
    align-items: center;
  }

  .dynamic-widget-inline-controls .network-switch-control__network-name {
    color: var(--color-white);

    @media screen and (max-width: 767px) {
      display: none;
    }
  }

  .dynamic-widget-inline-controls .network-switch-control__container--error {
    border-radius: 1.875rem 0 0 1.875rem;
  }

  .dynamic-widget-inline-controls .network-switch-control__arrow-icon {
    color: var(--color-white);
  }

  .account-control__name {
    color: var(--color-indigo);

    @media screen and (max-width: 767px) {
      display: none;
    }
  }

  .account-control__icon {
    color: var(--color-silver);

    @media screen and (max-width: 767px) {
      margin-left: 0 !important;
    }
  }

  .account-control__container {
    @media screen and (max-width: 767px) {
      justify-content: center;
    }
  }

  .account-control__container:hover,
  .account-control__container--active {
    background-color: unset;
  }
`;

export function Web3Provider({ children }: Web3ProviderProps) {
  return (
    <DynamicContextProvider
      settings={{
        environmentId: config.dynamicEnvironmentId,
        walletConnectors: [EthereumWalletConnectors, SolanaWalletConnectors],
        initialAuthenticationMode: "connect-only",
        mobileExperience: "redirect",
        appName: "Linea Bridge",
        cssOverrides,
      }}
    >
      <WagmiProvider config={wagmiConfig}>
        <DynamicWagmiConnector>
          <SolanaWalletProvider>{children}</SolanaWalletProvider>
        </DynamicWagmiConnector>
      </WagmiProvider>
    </DynamicContextProvider>
  );
}
