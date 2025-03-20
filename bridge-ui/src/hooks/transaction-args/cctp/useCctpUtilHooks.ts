// Break pattern of 1 hook-1 file because TypeScript and CI were going nuts on filenames such as useCCTPDestinationDomain.ts

import { useChainStore } from "@/stores";
import { getCCTPFee } from "@/services/cctp";
import { useQuery } from "@tanstack/react-query";
import { CCTP_TRANSFER_MAX_FEE_FALLBACK } from "@/utils/cctp";

export const useCCTPSrcDomain = () => {
  const fromChain = useChainStore.useFromChain();
  return fromChain.cctpDomain;
};

export const useCCTPDestinationDomain = () => {
  const toChain = useChainStore.useToChain();
  return toChain.cctpDomain;
};

export const useCCTPFee = (): bigint => {
  const srcDomain = useCCTPSrcDomain();
  const dstDomain = useCCTPDestinationDomain();
  const { data } = useQuery({
    queryKey: ["useCCTPFee", srcDomain, dstDomain],
    queryFn: async () => getCCTPFee(srcDomain, dstDomain),
  });
  if (!data) return CCTP_TRANSFER_MAX_FEE_FALLBACK;
  return BigInt(data.minimumFee);
};
