import { useMemo, useEffect, useState } from "react";
import { formatEther, parseEther, parseUnits } from "viem";
import { TokenType } from "@/config";
import { useGasEstimation, useApprove, useMinimumFee, useExecutionFee } from "@/hooks";
import { useChainStore } from "@/stores/chainStore";
import { useSelectedToken } from "./useSelectedToken";

type UseReceivedAmountProps = {
  amount: string;
  claim?: string;
  enoughAllowance: boolean;
};

export function useReceivedAmount({ amount, enoughAllowance, claim }: UseReceivedAmountProps) {
  const [estimatedGasFee, setEstimatedGasFee] = useState<bigint>(0n);

  const token = useSelectedToken();
  const fromChain = useChainStore.useFromChain();

  const { minimumFee } = useMinimumFee();
  const { estimateGasBridge } = useGasEstimation();
  const { estimateApprove } = useApprove();

  const executionFee = useExecutionFee({
    token,
    claim,
    networkLayer: fromChain?.layer,
    minimumFee,
  });

  useEffect(() => {
    const estimate = async () => {
      if (!amount || minimumFee === null || !token?.decimals) {
        setEstimatedGasFee(0n);
        return;
      }

      let calculatedGasFee = 0n;
      if (enoughAllowance) {
        calculatedGasFee = (await estimateGasBridge(amount, minimumFee)) || 0n;
      } else {
        calculatedGasFee =
          (await estimateApprove(parseUnits(amount, token.decimals), fromChain?.tokenBridgeAddress || null)) || 0n;
      }

      setEstimatedGasFee(calculatedGasFee);
    };

    estimate();
  }, [
    amount,
    minimumFee,
    enoughAllowance,
    estimateGasBridge,
    estimateApprove,
    token.decimals,
    fromChain?.tokenBridgeAddress,
  ]);

  const totalReceived = useMemo(() => {
    if (!amount || !token?.decimals) {
      return "0";
    }

    if (token.type !== TokenType.ETH) {
      return amount;
    }

    const amountInWei = parseEther(amount);
    const gasFee = estimatedGasFee || BigInt(0);

    return formatEther(amountInWei - executionFee - gasFee);
  }, [amount, token?.decimals, token?.type, executionFee, estimatedGasFee]);

  return {
    totalReceived,
    fees: {
      total: executionFee + estimatedGasFee,
      bridgingFeeInWei: executionFee,
      transactionFeeInWei: estimatedGasFee,
    },
  };
}
