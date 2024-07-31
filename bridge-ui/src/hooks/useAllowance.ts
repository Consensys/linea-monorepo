import { useState, useEffect, useCallback } from "react";
import { readContract } from "@wagmi/core";
import { useAccount } from "wagmi";
import { Address } from "viem";
import ERC20Abi from "@/abis/ERC20.json";
import log from "loglevel";
import { wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";

type UseAllowance = {
  allowance: bigint | null;
  fetchAllowance: () => Promise<void>;
};

const useAllowance = (): UseAllowance => {
  const [allowance, setAllowance] = useState<bigint | null>(null);

  // Wagmi
  const { address } = useAccount();

  // Context
  const { token, networkLayer, tokenBridgeAddress, fromChain } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
    tokenBridgeAddress: state.tokenBridgeAddress,
    fromChain: state.fromChain,
  }));

  const fetchAllowance = useCallback(async () => {
    if (!address || !token || !networkLayer || !token[networkLayer]) {
      return;
    }

    // Here we need to specify the chain because we want to be able
    // to read a contract on both chains without having to connect
    // to one or the other
    try {
      const allowance = (await readContract(wagmiConfig, {
        address: token[networkLayer] as Address,
        abi: ERC20Abi,
        functionName: "allowance",
        args: [address, tokenBridgeAddress],
        chainId: fromChain?.id,
      })) as bigint;

      setAllowance(allowance);
    } catch (error) {
      console.log(error);
      log.error("Unable to fetch allowance", { address });
    }
  }, [address, tokenBridgeAddress, token, networkLayer, fromChain]);

  useEffect(() => {
    fetchAllowance();
  }, [fetchAllowance]);

  return { allowance, fetchAllowance };
};

export default useAllowance;
