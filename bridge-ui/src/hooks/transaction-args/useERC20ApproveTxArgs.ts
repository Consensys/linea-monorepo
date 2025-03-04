import { useAccount } from "wagmi";
import { encodeFunctionData } from "viem";
import { useChainStore } from "@/stores/chainStore";
import { useFormStore } from "@/stores/formStoreProvider";
import ERC20Abi from "@/abis/ERC20.json";
import { isEth } from "@/utils/tokens";
import { BridgeProvider } from "@/config/config";

type UseERC20ApproveTxArgsProps = {
  allowance?: bigint;
};

const useERC20ApproveTxArgs = ({ allowance }: UseERC20ApproveTxArgsProps) => {
  const { address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const token = useFormStore((state) => state.token);
  const amount = useFormStore((state) => state.amount);

  if (
    !address ||
    !fromChain ||
    !token ||
    isEth(token) ||
    !amount ||
    !allowance ||
    allowance >= amount ||
    token.bridgeProvider !== BridgeProvider.NATIVE
  ) {
    return;
  }

  return {
    type: "approve",
    args: {
      to: token[fromChain.layer],
      data: encodeFunctionData({
        abi: ERC20Abi,
        functionName: "approve",
        args: [fromChain.tokenBridgeAddress, amount],
      }),
      value: 0n,
      chainId: fromChain.id,
    },
  };
};

export default useERC20ApproveTxArgs;
