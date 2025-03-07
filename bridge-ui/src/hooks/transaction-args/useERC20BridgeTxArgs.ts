import { useMemo } from "react";
import { useAccount } from "wagmi";
import { encodeFunctionData } from "viem";
import { useFormStore, useChainStore } from "@/stores";
import TokenBridge from "@/abis/TokenBridge.json";
import { isEth } from "@/utils";
import { BridgeProvider } from "@/types";

type UseERC20BridgeTxArgsProps = {
  allowance?: bigint;
};

const useERC20BridgeTxArgs = ({ allowance }: UseERC20BridgeTxArgsProps) => {
  const { address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const token = useFormStore((state) => state.token);
  const amount = useFormStore((state) => state.amount);
  const recipient = useFormStore((state) => state.recipient);
  const minimumFees = useFormStore((state) => state.minimumFees);
  const bridgingFees = useFormStore((state) => state.bridgingFees);

  return useMemo(() => {
    if (
      !address ||
      !fromChain ||
      !token ||
      !amount ||
      !allowance ||
      allowance < amount ||
      !recipient ||
      !minimumFees ||
      !bridgingFees ||
      isEth(token) ||
      token.bridgeProvider !== BridgeProvider.NATIVE
    ) {
      return;
    }

    return {
      type: "bridge",
      args: {
        to: fromChain.tokenBridgeAddress,
        data: encodeFunctionData({
          abi: TokenBridge.abi,
          functionName: "bridgeToken",
          args: [token[fromChain.layer], amount, recipient],
        }),
        value: minimumFees + bridgingFees,
        chainId: fromChain.id,
      },
    };
  }, [address, allowance, amount, bridgingFees, fromChain, minimumFees, recipient, token]);
};

export default useERC20BridgeTxArgs;
