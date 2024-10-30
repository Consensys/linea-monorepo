import { watchAccount } from "@wagmi/core";
import { useEffect } from "react";
import { Chain } from "viem";
import { Config, config, NetworkLayer, NetworkType, TokenType, wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import { defaultTokensConfig } from "@/stores/tokenStore";

function findLayerByChainId(config: Config, chainId: number | undefined): { network: string; layer: string } | null {
  // Iterate over each network in the config
  for (const networkName in config.networks) {
    const network = config.networks[networkName];
    // Check if the chainId matches either the L1 or L2 chainId for this network
    if (network.L1.chainId === chainId) {
      return { network: networkName, layer: "L1" };
    } else if (network.L2.chainId === chainId) {
      return { network: networkName, layer: "L2" };
    }
  }
  // If no match was found, return null
  return null;
}

const findChainById = (chainId: number) => wagmiConfig.chains.find((chain) => chain.id === chainId);

const useInitialiseChain = () => {
  const token = useChainStore((state) => state.token);
  const networkType = useChainStore((state) => state.networkType);
  const {
    setNetworkType,
    setNetworkLayer,
    setToken,
    setTokenBridgeAddress,
    setMessageServiceAddress,
    setL1Chain,
    setL2Chain,
    setFromChain,
    setToChain,
  } = useChainStore((state) => ({
    setNetworkType: state.setNetworkType,
    setNetworkLayer: state.setNetworkLayer,
    setToken: state.setToken,
    setTokenBridgeAddress: state.setTokenBridgeAddress,
    setMessageServiceAddress: state.setMessageServiceAddress,
    setL1Chain: state.setL1Chain,
    setL2Chain: state.setL2Chain,
    setFromChain: state.setFromChain,
    setToChain: state.setToChain,
  }));

  useEffect(() => {
    const unwatch = watchAccount(wagmiConfig, {
      onChange(account) {
        let networkType = NetworkType.UNKNOWN;
        let networkLayer = NetworkLayer.UNKNOWN;

        // Wrong network
        const chainExistsInChains = wagmiConfig.chains.find((chain: Chain) => chain.id === account?.chain?.id);

        if (!chainExistsInChains) {
          setNetworkType(NetworkType.WRONG_NETWORK);
          return;
        }

        // Get Network Type
        if (account?.chain?.testnet === true) {
          networkType = NetworkType.SEPOLIA;
          setNetworkType(networkType);
          !token && setToken(defaultTokensConfig.SEPOLIA[0]);
        } else if (account?.chain) {
          networkType = NetworkType.MAINNET;
          setNetworkType(networkType);
          !token && setToken(defaultTokensConfig.MAINNET[0]);
        }

        // Get Network Layer
        const layerFound = findLayerByChainId(config, account?.chain?.id);
        if (layerFound) {
          networkLayer = NetworkLayer[layerFound.layer as keyof typeof NetworkLayer];
          setNetworkLayer(networkLayer);
        }

        // Get Token Bridge
        if (networkType !== NetworkType.UNKNOWN && networkLayer !== NetworkLayer.UNKNOWN) {
          setMessageServiceAddress(config.networks[networkType][networkLayer].messageServiceAddress);
          if (token?.type === TokenType.USDC) {
            setTokenBridgeAddress(config.networks[networkType][networkLayer].usdcBridgeAddress);
          } else {
            setTokenBridgeAddress(config.networks[networkType][networkLayer].tokenBridgeAddress);
          }
        } else {
          return;
        }

        // From chain, To chain
        let fromChain: Chain | undefined;
        let toChain: Chain | undefined;

        switch (networkLayer) {
          case "L1":
            fromChain = findChainById(config.networks[networkType]["L1"].chainId);
            toChain = findChainById(config.networks[networkType]["L2"].chainId);

            fromChain && setL1Chain(fromChain);
            toChain && setL2Chain(toChain);
            break;
          case "L2":
            fromChain = findChainById(config.networks[networkType]["L2"].chainId);
            toChain = findChainById(config.networks[networkType]["L1"].chainId);

            toChain && setL1Chain(toChain);
            fromChain && setL2Chain(fromChain);
            break;
        }

        setFromChain(fromChain);
        setToChain(toChain);
      },
    });

    return () => {
      unwatch();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, token?.type]);

  useEffect(() => {
    // Reset token if network type changes
    if (networkType === NetworkType.MAINNET || networkType === NetworkType.SEPOLIA) {
      setToken(defaultTokensConfig[networkType][0]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [networkType]);
};

export default useInitialiseChain;
