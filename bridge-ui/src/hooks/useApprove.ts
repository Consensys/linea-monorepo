import { useState, useCallback, useEffect } from "react";
import { writeContract, getPublicClient, simulateContract } from "@wagmi/core";
import { Address } from "viem";
import { useAccount, useBlockNumber, useEstimateFeesPerGas } from "wagmi";
import log from "loglevel";
import ERC20Abi from "@/abis/ERC20.json";
import { wagmiConfig } from "@/config";
import { useQueryClient } from "@tanstack/react-query";
import { useChainStore } from "@/stores/chainStore";

const useApprove = () => {
  const [hash, setHash] = useState<Address | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const { token, networkLayer, fromChain } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));

  const { address } = useAccount();

  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { data: feeData, queryKey } = useEstimateFeesPerGas({ chainId: fromChain?.id, type: "legacy" });

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber, queryClient]);

  const writeApprove = useCallback(
    async (amount: bigint, spender: Address | null) => {
      setError(null);
      setIsLoading(true);
      if (!amount) {
        setIsLoading(false);
        return;
      }

      if (!token) {
        return;
      }

      const tokenAddress = token[networkLayer];
      if (!tokenAddress) {
        return;
      }

      try {
        const { request } = await simulateContract(wagmiConfig, {
          address: tokenAddress,
          abi: ERC20Abi,
          functionName: "approve",
          args: [spender, amount],
        });

        const hash = await writeContract(wagmiConfig, request);
        setHash(hash);
      } catch (error) {
        log.error(error);
        setError(error as Error);
      }

      setIsLoading(false);
    },
    [token, networkLayer],
  );

  const estimateApprove = useCallback(
    async (_amount: bigint, _spender: Address | null) => {
      if (!token || !address) {
        return;
      }
      try {
        if (!feeData?.gasPrice) {
          return;
        }

        const publicClient = getPublicClient(wagmiConfig, {
          // eslint-disable-next-line @typescript-eslint/ban-ts-comment
          //@ts-ignore
          chainId: fromChain?.id,
        });

        if (!publicClient) {
          return;
        }

        const estimatedGasFee = await publicClient.estimateContractGas({
          abi: ERC20Abi,
          functionName: "approve",
          address: token[networkLayer] ?? "0x",
          args: [_spender, _amount],
          account: address,
        });

        return estimatedGasFee * feeData.gasPrice;
      } catch (error) {
        log.error(error);
        return;
      }
    },
    [address, token, feeData, networkLayer, fromChain],
  );

  return {
    hash,
    isLoading,
    setHash,
    writeApprove,
    isError: error !== null,
    error,
    estimateApprove,
  };
};

export default useApprove;
