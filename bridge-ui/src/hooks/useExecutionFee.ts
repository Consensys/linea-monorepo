import { useState, useEffect, useCallback } from "react";
import { useBlockNumber, useEstimateFeesPerGas } from "wagmi";
import { NetworkLayer, TokenInfo, TokenType } from "@/config";
import { useQueryClient } from "@tanstack/react-query";
import { useChainStore } from "@/stores/chainStore";
import usePostmanFee from "./usePostmanFee";
import { useFormContext } from "react-hook-form";

type useExecutionFeeProps = {
  token: TokenInfo | null;
  claim: string | undefined;
  networkLayer: NetworkLayer;
  minimumFee: bigint;
};

const useExecutionFee = ({ token, claim, networkLayer, minimumFee }: useExecutionFeeProps) => {
  const [minFees, setMinFees] = useState<bigint>(0n);
  const toChain = useChainStore((state) => state.toChain);
  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { data: feeData, queryKey } = useEstimateFeesPerGas({ chainId: toChain?.id, type: "legacy" });
  const { calculatePostmanFee } = usePostmanFee({ currentLayer: networkLayer, claimingType: claim });

  const { watch } = useFormContext();

  const [amount, recipient] = watch(["amount", "recipient"]);

  const calculateFee = useCallback(
    async ({
      token,
      claim,
      networkLayer,
      minimumFee,
      gasPrice,
    }: {
      token: TokenInfo;
      claim: string | undefined;
      networkLayer: NetworkLayer | undefined;
      minimumFee: bigint;
      gasPrice: bigint | undefined;
    }): Promise<bigint | undefined> => {
      const isETH = token.type === TokenType.ETH;
      const isL1 = networkLayer === NetworkLayer.L1;
      const isL2 = networkLayer === NetworkLayer.L2;
      const isAutoClaim = claim === "auto";
      const isManualClaim = claim === "manual";
      const isERC20orUSDC = token.type === TokenType.ERC20 || token.type === TokenType.USDC;

      // postman fee
      if (isETH && isL1 && isAutoClaim && gasPrice) {
        return calculatePostmanFee(amount, recipient);
      }

      // 0
      if (isETH && isL1 && isManualClaim) {
        return BigInt(0);
      }

      // anti-DDoS fee + postman fee
      if (isETH && isL2 && isAutoClaim && gasPrice) {
        const postmanFee = await calculatePostmanFee(amount, recipient);
        return postmanFee + minimumFee;
      }

      // anti-DDoS fee
      if (isETH && isL2 && isManualClaim) {
        return minimumFee;
      }

      // Postman fee
      if (isERC20orUSDC && isL1 && isAutoClaim && gasPrice) {
        return calculatePostmanFee(amount, recipient);
      }

      // 0
      if (isERC20orUSDC && isL1 && isManualClaim) {
        return BigInt(0);
      }

      // anti-DDoS fee
      if (isERC20orUSDC && isL2) {
        return minimumFee;
      }

      return undefined;
    },
    [amount, calculatePostmanFee, recipient],
  );

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber, queryClient]);

  useEffect(() => {
    setMinFees(0n);
    if (!token) return;

    async function calculateExecutionFee(token: TokenInfo) {
      const fee = await calculateFee({
        token,
        claim,
        networkLayer,
        minimumFee,
        gasPrice: feeData?.gasPrice,
      });

      if (fee !== undefined) {
        setMinFees(fee);
      }
    }

    calculateExecutionFee(token);
  }, [claim, networkLayer, token, minimumFee, feeData?.gasPrice, calculateFee]);

  return minFees;
};

export default useExecutionFee;
