import { useChainStore } from "@/stores";

export const useCCTPSrcDomain = () => {
  const fromChain = useChainStore.useFromChain();
  return fromChain.cctpDomain;
};

export default useCCTPSrcDomain;
