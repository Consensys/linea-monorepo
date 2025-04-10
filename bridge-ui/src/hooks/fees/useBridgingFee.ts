import { useMemo, useEffect } from "react";
import { Address } from "viem";
import useFeeData from "./useFeeData";
import useMessageNumber from "../useMessageNumber";
import useERC20BridgingGasUsed from "./useERC20BridgingGasUsed";
import useEthBridgingGasUsed from "./useEthBridgingGasUsed";
import { useFormStore, useChainStore } from "@/stores";
import { Token, ClaimType } from "@/types";
import { isEth, isUndefined } from "@/utils";
import { DEFAULT_ADDRESS_FOR_NON_CONNECTED_USER } from "@/constants";

type UseBridgingFeeProps = {
  isConnected: boolean;
  token: Token;
  account?: Address;
  recipient: Address;
  amount: bigint;
  claimingType: ClaimType;
};

const useBridgingFee = ({ isConnected, account, token, claimingType, amount, recipient }: UseBridgingFeeProps) => {
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const setBridgingFees = useFormStore((state) => state.setBridgingFees);

  const { feeData } = useFeeData(toChain.id);
  const nextMessageNumber = useMessageNumber({ fromChain, claimingType });

  const fromAddress = isConnected ? account : DEFAULT_ADDRESS_FOR_NON_CONNECTED_USER;
  const toAddress = isConnected ? recipient : DEFAULT_ADDRESS_FOR_NON_CONNECTED_USER;

  const eth = useEthBridgingGasUsed({
    account: fromAddress,
    fromChain,
    toChain,
    nextMessageNumber,
    amount,
    recipient: toAddress,
    token,
    claimingType,
  });

  const erc20 = useERC20BridgingGasUsed({
    account: fromAddress,
    token,
    fromChain,
    toChain,
    nextMessageNumber,
    amount,
    recipient: toAddress,
    claimingType,
  });

  const isError = eth.isError || erc20.isError;
  const isLoading = eth.isLoading || erc20.isLoading;
  const gasLimit = isEth(token) ? eth.data : erc20.data;

  const computedBridgingFees = useMemo(() => {
    if (claimingType === ClaimType.MANUAL) {
      return 0n;
    }
    if (isLoading || isError || isUndefined(gasLimit) || isUndefined(feeData)) {
      return null;
    }
    return feeData * (gasLimit + fromChain.gasLimitSurplus) * fromChain.profitMargin;
  }, [isLoading, isError, gasLimit, feeData, claimingType, fromChain.gasLimitSurplus, fromChain.profitMargin]);

  useEffect(() => {
    if (computedBridgingFees !== null) {
      setBridgingFees(computedBridgingFees);
    }
  }, [computedBridgingFees, setBridgingFees]);

  return { bridgingFees: computedBridgingFees, isLoading };
};

export default useBridgingFee;
