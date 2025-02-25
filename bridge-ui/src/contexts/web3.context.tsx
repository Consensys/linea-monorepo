"use client";

import { DynamicWagmiConnector, EthereumWalletConnectors, DynamicContextProvider } from "@/lib/dynamic";
import { ReactNode } from "react";
import { config } from "@/lib/wagmi";
import { WagmiProvider } from "wagmi";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const queryClient = new QueryClient();

type Web3ProviderProps = {
  children: ReactNode;
};

// export const cssOverrides = `
//     .modal > * {
//         font-size: 2rem !important;
//     }
//     .connect-button {
//         border-radius: 3rem;
//         padding: 1.2rem 2.4rem;
//         line-height: 1;
//         width: 100%;
//     }

//     .connect-button .typography {
//         font-size: 1.4rem;
//     }

//     .dynamic-widget-inline-controls__network-picker-main {
//         display: none;
//     }

//     .dynamic-widget-inline-controls__account-control-container {
//         border-radius: 3rem;
//         min-width: 100%;
//     }

//     .wallet-icon-with-network__container {
//         display: none;
//         width: 32px;
//         height: 32px;
//     }

//     .account-control__name {
//         display: none;
//     }

//     .dynamic-widget-inline-controls {
//         max-height: 100%;
//         width: 32px;
//         height: 32px;
//         border-radius: 100%;
//         background: linear-gradient(to top left,
//           #F8CECE 15.62%,
//           #ED7878 39.58%,
//           #E7781D 72.92%,
//           #D36E09 90.63%,
//           #D76E04 100%);
//     }

//     .dynamic-widget-inline-controls__account-control > svg {
//         display: none;
//     }
// `;

export function Web3Provider({ children }: Web3ProviderProps) {
  return (
    <DynamicContextProvider
      settings={{
        environmentId: process.env.NEXT_PUBLIC_DYNAMIC_ENVIRONMENT_ID!,
        walletConnectors: [EthereumWalletConnectors],
        appName: "Linea Bridge",
      }}
    >
      <WagmiProvider config={config}>
        <QueryClientProvider client={queryClient}>
          <DynamicWagmiConnector>{children}</DynamicWagmiConnector>
        </QueryClientProvider>
      </WagmiProvider>
    </DynamicContextProvider>
  );
}
