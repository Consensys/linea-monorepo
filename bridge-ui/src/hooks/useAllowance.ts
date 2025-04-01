import { useAccount, useReadContract } from "wagmi";
import { erc20Abi } from "viem";
import { useFormStore, useChainStore } from "@/stores";
import { isEth } from "@/utils";
import { isCctp } from "@/utils/tokens";

const useAllowance = () => {
  const { address } = useAccount();
  const token = useFormStore((state) => state.token);
  const { fromChainId, fromChainLayer, tokenBridgeAddress, cctpTokenMessengerV2Address } = useChainStore((state) => ({
    fromChainId: state.fromChain.id,
    fromChainLayer: state.fromChain.layer,
    tokenBridgeAddress: state.fromChain.tokenBridgeAddress,
    cctpTokenMessengerV2Address: state.fromChain.cctpTokenMessengerV2Address,
  }));
  const spender = !isCctp(token) ? tokenBridgeAddress : cctpTokenMessengerV2Address;

  const {
    data: allowance,
    queryKey,
    refetch,
  } = useReadContract({
    abi: erc20Abi,
    functionName: "allowance",
    args: [address ?? "0x", spender],
    address: token[fromChainLayer] ?? "0x",
    query: {
      enabled: !!token && !isEth(token) && !!address && !!fromChainLayer,
    },
    chainId: fromChainId,
  });

  return { allowance, queryKey, refetchAllowance: refetch };
};

export default useAllowance;
