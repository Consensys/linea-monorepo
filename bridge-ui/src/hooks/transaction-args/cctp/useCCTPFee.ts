import useCCTPDestinationDomain from "./useCCTPDestinationDomain";
import useCCTPSrcDomain from "./useCCTPSrcDomain";
import { getCCTPFee } from "@/services/cctp";
import { useQuery } from "@tanstack/react-query";
import { CCTP_TRANSFER_MAX_FEE_FALLBACK, CCTP_TRANSFER_FEE_BUFFER } from "@/utils/cctp";

export const useCCTPFee = (): bigint => {
  const srcDomain = useCCTPSrcDomain();
  const dstDomain = useCCTPDestinationDomain();
  const { data } = useQuery({
    queryKey: ["useCCTPFee", srcDomain, dstDomain],
    queryFn: async () => getCCTPFee(srcDomain, dstDomain),
  });
  if (!data) return CCTP_TRANSFER_MAX_FEE_FALLBACK;
  return BigInt(data.minimumFee) + CCTP_TRANSFER_FEE_BUFFER;
};

export default useCCTPFee;
