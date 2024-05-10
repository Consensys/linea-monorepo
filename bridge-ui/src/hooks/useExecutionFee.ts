import { useState, useEffect, useContext } from 'react';
import { useFeeData } from 'wagmi';
import { formatEther } from 'viem';

import { config, TokenInfo, TokenType } from '@/config/config';
import { ChainContext, NetworkLayer, NetworkType } from '@/contexts/chain.context';

type useExecutionFeeProps = {
  token: TokenInfo | null;
  claim: string | undefined;
  networkLayer: NetworkLayer | undefined;
  networkType: NetworkType;
  minimumFee: bigint;
};

const useExecutionFee = ({ token, claim, networkLayer, networkType, minimumFee }: useExecutionFeeProps) => {
  const [minFees, setMinFees] = useState<string>();
  const context = useContext(ChainContext);
  const { toChain } = context;
  const { data: feeData } = useFeeData({ chainId: toChain?.id, watch: true });
  useEffect(() => {
    setMinFees(undefined);
    if (!token) {
      return;
    }

    const isETH = token.type === TokenType.ETH;
    const isL1 = networkLayer === NetworkLayer.L1;
    const isL2 = networkLayer === NetworkLayer.L2;
    const isAutoClaim = claim === 'auto';
    const isManualClaim = claim === 'manual';
    const isERC20orUSDC = token.type === TokenType.ERC20 || token.type === TokenType.USDC;
    // postman fee
    if (isETH && isL1 && isAutoClaim && feeData?.gasPrice) {
      const postmanFee = calculatePostmanFee(feeData.gasPrice, networkType);
      postmanFee && setMinFees(formatEther(postmanFee));
      return;
    }

    // 0
    if (isETH && isL1 && isManualClaim) {
      setMinFees(formatEther(BigInt(0)));
      return;
    }

    // anti-DDoS fee + postman fee
    if (isETH && isL2 && isAutoClaim && feeData?.gasPrice) {
      const postmanFee = calculatePostmanFee(feeData.gasPrice, networkType);
      postmanFee && setMinFees(formatEther(postmanFee + minimumFee));
      return;
    }

    // anti-DDoS fee
    if (isETH && isL2 && isManualClaim) {
      setMinFees(formatEther(minimumFee));
      return;
    }

    // 0
    if (isERC20orUSDC && isL1) {
      setMinFees(formatEther(BigInt(0)));
      return;
    }

    // anti-DDoS fee
    if (isERC20orUSDC && isL2) {
      setMinFees(formatEther(minimumFee));
      return;
    }
  }, [claim, networkLayer, token, minimumFee, feeData, networkType]);

  const calculatePostmanFee = (gasPrice: bigint, networkType: NetworkType) =>
    config.networks[networkType] &&
    gasPrice *
      (config.networks[networkType].gasEstimated + config.networks[networkType].gasLimitSurplus) *
      config.networks[networkType].profitMargin;

  return minFees;
};

export default useExecutionFee;
