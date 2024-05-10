'use client';

import { useContext } from 'react';
import { switchNetwork } from '@wagmi/core';
import { config } from '@/config';
import log from 'loglevel';

import { ChainContext, NetworkLayer, NetworkType } from '@/contexts/chain.context';
import { useIsConnected } from '@/hooks';

export default function SwitchNetwork() {
  const context = useContext(ChainContext);
  const isConnected = useIsConnected();
  const { networkType, networkLayer } = context;

  const networks = config.networks;

  const switchNetworkHandler = async () => {
    if (networkLayer === NetworkLayer.UNKNOWN) {
      return;
    }

    try {
      if (networkType === NetworkType.SEPOLIA) {
        await switchNetwork({
          chainId: networks.MAINNET.L1.chainId,
        });
      } else {
        await switchNetwork({
          chainId: networks.SEPOLIA.L1.chainId,
        });
      }
    } catch (error) {
      log.error(error);
    }
  };

  if (!isConnected) return null;

  return (
    <div className="fixed bottom-24 left-4 md:left-10">
      <button id="try-network-btn" className="btn btn-info" onClick={() => switchNetworkHandler()}>
        Try {networkType === NetworkType.SEPOLIA ? 'Mainnet' : 'Testnet'}
      </button>
    </div>
  );
}
