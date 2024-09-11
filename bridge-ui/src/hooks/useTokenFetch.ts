import { useCallback } from "react";
import { readContract } from "@wagmi/core";
import { Address, Chain, zeroAddress } from "viem";
import log from "loglevel";
import TokenBridge from "@/abis/TokenBridge.json";
import { config, TokenInfo } from "@/config/config";
import { getChainNetworkLayer } from "@/utils/chainsUtil";
import { wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";

const useTokenFetch = () => {
  const { networkType, fromChain, toChain } = useChainStore((state) => ({
    networkType: state.networkType,
    fromChain: state.fromChain,
    toChain: state.toChain,
  }));

  const fetchBridgedToken = useCallback(
    async (fromChain: Chain, toChain: Chain, nativeToken: Address) => {
      const fromLayer = getChainNetworkLayer(fromChain);
      const toLayer = getChainNetworkLayer(toChain);
      if (!toLayer || !fromLayer) {
        return;
      }

      const _tokenBridgeAddress = config.networks[networkType][toLayer].tokenBridgeAddress;

      if (!_tokenBridgeAddress) {
        return;
      }

      try {
        const bridgedToken = (await readContract(wagmiConfig, {
          address: _tokenBridgeAddress,
          abi: TokenBridge.abi,
          functionName: "nativeToBridgedToken",
          args: [fromChain.id, nativeToken],
          chainId: toChain.id,
        })) as Address;

        return bridgedToken;
      } catch (error) {
        log.warn("Error fetching bridged token address");
      }
    },
    [networkType],
  );

  const fetchNativeToken = useCallback(
    async (chain: Chain, bridgedToken: Address) => {
      const layer = getChainNetworkLayer(chain);
      if (!layer) {
        return;
      }

      const _tokenBridgeAddress = config.networks[networkType][layer].tokenBridgeAddress;

      if (!_tokenBridgeAddress) {
        return;
      }

      try {
        const nativeToken = (await readContract(wagmiConfig, {
          address: _tokenBridgeAddress,
          abi: TokenBridge.abi,
          functionName: "bridgedToNativeToken",
          args: [bridgedToken],
          chainId: chain.id,
        })) as Address;

        return nativeToken;
      } catch (error) {
        log.warn("Error fetching native token address");
      }
    },
    [networkType],
  );

  const fillMissingTokenAddress = useCallback(
    async (token: TokenInfo) => {
      if (!fromChain || !toChain) {
        return;
      }

      // Since we don't if a token is native or bridged for a chain we try all the combinations
      // possible to find its counterpart on the other chain
      if (!token.L1 && token.L2) {
        token.L1 = (await fetchNativeToken(fromChain, token.L2)) || null;
        if (!token.L1 || token.L1 !== zeroAddress) return;

        token.L1 = (await fetchNativeToken(toChain, token.L2)) || null;
        if (!token.L1 || token.L1 !== zeroAddress) return;

        token.L1 = (await fetchBridgedToken(fromChain, toChain, token.L2)) || null;
        if (!token.L1 || token.L1 !== zeroAddress) return;

        token.L1 = (await fetchBridgedToken(toChain, fromChain, token.L2)) || null;
      } else if (token.L1) {
        token.L2 = (await fetchNativeToken(fromChain, token.L1)) || null;
        if (!token.L2 || token.L2 !== zeroAddress) return;

        token.L2 = (await fetchNativeToken(toChain, token.L1)) || null;
        if (!token.L2 || token.L2 !== zeroAddress) return;

        token.L2 = (await fetchBridgedToken(fromChain, toChain, token.L1)) || null;
        if (!token.L2 || token.L2 !== zeroAddress) return;

        token.L2 = (await fetchBridgedToken(toChain, fromChain, token.L1)) || null;
      }

      if (token.L1 === zeroAddress) token.L1 = null;
      if (token.L2 === zeroAddress) token.L2 = null;
    },
    [fromChain, toChain, fetchBridgedToken, fetchNativeToken],
  );

  return { fetchBridgedToken, fetchNativeToken, fillMissingTokenAddress };
};

export default useTokenFetch;
