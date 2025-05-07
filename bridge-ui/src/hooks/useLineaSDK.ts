import { useMemo } from "react";
import { LineaSDK, Network } from "@consensys/linea-sdk";
import { linea, lineaSepolia, mainnet, sepolia } from "viem/chains";
import { useChainStore } from "@/stores";
import { CHAINS_RPC_URLS } from "@/constants";

const useLineaSDK = () => {
  const fromChain = useChainStore.useFromChain();

  const { lineaSDK } = useMemo(() => {
    let l1RpcUrl;
    let l2RpcUrl;
    if (fromChain.testnet) {
      l1RpcUrl = CHAINS_RPC_URLS[sepolia.id];
      l2RpcUrl = CHAINS_RPC_URLS[lineaSepolia.id];
    } else {
      l1RpcUrl = CHAINS_RPC_URLS[mainnet.id];
      l2RpcUrl = CHAINS_RPC_URLS[linea.id];
    }

    const sdk = new LineaSDK({
      l1RpcUrlOrProvider: l1RpcUrl,
      l2RpcUrlOrProvider: l2RpcUrl,
      network: `linea-${fromChain.testnet ? "sepolia" : "mainnet"}` as Network,
      mode: "read-only",
    });

    return { lineaSDK: sdk };
  }, [fromChain.testnet]);

  return { lineaSDK };
};

export default useLineaSDK;
