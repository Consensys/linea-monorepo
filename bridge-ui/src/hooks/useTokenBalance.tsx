import { useAccount, useBalance, useReadContract } from "wagmi";
import { erc20Abi, formatUnits } from "viem";
import { useChainStore } from "@/stores/chainStore";

export function useTokenBalance(tokenAddress: `0x${string}` | null | undefined, tokenDecimals = 18) {
  const { address } = useAccount();
  const fromChain = useChainStore((state) => state.fromChain);

  const ethBalance = useBalance({
    chainId: fromChain?.id,
    address,
    query: {
      enabled: !!address && !tokenAddress,
    },
  });

  const erc20Balance = useReadContract({
    address: tokenAddress || "0x",
    abi: erc20Abi,
    functionName: "balanceOf",
    args: [address || "0x"],
    chainId: fromChain?.id,
    query: {
      enabled: !!tokenAddress && !!address,
    },
  });

  const isError = ethBalance.isError || erc20Balance.isError;
  const isLoading = ethBalance.isLoading || erc20Balance.isLoading;
  const balance = !tokenAddress ? ethBalance.data?.value : erc20Balance.data;
  const queryKey = ethBalance.queryKey || erc20Balance.queryKey;

  return {
    balance: formatUnits(balance || 0n, tokenDecimals),
    isError,
    isLoading,
    queryKey,
    refetch: () => {
      ethBalance.refetch();
      erc20Balance.refetch();
    },
  };
}
