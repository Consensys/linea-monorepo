// Break pattern of 1 hook-1 file because TypeScript and CI were going nuts on filenames such as useCctpDestinationDomain.ts

import { useChainStore } from "@/stores";
import { getCctpFee } from "@/services/cctp";
import { useQuery } from "@tanstack/react-query";
import { CCTP_MIN_FINALITY_THRESHOLD, USDC_DECIMALS } from "@/constants";
import { isUndefined } from "@/utils";
import { formatUnits, parseUnits } from "viem";
import { useMemo } from "react";

const useCctpSrcDomain = () => {
  const fromChain = useChainStore.useFromChain();
  return fromChain.cctpDomain;
};

export const useCctpDestinationDomain = () => {
  const toChain = useChainStore.useToChain();
  return toChain.cctpDomain;
};

export const useCctpFee = (amount: bigint | null, tokenDecimals: number): bigint | null => {
  const fromChain = useChainStore.useFromChain();
  const srcDomain = useCctpSrcDomain();
  const dstDomain = useCctpDestinationDomain();
  const { data } = useQuery({
    queryKey: ["useCctpFee", srcDomain, dstDomain],
    queryFn: async () => getCctpFee(srcDomain, dstDomain, fromChain.testnet),
    enabled: !!amount && tokenDecimals === USDC_DECIMALS,
  });

  return useMemo(() => {
    if (!amount || tokenDecimals !== USDC_DECIMALS || isUndefined(data)) {
      return null;
    }

    const fastFinalityFee = data.find((fee) => fee.finalityThreshold === CCTP_MIN_FINALITY_THRESHOLD)?.minimumFee;

    if (isUndefined(fastFinalityFee)) {
      return null;
    }

    const feeFraction = fastFinalityFee / 10_000; // Convert BPS to fraction (1 BPS = 0.01% and 10 000 BPS = 100%)
    const formattedAmount = formatUnits(amount, USDC_DECIMALS);
    const rawFee = parseFloat(formattedAmount) * feeFraction;

    return parseUnits(rawFee.toString(), USDC_DECIMALS);
  }, [amount, tokenDecimals, data]);
};
