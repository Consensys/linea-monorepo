import { useMemo } from "react";
import { useChainStore, useTokenStore } from "@/stores";
import { ChainLayer, Token } from "@/types";
import { config } from "@/config";
import { USDC_SYMBOL } from "@/constants";

const useTokens = (): Token[] => {
  const tokensList = useTokenStore((state) => state.tokensList);
  const { fromChainLayer, isFromChainTestnet } = useChainStore((state) => ({
    fromChainLayer: state.fromChain.layer,
    isFromChainTestnet: state.fromChain.testnet,
  }));

  return useMemo(() => {
    if (isFromChainTestnet) {
      if (fromChainLayer !== ChainLayer.L2) {
        return tokensList.SEPOLIA.filter(
          (token) => !token.type.includes("native") && (config.isCctpEnabled || token.symbol !== USDC_SYMBOL),
        );
      }
      return config.isCctpEnabled
        ? tokensList.SEPOLIA
        : tokensList.SEPOLIA.filter((token) => token.symbol !== USDC_SYMBOL);
    }

    if (fromChainLayer !== ChainLayer.L2) {
      return tokensList.MAINNET.filter(
        (token) => !token.type.includes("native") && (config.isCctpEnabled || token.symbol !== USDC_SYMBOL),
      );
    }

    return config.isCctpEnabled
      ? tokensList.MAINNET
      : tokensList.MAINNET.filter((token) => token.symbol !== USDC_SYMBOL);
  }, [fromChainLayer, isFromChainTestnet, tokensList.MAINNET, tokensList.SEPOLIA]);
};

export default useTokens;
