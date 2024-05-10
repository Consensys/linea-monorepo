import { useState, useCallback, useContext } from 'react';
import { prepareWriteContract, writeContract, getPublicClient } from '@wagmi/core';
import { Address } from 'viem';
import { useAccount, useFeeData } from 'wagmi';
import log from 'loglevel';

import ERC20Abi from '@/abis/ERC20.json';
import { ChainContext } from '@/contexts/chain.context';

const useApprove = () => {
  const [hash, setHash] = useState<Address | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const context = useContext(ChainContext);
  const { token, networkLayer, fromChain } = context;

  const { address } = useAccount();
  const { data: feeData } = useFeeData({ chainId: fromChain?.id, watch: true });

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
        const config = await prepareWriteContract({
          address: tokenAddress,
          abi: ERC20Abi,
          functionName: 'approve',
          args: [spender, amount],
        });
        const { hash } = await writeContract(config);
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

        const publicClient = getPublicClient({
          chainId: fromChain?.id,
        });

        const estimatedGasFee = await publicClient.estimateContractGas({
          abi: ERC20Abi,
          functionName: 'approve',
          address: token[networkLayer] ?? '0x',
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
