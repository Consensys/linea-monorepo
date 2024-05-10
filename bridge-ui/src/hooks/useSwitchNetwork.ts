import { useState } from 'react';
import { useNetwork } from 'wagmi';
import { switchNetwork } from '@wagmi/core';

interface MetaMaskError extends Error {
  code: number;
  message: string;
}

const useSwitchNetwork = (fromChainId: number | undefined) => {
  const { chain } = useNetwork();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const switchChain = async () => {
    setIsLoading(true);
    if (chain && fromChainId && chain.id !== fromChainId) {
      try {
        await switchNetwork({
          chainId: fromChainId,
        });
      } catch (error) {
        const metaMaskError = error as MetaMaskError;
        setError(metaMaskError.message);
        throw new Error(metaMaskError.message);
      } finally {
        setIsLoading(false);
      }
    }
  };

  const switchChainById = async (toChainId: number) => {
    setIsLoading(true);
    if (chain && fromChainId && toChainId !== chain.id) {
      try {
        await switchNetwork({
          chainId: toChainId,
        });
      } catch (error) {
        const metaMaskError = error as MetaMaskError;
        setError(metaMaskError.message);
        throw new Error(metaMaskError.message);
      } finally {
        setIsLoading(false);
      }
    }
  };

  return { switchChain, error, isLoading, switchChainById };
};

export default useSwitchNetwork;
