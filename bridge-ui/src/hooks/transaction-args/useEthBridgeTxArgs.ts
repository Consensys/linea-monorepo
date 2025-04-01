import { useMemo } from "react";
import { encodeFunctionData } from "viem";
import { useFormStore, useChainStore } from "@/stores";
import MessageService from "@/abis/MessageService.json";
import { isEth, isUndefinedOrNull, isZero } from "@/utils";
import { BridgeProvider, ChainLayer } from "@/types";
import { DEFAULT_ADDRESS_FOR_NON_CONNECTED_USER } from "@/constants";

type UseEthBridgeTxArgsProps = {
  isConnected: boolean;
};

const useEthBridgeTxArgs = ({ isConnected }: UseEthBridgeTxArgsProps) => {
  const { isL2Network, isL1Network, fromChainId, messageServiceAddress } = useChainStore((state) => ({
    fromChainId: state.fromChain.id,
    messageServiceAddress: state.fromChain.messageServiceAddress,
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
      !amount ||
      !toAddress ||
      (isZero(minimumFees) && isL2Network) ||
      (isUndefinedOrNull(bridgingFees) && isL1Network) ||
      (isZero(bridgingFees) && claim === "auto") ||
      !isEth(token) ||
      token.bridgeProvider !== BridgeProvider.NATIVE
    ) {
      return;
    }

    return {
      type: "bridge",
      args: {
        to: messageServiceAddress,
        data: encodeFunctionData({
          abi: MessageService.abi,
          functionName: "sendMessage",
          args: [toAddress, minimumFees + bridgingFees, "0x"],
        }),
        value: amount + minimumFees + bridgingFees,
        chainId: fromChainId,
      },
    };
  }, [
    amount,
    toAddress,
    minimumFees,
    isL2Network,
    bridgingFees,
    isL1Network,
    claim,
    token,
    messageServiceAddress,
    fromChainId,
  ]);
};

export default useEthBridgeTxArgs;
