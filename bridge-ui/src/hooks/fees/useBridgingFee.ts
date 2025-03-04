import { useMemo } from "react";
import { Address } from "viem";
import { TokenInfo } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import useFeeData from "./useFeeData";
import useMessageNumber from "../useMessageNumber";
import useERC20BridgingFee from "./useERC20BridgingFee";
import useEthBridgingFee from "./useEthBridgingFee";
import { isEth } from "@/utils/tokens";
import { useFormStore } from "@/stores/formStoreProvider";

type UseBridgingFeeProps = {
  token: TokenInfo;
  account?: Address;
  recipient: Address;
  amount: bigint;
  claimingType: "auto" | "manual";
};

const useBridgingFee = ({ account, token, claimingType, amount, recipient }: UseBridgingFeeProps) => {
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const setBridgingFees = useFormStore((state) => state.setBridgingFees);

  const { feeData } = useFeeData(toChain.id);
  const nextMessageNumber = useMessageNumber({ fromChain, claimingType });

  const eth = useEthBridgingFee({
    account,
    fromChain,
    toChain,
    nextMessageNumber,
    amount,
    recipient,
    token,
    claimingType,
  });

  const erc20 = useERC20BridgingFee({
    account,
    token,
    fromChain,
    toChain,
    nextMessageNumber,
    amount,
    recipient,
    claimingType,
  });

  const isError = eth.isError || erc20.isError;
  const isLoading = eth.isLoading || erc20.isLoading;
  const gasLimit = isEth(token) ? eth.data : erc20.data;

  const bridgingFees = useMemo(() => {
    if (isLoading) {
      return null;
    }

    if (isError) {
      return null;
    }

    if (!gasLimit || !feeData) {
      return null;
    }
    const fees = feeData * (gasLimit + fromChain.gasLimitSurplus) * fromChain.profitMargin;
    setBridgingFees(fees);
    return fees;
  }, [isLoading, isError, gasLimit, feeData, fromChain.gasLimitSurplus, fromChain.profitMargin, setBridgingFees]);

  return bridgingFees;
};

export default useBridgingFee;
