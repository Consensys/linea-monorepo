"use client";

import { switchChain } from "@wagmi/core";
import { NetworkLayer, NetworkType, config, wagmiConfig } from "@/config";
import log from "loglevel";
import { useAccount } from "wagmi";
import { useChainStore } from "@/stores/chainStore";

export default function SwitchNetwork() {
  const { networkType, networkLayer } = useChainStore((state) => ({
    networkType: state.networkType,
    networkLayer: state.networkLayer,
  }));
  const { isConnected } = useAccount();

  const networks = config.networks;

  const switchNetworkHandler = async () => {
    if (networkLayer === NetworkLayer.UNKNOWN) {
      return;
    }

    try {
      if (networkType === NetworkType.SEPOLIA) {
        await switchChain(wagmiConfig, {
          chainId: networks.MAINNET.L1.chainId,
        });
      } else {
        await switchChain(wagmiConfig, {
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
      <button id="try-network-btn" className="btn btn-info uppercase" onClick={() => switchNetworkHandler()}>
        Try {networkType === NetworkType.SEPOLIA ? "Mainnet" : "Testnet"}
      </button>
    </div>
  );
}
