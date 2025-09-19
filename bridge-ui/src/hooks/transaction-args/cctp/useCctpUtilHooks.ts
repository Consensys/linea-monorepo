// Break pattern of 1 hook-1 file because TypeScript and CI were going nuts on filenames such as useCctpDestinationDomain.ts

import { useMemo } from "react";
import { useChainStore, useFormStore } from "@/stores";
import { getCctpFee } from "@/services/cctp";
import { useQuery } from "@tanstack/react-query";
import { CCTP_MAX_FINALITY_THRESHOLD, CCTP_MIN_FINALITY_THRESHOLD, USDC_DECIMALS } from "@/constants";
import { isUndefined } from "@/utils";
import { CCTPMode } from "@/types";

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
  const cctpMode = useFormStore((state) => state.cctpMode);
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

    const THRESHOLD = cctpMode === CCTPMode.FAST ? CCTP_MAX_FINALITY_THRESHOLD : CCTP_MIN_FINALITY_THRESHOLD;

    let finalityFee = data.find((fee) => fee.finalityThreshold === THRESHOLD)?.minimumFee;

    // TODO: remove this once we have the correct fee for fast mode
    // Temporary using 0.14% fee for fast mode
    if (cctpMode === CCTPMode.FAST) {
      finalityFee = 14;
    }

    if (isUndefined(finalityFee)) {
      return null;
    }

    // Convert BPS (basis points) to multiplier: multiply amount * fee (bps), then divide by 10_000 (1 BPS = 0.01%, 10_000 BPS = 100%)
    return (amount * BigInt(finalityFee)) / 10_000n;
  }, [amount, tokenDecimals, data, cctpMode]);
};
