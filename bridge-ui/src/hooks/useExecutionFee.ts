import { useState, useEffect } from "react";
import { useBlockNumber, useEstimateFeesPerGas } from "wagmi";
import { config, NetworkLayer, NetworkType, TokenInfo, TokenType } from "@/config";
import { useQueryClient } from "@tanstack/react-query";
import { useChainStore } from "@/stores/chainStore";

type useExecutionFeeProps = {
  token: TokenInfo | null;
  claim: string | undefined;
  networkLayer: NetworkLayer | undefined;
  networkType: NetworkType;
  minimumFee: bigint;
};

const useExecutionFee = ({ token, claim, networkLayer, networkType, minimumFee }: useExecutionFeeProps) => {
  const [minFees, setMinFees] = useState<bigint>(0n);
  const toChain = useChainStore((state) => state.toChain);
  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { data: feeData, queryKey } = useEstimateFeesPerGas({ chainId: toChain?.id, type: "legacy" });

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber, queryClient]);

  useEffect(() => {
    setMinFees(0n);
    if (!token) return;

    const fee = calculateFee({
      token,
      claim,
      networkLayer,
      networkType,
      minimumFee,
      gasPrice: feeData?.gasPrice,
    });

    if (fee !== undefined) {
      setMinFees(fee);
    }
  }, [claim, networkLayer, token, minimumFee, networkType, feeData?.gasPrice]);

  return minFees;
};

export default useExecutionFee;

const calculateFee = ({
  token,
  claim,
  networkLayer,
  networkType,
  minimumFee,
  gasPrice,
}: {
  token: TokenInfo;
  claim: string | undefined;
  networkLayer: NetworkLayer | undefined;
  networkType: NetworkType;
  minimumFee: bigint;
  gasPrice: bigint | undefined;
}): bigint | undefined => {
  const isETH = token.type === TokenType.ETH;
  const isL1 = networkLayer === NetworkLayer.L1;
  const isL2 = networkLayer === NetworkLayer.L2;
  const isAutoClaim = claim === "auto";
  const isManualClaim = claim === "manual";
  const isERC20orUSDC = token.type === TokenType.ERC20 || token.type === TokenType.USDC;

  // postman fee
  if (isETH && isL1 && isAutoClaim && gasPrice) {
    return calculatePostmanFee(gasPrice, networkType);
  }

  // 0
  if (isETH && isL1 && isManualClaim) {
    return BigInt(0);
  }

  // anti-DDoS fee + postman fee
  if (isETH && isL2 && isAutoClaim && gasPrice) {
    return calculatePostmanFee(gasPrice, networkType) + minimumFee;
  }

  // anti-DDoS fee
  if (isETH && isL2 && isManualClaim) {
    return minimumFee;
  }

  // 0
  if (isERC20orUSDC && isL1) {
    return BigInt(0);
  }

  // anti-DDoS fee
  if (isERC20orUSDC && isL2) {
    return minimumFee;
  }

  return undefined;
};

const calculatePostmanFee = (gasPrice: bigint, networkType: NetworkType) =>
  config.networks[networkType] &&
  gasPrice *
    (config.networks[networkType].gasEstimated + config.networks[networkType].gasLimitSurplus) *
    config.networks[networkType].profitMargin;
