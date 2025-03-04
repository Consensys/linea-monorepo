import { useMemo } from "react";
import { useAccount } from "wagmi";
import { encodeFunctionData } from "viem";
import { useChainStore } from "@/stores/chainStore";
import { useFormStore } from "@/stores/formStoreProvider";
import MessageService from "@/abis/MessageService.json";
import { isEth } from "@/utils/tokens";
import { BridgeProvider } from "@/config/config";
import { ChainLayer } from "@/types";

const useEthBridgeTxArgs = () => {
  const { address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const token = useFormStore((state) => state.token);
  const amount = useFormStore((state) => state.amount);
  const recipient = useFormStore((state) => state.recipient);
  const minimumFees = useFormStore((state) => state.minimumFees);
  const bridgingFees = useFormStore((state) => state.bridgingFees);
  const claim = useFormStore((state) => state.claim);

  return useMemo(() => {
    if (
      !address ||
      !fromChain ||
      !token ||
      !amount ||
      !recipient ||
      (minimumFees === 0n && fromChain.layer === ChainLayer.L2) ||
      (!bridgingFees && fromChain.layer === ChainLayer.L1) ||
      (bridgingFees === 0n && claim === "auto") ||
      !isEth(token) ||
      token.bridgeProvider !== BridgeProvider.NATIVE
    ) {
      return;
    }

    return {
      type: "bridge",
      args: {
        to: fromChain.messageServiceAddress,
        data: encodeFunctionData({
          abi: MessageService.abi,
          functionName: "sendMessage",
          args: [recipient, minimumFees + bridgingFees, "0x"],
        }),
        value: amount + minimumFees + bridgingFees,
        chainId: fromChain.id,
      },
    };
  }, [address, fromChain, token, amount, recipient, minimumFees, bridgingFees, claim]);
};

export default useEthBridgeTxArgs;
