import { Address } from "viem";
import { estimateERC20GasFee } from "@/utils/fees";
import { Chain, ChainLayer } from "@/types";
import { TokenInfo } from "@/config";
import { useQuery } from "@tanstack/react-query";
import { isEth } from "@/utils/tokens";
import { BridgeProvider } from "@/config/config";

type UseERC20BridgingFeeProps = {
  account?: Address;
  recipient: Address;
  amount: bigint;
  token: TokenInfo;
  fromChain: Chain;
  toChain: Chain;
  nextMessageNumber: bigint;
  claimingType: "auto" | "manual";
};

const useERC20BridgingFee = ({
  account,
  token,
  fromChain,
  toChain,
  nextMessageNumber,
  recipient,
  amount,
  claimingType,
}: UseERC20BridgingFeeProps) => {
  const queryKey = [
    "erc20BridgingFee",
    account,
    token,
    fromChain.id,
    toChain.id,
    nextMessageNumber?.toString(),
    amount?.toString(),
    recipient,
  ];

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey,
    enabled:
      !!token &&
      !isEth(token) &&
      !!(token.bridgeProvider === BridgeProvider.NATIVE) &&
      !!(fromChain.layer === ChainLayer.L1) &&
      !!account &&
      !!fromChain &&
      !!toChain &&
      !!nextMessageNumber &&
      !!amount &&
      !!recipient &&
      !!(claimingType === "auto"),
    queryFn: async () =>
      await estimateERC20GasFee({
        address: account!,
        recipient,
        token,
        amount,
        nextMessageNumber,
        fromChain,
        toChain,
      }),
  });

  return {
    data,
    isLoading,
    isError,
    refetch,
  };
};

export default useERC20BridgingFee;
