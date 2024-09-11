import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useBlockNumber } from "wagmi";
import { formatBalance } from "@/utils/format";
import { useFormContext } from "react-hook-form";
import { useChainStore } from "@/stores/chainStore";
import { useTokenBalance } from "@/hooks/useTokenBalance";

export function Balance() {
  // Context
  const { token, networkLayer } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
  }));

  const tokenAddress = token?.[networkLayer];
  // Wagmi
  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { balance, queryKey } = useTokenBalance(tokenAddress, token?.decimals);

  // Form
  const { setValue } = useFormContext();

  useEffect(() => {
    setValue("balance", balance);
  }, [balance, setValue, token?.decimals]);

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
  }, [blockNumber, queryClient, queryKey]);

  return <span className="label-text ml-1">{balance && `Balance: ${formatBalance(balance)} ${token?.symbol}`}</span>;
}
