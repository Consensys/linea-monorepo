import { useState } from "react";
import { switchChain } from "@wagmi/core";
import { useAccount } from "wagmi";
import { wagmiConfig } from "@/config";

interface MetaMaskError extends Error {
  code: number;
  message: string;
}

const useSwitchNetwork = (fromChainId: number | undefined) => {
  const { chain } = useAccount();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const switchNetwork = async () => {
    setIsLoading(true);
    if (chain && fromChainId && chain.id !== fromChainId) {
      try {
        await switchChain(wagmiConfig, {
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
        await switchChain(wagmiConfig, {
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

  return { switchChain: switchNetwork, error, isLoading, switchChainById };
};

export default useSwitchNetwork;
