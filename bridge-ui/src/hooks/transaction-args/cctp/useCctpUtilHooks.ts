// Break pattern of 1 hook-1 file because TypeScript and CI were going nuts on filenames such as useCctpDestinationDomain.ts

import { useChainStore } from "@/stores";
import { getCctpFee } from "@/services/cctp";
import { useQuery } from "@tanstack/react-query";
import { CCTP_TRANSFER_MAX_FEE_FALLBACK } from "@/constants";

const useCctpSrcDomain = () => {
  return useChainStore((state) => state.fromChain.cctpDomain);
};

export const useCctpDestinationDomain = () => {
  return useChainStore((state) => state.toChain.cctpDomain);
};

export const useCctpFee = (): bigint => {
  const isFromChainTestnet = useChainStore((state) => state.fromChain.testnet);
  const srcDomain = useCctpSrcDomain();
  const dstDomain = useCctpDestinationDomain();
  const { data } = useQuery({
    queryKey: ["useCctpFee", srcDomain, dstDomain],
    queryFn: async () => getCctpFee(srcDomain, dstDomain, isFromChainTestnet),
  });
  if (!data) return CCTP_TRANSFER_MAX_FEE_FALLBACK;
  return BigInt(data.minimumFee);
};
