import { NetworkLayer, NetworkType } from "@/config";
import { linea, mainnet, Chain, sepolia, lineaSepolia } from "viem/chains";

export const getChainNetworkLayer = (chain: Chain) => {
  switch (chain.id) {
    case linea.id:
    case lineaSepolia.id:
      return NetworkLayer.L2;
    case mainnet.id:
    case sepolia.id:
      return NetworkLayer.L1;
  }

  return;
};

export const getChainNetworkLayerByChainId = (chainId: number) => {
  switch (chainId) {
    case linea.id:
    case lineaSepolia.id:
      return NetworkLayer.L2;
    case mainnet.id:
    case sepolia.id:
      return NetworkLayer.L1;
  }

  return;
};

export const getChainNetworkType = (chain: Chain) => {
  switch (chain.id) {
    case linea.id:
    case mainnet.id:
      return NetworkType.MAINNET;
    case lineaSepolia.id:
    case sepolia.id:
      return NetworkType.SEPOLIA;
  }

  return;
};
