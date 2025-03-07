import { useMemo } from "react";
import { TokenInfo } from "@/config";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { useChainStore } from "@/stores/chainStore";
import { ChainLayer } from "@/types";

export function useTokens(): TokenInfo[] {
  const tokensList = useTokenStore((state) => state.tokensList);
  const fromChain = useChainStore.useFromChain();

  return useMemo(() => {
    if (!fromChain) return [];

    if (fromChain.testnet) {
      if (fromChain.layer !== ChainLayer.L2) {
        return tokensList.SEPOLIA.filter((token) => !token.type.includes("native"));
      }
      return tokensList.SEPOLIA;
    }

    if (fromChain.layer !== ChainLayer.L2) {
      return tokensList.MAINNET.filter((token) => !token.type.includes("native"));
    }

    return tokensList.MAINNET;
  }, [fromChain, tokensList.MAINNET, tokensList.SEPOLIA]);
}
