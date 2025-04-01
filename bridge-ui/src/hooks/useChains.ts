import { useChainStore } from "@/stores";

const useAvailableChains = () => {
  return useChainStore((state) => state.availableChains);
};

export default useAvailableChains;
