import { Address, getAddress } from "viem";
import { configSchema, Config } from "./config.schema";
import { TokenType } from "@/models/token";

export enum BridgeProvider {
  NATIVE = "NATIVE",
  CCTP = "CCTP",
}

export interface TokenInfo {
  type: TokenType[];
  name: string;
  symbol: string;
  decimals: number;
  L1: Address;
  L2: Address;
  image: string;
  isDefault: boolean;
  bridgeProvider: BridgeProvider;
}

export enum BridgeType {
  NATIVE = 1,
  ACROSS = 2,
}

export type NetworkTokens = {
  MAINNET: TokenInfo[];
  SEPOLIA: TokenInfo[];
};

export const config: Config = {
  chains: {
    1: {
      iconPath: "/images/logo/ethereum-rounded.svg",
      messageServiceAddress: getAddress(process.env.NEXT_PUBLIC_MAINNET_L1_MESSAGE_SERVICE ?? ""),
      tokenBridgeAddress: getAddress(process.env.NEXT_PUBLIC_MAINNET_L1_TOKEN_BRIDGE ?? ""),
      gasLimitSurplus: process.env.NEXT_PUBLIC_MAINNET_DEFAULT_GAS_LIMIT_SURPLUS
        ? BigInt(process.env.NEXT_PUBLIC_MAINNET_DEFAULT_GAS_LIMIT_SURPLUS)
        : BigInt(6000),
      profitMargin: process.env.NEXT_PUBLIC_MAINNET_PROFIT_MARGIN
        ? BigInt(process.env.NEXT_PUBLIC_MAINNET_PROFIT_MARGIN)
        : BigInt(1),
    },
    59144: {
      iconPath: "/images/logo/linea-mainnet.svg",
      messageServiceAddress: getAddress(process.env.NEXT_PUBLIC_MAINNET_LINEA_MESSAGE_SERVICE ?? ""),
      tokenBridgeAddress: getAddress(process.env.NEXT_PUBLIC_MAINNET_LINEA_TOKEN_BRIDGE ?? ""),
      gasLimitSurplus: process.env.NEXT_PUBLIC_MAINNET_DEFAULT_GAS_LIMIT_SURPLUS
        ? BigInt(process.env.NEXT_PUBLIC_MAINNET_DEFAULT_GAS_LIMIT_SURPLUS)
        : BigInt(6000),
      profitMargin: process.env.NEXT_PUBLIC_MAINNET_PROFIT_MARGIN
        ? BigInt(process.env.NEXT_PUBLIC_MAINNET_PROFIT_MARGIN)
        : BigInt(1),
    },
    11155111: {
      iconPath: "/images/logo/ethereum-rounded.svg",
      messageServiceAddress: getAddress(process.env.NEXT_PUBLIC_SEPOLIA_L1_MESSAGE_SERVICE ?? ""),
      tokenBridgeAddress: getAddress(process.env.NEXT_PUBLIC_SEPOLIA_L1_TOKEN_BRIDGE ?? ""),
      gasLimitSurplus: process.env.NEXT_PUBLIC_SEPOLIA_DEFAULT_GAS_LIMIT_SURPLUS
        ? BigInt(process.env.NEXT_PUBLIC_SEPOLIA_DEFAULT_GAS_LIMIT_SURPLUS)
        : BigInt(6000),
      profitMargin: process.env.NEXT_PUBLIC_SEPOLIA_PROFIT_MARGIN
        ? BigInt(process.env.NEXT_PUBLIC_SEPOLIA_PROFIT_MARGIN)
        : BigInt(1),
    },
    59141: {
      iconPath: "/images/logo/linea-sepolia.svg",
      messageServiceAddress: getAddress(process.env.NEXT_PUBLIC_SEPOLIA_LINEA_MESSAGE_SERVICE ?? ""),
      tokenBridgeAddress: getAddress(process.env.NEXT_PUBLIC_SEPOLIA_LINEA_TOKEN_BRIDGE ?? ""),
      gasLimitSurplus: process.env.NEXT_PUBLIC_SEPOLIA_DEFAULT_GAS_LIMIT_SURPLUS
        ? BigInt(process.env.NEXT_PUBLIC_SEPOLIA_DEFAULT_GAS_LIMIT_SURPLUS)
        : BigInt(6000),
      profitMargin: process.env.NEXT_PUBLIC_SEPOLIA_PROFIT_MARGIN
        ? BigInt(process.env.NEXT_PUBLIC_SEPOLIA_PROFIT_MARGIN)
        : BigInt(1),
    },
  },
  walletConnectId: process.env.NEXT_PUBLIC_WALLET_CONNECT_ID ?? "",
  storage: {
    // The storage will be cleared if its version is smaller than the one configured
    minVersion: process.env.NEXT_PUBLIC_STORAGE_MIN_VERSION ? parseInt(process.env.NEXT_PUBLIC_STORAGE_MIN_VERSION) : 1,
  },
};

export async function getConfiguration(): Promise<Config> {
  return config;
}

// Schema validation
getConfiguration().then((config) => {
  configSchema.parse(config);
});
