import { useMemo } from "react";
import { LineaSDK, Network } from "@consensys/linea-sdk";
import { L1MessageServiceContract, L2MessageServiceContract } from "@consensys/linea-sdk/dist/lib/contracts";
import { NetworkType } from "@/config";
import { useChainStore } from "@/stores/chainStore";

interface LineaSDKContracts {
  L1: L1MessageServiceContract;
  L2: L2MessageServiceContract;
}

const useLineaSDK = () => {
  const networkType = useChainStore((state) => state.networkType);

  const { lineaSDK, lineaSDKContracts } = useMemo(() => {
    const infuraKey = process.env.NEXT_PUBLIC_INFURA_ID;
    if (!infuraKey) return { lineaSDK: null, lineaSDKContracts: null };

    let l1RpcUrl;
    let l2RpcUrl;
    switch (networkType) {
      case NetworkType.MAINNET:
        l1RpcUrl = `https://mainnet.infura.io/v3/${infuraKey}`;
        l2RpcUrl = `https://linea-mainnet.infura.io/v3/${infuraKey}`;
        break;
      case NetworkType.SEPOLIA:
        l1RpcUrl = `https://sepolia.infura.io/v3/${infuraKey}`;
        l2RpcUrl = `https://linea-sepolia.infura.io/v3/${infuraKey}`;
        break;
      default:
        return { lineaSDK: null, lineaSDKContracts: null };
    }

    const sdk = new LineaSDK({
      l1RpcUrl,
      l2RpcUrl,
      network: `linea-${networkType.toLowerCase()}` as Network,
      mode: "read-only",
    });

    const newLineaSDKContracts: LineaSDKContracts = {
      L1: sdk.getL1Contract(),
      L2: sdk.getL2Contract(),
    };

    return { lineaSDK: sdk, lineaSDKContracts: newLineaSDKContracts };
  }, [networkType]);

  return { lineaSDK, lineaSDKContracts };
};

export default useLineaSDK;
