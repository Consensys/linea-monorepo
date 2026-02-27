import { useBalance } from "wagmi";

import { Chain, ChainLayer } from "@/types";

type UseL1MessageServiceLiquidityProps = {
  toChain: Chain | undefined;
  isL2ToL1Eth: boolean;
  withdrawalAmount: bigint;
};

type UseL1MessageServiceLiquidityResult = {
  balance: bigint | undefined;
  isLowLiquidity: boolean;
  isLoading: boolean;
};

const useL1MessageServiceLiquidity = ({
  toChain,
  isL2ToL1Eth,
  withdrawalAmount,
}: UseL1MessageServiceLiquidityProps): UseL1MessageServiceLiquidityResult => {
  const enabled = isL2ToL1Eth && toChain?.layer === ChainLayer.L1 && !!toChain.yieldProviderAddress;

  const { data: balanceData, isLoading } = useBalance({
    chainId: toChain?.id,
    address: toChain?.messageServiceAddress,
    query: {
      enabled,
    },
  });

  const balance = balanceData?.value;

  const isLowLiquidity =
    enabled && !isLoading && balance !== undefined && withdrawalAmount > 0n && withdrawalAmount > balance;

  return { balance, isLowLiquidity, isLoading };
};

export default useL1MessageServiceLiquidity;
