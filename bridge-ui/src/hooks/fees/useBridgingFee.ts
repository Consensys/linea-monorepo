import { Address } from "viem";
import { TokenInfo } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import useFeeData from "./useFeeData";
import useMessageNumber from "../useMessageNumber";
import useERC20BridgingFee from "./useERC20BridgingFee";
import useEthBridgingFee from "./useEthBridgingFee";
import { isEth } from "@/utils/tokens";

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

  if (isLoading) {
    return null;
  }

  if (isError) {
    return null;
  }

  return gasLimit && feeData ? feeData * (gasLimit + fromChain.gasLimitSurplus) * fromChain.profitMargin : null;
};

export default useBridgingFee;
