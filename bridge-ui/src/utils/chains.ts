import { Address } from "viem";
import { linea, mainnet, Chain as ViemChain, sepolia, lineaSepolia } from "viem/chains";
import { config } from "@/config";
import { Chain, ChainLayer, SupportedChainIds } from "@/types";

const getChainName = (chainId: number) => {
  switch (chainId) {
    case linea.id:
      return "Linea";
    case lineaSepolia.id:
      return "Linea Sepolia";
    case mainnet.id:
      return "Ethereum";
    case sepolia.id:
      return "Sepolia";
    default:
      return "";
  }
};

export const generateChain = (chain: ViemChain): Chain => {
  return {
    id: chain.id as SupportedChainIds,
    name: getChainName(chain.id),
    iconPath: config.chains[chain.id].iconPath,
    nativeCurrency: chain.nativeCurrency,
    blockExplorers: chain.blockExplorers,
    testnet: Boolean(chain.testnet),
    layer: getChainNetworkLayer(chain.id),
    messageServiceAddress: config.chains[chain.id].messageServiceAddress as Address,
    tokenBridgeAddress: config.chains[chain.id].tokenBridgeAddress as Address,
    gasLimitSurplus: config.chains[chain.id].gasLimitSurplus,
    profitMargin: config.chains[chain.id].profitMargin,
    cctpDomain: config.chains[chain.id].cctpDomain,
    cctpTokenMessengerV2Address: config.chains[chain.id].cctpTokenMessengerV2Address as Address,
    cctpMessageTransmitterV2Address: config.chains[chain.id].cctpMessageTransmitterV2Address as Address,
  };
};

export const generateChains = (chains: ViemChain[]): Chain[] => {
  return chains.map(generateChain);
};

export const getChainNetworkLayer = (chainId: number) => {
  switch (chainId) {
    case linea.id:
    case lineaSepolia.id:
      return ChainLayer.L2;
    case mainnet.id:
    case sepolia.id:
      return ChainLayer.L1;
    default:
      throw new Error(`Unsupported chain id: ${chainId}`);
  }
};

export const getChainLogoPath = (chainId: number) => {
  switch (chainId) {
    case linea.id:
      return `${process.env.NEXT_PUBLIC_BASE_PATH}/images/logo/linea-rounded.svg`;
    case lineaSepolia.id:
      return `${process.env.NEXT_PUBLIC_BASE_PATH}/images/logo/linea-sepolia.svg`;
    case mainnet.id:
    case sepolia.id:
      return `${process.env.NEXT_PUBLIC_BASE_PATH}/images/logo/ethereum-rounded.svg`;
    default:
      return "";
  }
};
