import { useMemo } from "react";
import { encodeFunctionData } from "viem";
import { useFormStore, useChainStore } from "@/stores";
import TokenBridge from "@/abis/TokenBridge.json";
import { isEth, isNull, isUndefined, isUndefinedOrNull, isZero } from "@/utils";
import { BridgeProvider, ChainLayer } from "@/types";
import { DEFAULT_ADDRESS_FOR_NON_CONNECTED_USER } from "@/constants";

type UseERC20BridgeTxArgsProps = {
  isConnected: boolean;
  allowance?: bigint;
};

const useERC20BridgeTxArgs = ({ isConnected, allowance }: UseERC20BridgeTxArgsProps) => {
  const { isL2Network, isL1Network, fromChainLayer, fromChainId, tokenBridgeAddress } = useChainStore((state) => ({
    fromChainLayer: state.fromChain.layer,
    fromChainId: state.fromChain.id,
    tokenBridgeAddress: state.fromChain.tokenBridgeAddress,
    isL2Network: state.fromChain.layer === ChainLayer.L2,
    isL1Network: state.fromChain.layer === ChainLayer.L1,
  }));
  const token = useFormStore((state) => state.token);
  const amount = useFormStore((state) => state.amount);
  const recipient = useFormStore((state) => state.recipient);
  const minimumFees = useFormStore((state) => state.minimumFees);
  const bridgingFees = useFormStore((state) => state.bridgingFees);
  const claim = useFormStore((state) => state.claim);

  const toAddress = isConnected ? recipient : DEFAULT_ADDRESS_FOR_NON_CONNECTED_USER;

  return useMemo(() => {
    if (
      isEth(token) ||
      isNull(amount) ||
      (isConnected && (isUndefined(allowance) || allowance < amount)) ||
      !toAddress ||
      (isZero(minimumFees) && isL2Network) ||
      (isUndefinedOrNull(bridgingFees) && isL1Network) ||
      (isZero(bridgingFees) && claim === "auto") ||
      token.bridgeProvider !== BridgeProvider.NATIVE
    ) {
      return;
    }

    return {
      type: "bridge",
      args: {
        to: tokenBridgeAddress,
        data: encodeFunctionData({
          abi: TokenBridge.abi,
          functionName: "bridgeToken",
          args: [token[fromChainLayer], amount, toAddress],
        }),
        value: minimumFees + bridgingFees,
        chainId: fromChainId,
      },
    };
  }, [
    token,
    amount,
    isConnected,
    allowance,
    toAddress,
    minimumFees,
    isL2Network,
    bridgingFees,
    isL1Network,
    claim,
    tokenBridgeAddress,
    fromChainLayer,
    fromChainId,
  ]);
};

export default useERC20BridgeTxArgs;
