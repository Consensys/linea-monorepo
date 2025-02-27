import { useAccount, useReadContract } from "wagmi";
import { erc20Abi } from "viem";
import { useChainStore } from "@/stores/chainStore";
import { useSelectedToken } from "./useSelectedToken";

const useAllowance = () => {
  const { address } = useAccount();
  const token = useSelectedToken();
  const fromChain = useChainStore.useFromChain();

  const {
    data: allowance,
    queryKey,
    refetch,
  } = useReadContract({
    abi: erc20Abi,
    functionName: "allowance",
    args: [address ?? "0x", fromChain?.tokenBridgeAddress ?? "0x"],
    address: token?.[fromChain.layer] ?? "0x",
    query: {
      enabled: !!token && !!address && !!fromChain,
    },
    chainId: fromChain.id,
  });

  return { allowance, queryKey, refetchAllowance: refetch };
};

export default useAllowance;
