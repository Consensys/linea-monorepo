import { useAccount, useReadContract } from "wagmi";
import { erc20Abi } from "viem";
import { useChainStore } from "@/stores/chainStore";
import { useFormStore } from "@/stores/formStoreProvider";
import { isEth } from "@/utils/tokens";

const useAllowance = () => {
  const { address } = useAccount();
  const token = useFormStore((state) => state.token);
  const fromChain = useChainStore.useFromChain();

  const {
    data: allowance,
    queryKey,
    refetch,
  } = useReadContract({
    abi: erc20Abi,
    functionName: "allowance",
    args: [address ?? "0x", fromChain.tokenBridgeAddress],
    address: token[fromChain.layer] ?? "0x",
    query: {
      enabled: !!token && !isEth(token) && !!address && !!fromChain,
    },
    chainId: fromChain.id,
  });

  return { allowance, queryKey, refetchAllowance: refetch };
};

export default useAllowance;
