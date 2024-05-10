import React, { createContext, useEffect, useState } from 'react';

import { watchNetwork } from '@wagmi/core';
import { config } from '@/config';
import { Config, TokenInfo, TokenType } from '@/config/config';
import { Address, Chain } from 'viem';
import { defaultTokensConfig } from './token.context';

export enum NetworkType {
  UNKNOWN = 'UNKNOWN',
  MAINNET = 'MAINNET',
  SEPOLIA = 'SEPOLIA',
  WRONG_NETWORK = 'WRONG_NETWORK',
}

export enum NetworkLayer {
  UNKNOWN = 'UNKNOWN',
  L1 = 'L1',
  L2 = 'L2',
}

interface ChainContextData {
  networkType: NetworkType;
  networkLayer: NetworkLayer;
  messageServiceAddress: Address | null;
  tokenBridgeAddress: Address | null;

  l1Chain: Chain | undefined;
  l2Chain: Chain | undefined;
  activeChain: Chain | undefined;
  alternativeChain: Chain | undefined;
  fromChain: Chain | undefined;
  toChain: Chain | undefined;

  token: TokenInfo | null;

  setToken(token: TokenInfo | null): void;
  resetToken(): void;
  setTokenBridgeAddress(address: Address | null): void;
  switchChain(): void;
}

type Props = {
  children: JSX.Element;
};

export const ChainContext = createContext<ChainContextData>({} as ChainContextData);

function findLayerByChainId(config: Config, chainId: number | undefined): { network: string; layer: string } | null {
  // Iterate over each network in the config
  for (const networkName in config.networks) {
    const network = config.networks[networkName];
    // Check if the chainId matches either the L1 or L2 chainId for this network
    if (network.L1.chainId === chainId) {
      return { network: networkName, layer: 'L1' };
    } else if (network.L2.chainId === chainId) {
      return { network: networkName, layer: 'L2' };
    }
  }
  // If no match was found, return null
  return null;
}

export const ChainProvider = ({ children }: Props) => {
  const [networkType, setNetworkType] = useState<NetworkType>(NetworkType.UNKNOWN);
  const [networkLayer, setNetworkLayer] = useState<NetworkLayer>(NetworkLayer.UNKNOWN);
  const [tokenBridgeAddress, setTokenBridgeAddress] = useState<Address | null>(null);
  const [l1Chain, setL1Chain] = useState<Chain>();
  const [l2Chain, setL2Chain] = useState<Chain>();
  const [activeChain, setActiveChain] = useState<Chain | undefined>(undefined);
  const [alternativeChain, setAlternativeChain] = useState<Chain | undefined>(undefined);
  const [fromChain, setFromChain] = useState<Chain | undefined>(undefined);
  const [toChain, setToChain] = useState<Chain | undefined>(undefined);
  const [token, setToken] = useState<TokenInfo | null>(defaultTokensConfig.UNKNOWN[0]);
  const [messageServiceAddress, setMessageService] = useState<Address | null>(null);

  // Detect network changes
  useEffect(() => {
    watchNetwork((network) => {
      let networkType = NetworkType.UNKNOWN;
      let networkLayer = NetworkLayer.UNKNOWN;

      // Wrong network
      const chainExistsInChains = network.chains.find((chain: Chain) => chain.id === network?.chain?.id);

      if (!chainExistsInChains) {
        setNetworkType(NetworkType.WRONG_NETWORK);
        return;
      }

      // Get Network Type
      if (network?.chain?.testnet === true) {
        networkType = NetworkType.SEPOLIA;
        setNetworkType(networkType);
        !token && setToken(defaultTokensConfig.SEPOLIA[0]);
      } else if (network?.chain) {
        networkType = NetworkType.MAINNET;
        setNetworkType(networkType);
        !token && setToken(defaultTokensConfig.MAINNET[0]);
      }

      // Get Network Layer
      const layerFound = findLayerByChainId(config, network?.chain?.id);
      if (layerFound) {
        networkLayer = NetworkLayer[layerFound.layer as keyof typeof NetworkLayer];
        setNetworkLayer(networkLayer);
      }

      // Get Token Bridge
      if (networkType !== NetworkType.UNKNOWN && networkLayer !== NetworkLayer.UNKNOWN) {
        setMessageService(config.networks[networkType][networkLayer].messageServiceAddress);
        if (token?.type == TokenType.USDC) {
          setTokenBridgeAddress(config.networks[networkType][networkLayer].usdcBridgeAddress);
        } else {
          setTokenBridgeAddress(config.networks[networkType][networkLayer].tokenBridgeAddress);
        }
      } else {
        return;
      }

      // From chain, To chain
      let activeChain;
      let alternativeChain;

      switch (networkLayer) {
        case 'L1':
          activeChain = network.chains.find((chain) => chain.id === config.networks[networkType]['L1'].chainId);
          alternativeChain = network.chains.find((chain) => chain.id === config.networks[networkType]['L2'].chainId);

          activeChain && setL1Chain(activeChain);
          alternativeChain && setL2Chain(alternativeChain);
          break;
        case 'L2':
          activeChain = network.chains.find((chain) => chain.id === config.networks[networkType]['L2'].chainId);
          alternativeChain = network.chains.find((chain) => chain.id === config.networks[networkType]['L1'].chainId);

          alternativeChain && setL1Chain(alternativeChain);
          activeChain && setL2Chain(activeChain);
          break;
      }

      setActiveChain(activeChain);
      setAlternativeChain(alternativeChain);
      setFromChain(activeChain);
      setToChain(alternativeChain);
    });
  }, [token, token?.type]);

  // Switch bridge chain
  const switchChain = () => {
    const tempFromChain = fromChain;
    setFromChain(toChain);
    setToChain(tempFromChain);

    let newNetworkLayer;
    messageServiceAddress;
    if (networkLayer === 'L1') {
      newNetworkLayer = NetworkLayer.L2;
    } else {
      newNetworkLayer = NetworkLayer.L1;
    }
    setNetworkLayer(newNetworkLayer);

    if (newNetworkLayer == NetworkLayer.UNKNOWN) {
      return;
    }

    setMessageService(config.networks[networkType][newNetworkLayer].messageServiceAddress);
    if (token?.type == TokenType.ERC20) {
      setTokenBridgeAddress(config.networks[networkType][newNetworkLayer].tokenBridgeAddress);
    } else if (token?.type == TokenType.USDC) {
      setTokenBridgeAddress(config.networks[networkType][newNetworkLayer].usdcBridgeAddress);
    }
  };

  /**
   * Reset token to ETH if selected token has not been bridged to other chain
   */
  const resetToken = () => {
    const networkLayerTo = networkLayer === NetworkLayer.L1 ? NetworkLayer.L2 : NetworkLayer.L1;
    if (networkType !== NetworkType.WRONG_NETWORK) {
      token && !token[networkLayerTo] && setToken(defaultTokensConfig[networkType][0]);
    }
  };

  useEffect(() => {
    // Reset token if network type changes
    if (networkType === NetworkType.MAINNET || networkType === NetworkType.SEPOLIA) {
      setToken(defaultTokensConfig[networkType][0]);
    }
  }, [networkType]);

  return (
    <ChainContext.Provider
      value={{
        networkType,
        networkLayer,
        tokenBridgeAddress,
        l1Chain,
        l2Chain,
        activeChain,
        alternativeChain,
        fromChain,
        toChain,
        token,
        messageServiceAddress,
        setToken,
        resetToken,
        switchChain,
        setTokenBridgeAddress,
      }}
    >
      {children}
    </ChainContext.Provider>
  );
};
