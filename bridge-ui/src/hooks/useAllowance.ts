import { useAccount, useReadContract } from "wagmi";
import { erc20Abi } from "viem";
import { useChainStore } from "@/stores/chainStore";

const useAllowance = () => {
  // Wagmi
  const { address } = useAccount();

  // Context
  const { token, networkLayer, tokenBridgeAddress, fromChain } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
    tokenBridgeAddress: state.tokenBridgeAddress,
    fromChain: state.fromChain,
  }));

  const {
    data: allowance,
    queryKey,
    refetch,
  } = useReadContract({
    abi: erc20Abi,
    functionName: "allowance",
    args: [address ?? "0x", tokenBridgeAddress ?? "0x"],
    address: token?.[networkLayer] ?? "0x",
    query: {
      enabled: !!token && !!address && !!networkLayer && !!tokenBridgeAddress,
    },
    chainId: fromChain?.id,
  });

  return { allowance, queryKey, refetchAllowance: refetch };
};

export default useAllowance;
