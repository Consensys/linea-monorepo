import { useChainStore } from "@/stores";

export const useCCTPDestinationDomain = () => {
  const toChain = useChainStore.useToChain();
  return toChain.cctpDomain;
};

export default useCCTPDestinationDomain;
