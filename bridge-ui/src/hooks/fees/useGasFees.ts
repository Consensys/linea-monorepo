import { useCallback } from "react";
import { Address, encodeFunctionData } from "viem";
import { useEstimateGas } from "wagmi";
import TokenBridge from "@/abis/TokenBridge.json";
import MessageService from "@/abis/MessageService.json";
import { Chain, Token } from "@/types";
import useFeeData from "./useFeeData";
import { isEth } from "@/utils";

type UseGasFeesProps = {
  address?: Address;
  recipient: Address;
  token: Token;
  fromChain: Chain;
  amount: bigint;
  minimumFee: bigint;
};

const useGasFees = ({ address, recipient, amount, token, fromChain, minimumFee }: UseGasFeesProps) => {
  const { feeData } = useFeeData(fromChain.id);

  const isEther = isEth(token);

  const eth = useEstimateGas({
    chainId: fromChain.id,
    account: address,
    value: minimumFee + amount,
    to: fromChain.messageServiceAddress,
    data: encodeFunctionData({
      abi: MessageService.abi,
      functionName: "sendMessage",
      args: [recipient, minimumFee, "0x"],
    }),
    query: {
      enabled: !!isEther && amount > 0n && !!address && !!recipient,
    },
  });

  const erc20 = useEstimateGas({
    chainId: fromChain.id,
    account: address,
    value: minimumFee,
    to: fromChain.tokenBridgeAddress,
    data: encodeFunctionData({
      abi: TokenBridge.abi,
      functionName: "bridgeToken",
      args: [token[fromChain.layer], amount, recipient],
    }),
    query: {
      enabled: !isEther && amount > 0n && !!address && !!recipient,
    },
  });

  const isError = eth.isError || erc20.isError;
  const isLoading = eth.isLoading || erc20.isLoading;
  const gasLimit = isEther ? eth.data : erc20.data;

  const refetch = useCallback(() => {
    if (isEth(token)) {
      eth.refetch();
    } else {
      erc20.refetch();
    }
  }, [token, eth, erc20]);

  if (isLoading) {
    return null;
  }

  return {
    gasFees: gasLimit && feeData ? gasLimit * feeData : null,
    isError,
    isLoading,
    refetch,
  };
};

export default useGasFees;
