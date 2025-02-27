import { useTokenStore } from "@/stores/tokenStoreProvider";

export const useSelectedToken = () => {
  return useTokenStore((state) => state.selectedToken);
};
