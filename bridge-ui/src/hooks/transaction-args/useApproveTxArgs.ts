import { useMemo } from "react";
import { encodeFunctionData, erc20Abi } from "viem";
import { useFormStore, useChainStore } from "@/stores";
import { isEth, isNull, isUndefined } from "@/utils";
import { isCctp } from "@/utils/tokens";

type UseERC20ApproveTxArgsProps = {
  isConnected: boolean;
  allowance?: bigint;
};

const useApproveTxArgs = ({ allowance }: UseERC20ApproveTxArgsProps) => {
  const { fromChainId, fromChainLayer, tokenBridgeAddress, cctpTokenMessengerV2Address } = useChainStore((state) => ({
    fromChainId: state.fromChain.id,
    fromChainLayer: state.fromChain.layer,
    tokenBridgeAddress: state.fromChain.tokenBridgeAddress,
    cctpTokenMessengerV2Address: state.fromChain.cctpTokenMessengerV2Address,
  }));
  const token = useFormStore((state) => state.token);
  const amount = useFormStore((state) => state.amount);

  return useMemo(() => {
    if (isEth(token) || isNull(amount) || isUndefined(allowance) || allowance >= amount) {
      return;
    }

    const spender = !isCctp(token) ? tokenBridgeAddress : cctpTokenMessengerV2Address;

    return {
      type: "approve",
      args: {
        to: token[fromChainLayer],
        data: encodeFunctionData({
          abi: erc20Abi,
          functionName: "approve",
          args: [spender, amount],
        }),
        value: 0n,
        chainId: fromChainId,
      },
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allowance, amount, fromChainId, token]);
};

export default useApproveTxArgs;
