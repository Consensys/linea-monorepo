import { useAccount, useBalance, useReadContract } from "wagmi";
import { erc20Abi } from "viem";
import { useChainStore } from "@/stores";
import { isEth } from "@/utils";
import { Token } from "@/types";

const useTokenBalance = (token: Token) => {
  const { address } = useAccount();
  const { fromChainId, fromChainLayer } = useChainStore((state) => ({
    fromChainId: state.fromChain.id,
    fromChainLayer: state.fromChain.layer,
  }));

  const ethBalance = useBalance({
    chainId: fromChainId,
    address,
    query: {
      enabled: !!address && !!isEth(token),
    },
  });

  const erc20Balance = useReadContract({
    address: token[fromChainLayer] || "0x",
    abi: erc20Abi,
    functionName: "balanceOf",
    args: [address || "0x"],
    chainId: fromChainId,
    query: {
      enabled: !isEth(token) && !!address,
    },
  });

  const isError = ethBalance.isError || erc20Balance.isError;
  const isLoading = ethBalance.isLoading || erc20Balance.isLoading;
  const balance = isEth(token) ? ethBalance.data?.value : erc20Balance.data;
  const queryKey = ethBalance.queryKey || erc20Balance.queryKey;

  return {
    balance: balance || 0n,
    isError,
    isLoading,
    queryKey,
    refetch: () => {
      if (isEth(token)) {
        ethBalance.refetch();
      } else {
        erc20Balance.refetch();
      }
    },
  };
};

export default useTokenBalance;
