import { useMemo } from "react";
import { TokenInfo } from "@/config";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { useChainStore } from "@/stores/chainStore";

export function useTokens(): TokenInfo[] {
  const tokensList = useTokenStore((state) => state.tokensList);
  const fromChain = useChainStore.useFromChain();

  return useMemo(() => {
    if (!fromChain) return [];

    if (fromChain.testnet) {
      return tokensList.SEPOLIA;
    }
    return tokensList.MAINNET;
  }, [fromChain?.id]);
}
