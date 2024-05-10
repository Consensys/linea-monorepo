'use client';

import { EIP6963Connector, createWeb3Modal } from '@web3modal/wagmi/react';

import { WagmiConfig, createConfig } from 'wagmi';
import { configureChains } from '@wagmi/core';
import { sepolia, mainnet, linea } from 'viem/chains';
import { publicProvider } from 'wagmi/providers/public';
import { infuraProvider } from 'wagmi/providers/infura';
import { InjectedConnector } from 'wagmi/connectors/injected';
import { CoinbaseWalletConnector } from 'wagmi/connectors/coinbaseWallet';

import { WalletConnectConnector } from 'wagmi/connectors/walletConnect';
import { config } from '@/config';
import { lineaSepolia } from '@/utils/SepoliaChain';

const { chains, publicClient } = configureChains(
  [mainnet, sepolia, linea, lineaSepolia],
  [infuraProvider({ apiKey: process.env.NEXT_PUBLIC_INFURA_ID ?? '' }), publicProvider()]
);

const wagmiConfig = createConfig({
  autoConnect: true,
  connectors: [
    new WalletConnectConnector({
      chains,
      options: {
        projectId: config.walletConnectId,
        showQrModal: false,
      },
    }),
    new EIP6963Connector({ chains }),
    new InjectedConnector({ chains, options: { shimDisconnect: true } }),
    new CoinbaseWalletConnector({
      chains,
      options: { appName: 'Linea Bridge' },
    }),
  ],
  publicClient,
});

createWeb3Modal({ wagmiConfig, projectId: config.walletConnectId, chains });

type Props = {
  children: JSX.Element;
};

export function Web3ModalContext({ children }: Props) {
  return <WagmiConfig config={wagmiConfig}>{children}</WagmiConfig>;
}