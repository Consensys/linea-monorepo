import { useQuery } from "@tanstack/react-query";
import { Address } from "viem";
import { fetchTokenPrices } from "@/services/tokenService";
import log from "loglevel";

type UseTokenPrices = {
  data: Record<string, { usd: number }>;
  isLoading: boolean;
  refetch: () => void;
  error: Error | null;
};

export default function useTokenPrices(tokenAddresses: Address[], chainId?: number): UseTokenPrices {
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["tokenPrices", tokenAddresses, chainId],
    queryFn: () => fetchTokenPrices(tokenAddresses, chainId),
    enabled: !!chainId && tokenAddresses.length > 0 && [1, 59144].includes(chainId),
  });

  if (isError) {
    log.error("Error in useTokenPrices", { error });
    return { data: {}, isLoading, refetch, error };
  }

  return { data: data || {}, isLoading, refetch, error: error || null };
}
