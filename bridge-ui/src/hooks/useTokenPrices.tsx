import { useQuery } from "@tanstack/react-query";
import { Address } from "viem";
import log from "loglevel";
import { fetchTokenPrices } from "@/services/tokenService";
import { useConfigStore } from "@/stores";

type UseTokenPrices = {
  data: Record<string, number>;
  isLoading: boolean;
  refetch: () => void;
  error: Error | null;
};

export default function useTokenPrices(tokenAddresses: Address[], chainId?: number): UseTokenPrices {
  const currency = useConfigStore((state) => state.currency);

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["tokenPrices", tokenAddresses, chainId],
    queryFn: () => fetchTokenPrices(tokenAddresses, currency.value, chainId),
    enabled: !!chainId && tokenAddresses.length > 0 && [1, 59144].includes(chainId),
  });

  if (isError) {
    log.error("Error in useTokenPrices", { error });
    return { data: {}, isLoading, refetch, error };
  }

  return { data: data || {}, isLoading, refetch, error: error || null };
}
