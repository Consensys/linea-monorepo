import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useBlockNumber } from "wagmi";
import { formatBalance } from "@/utils/format";
import { useFormContext } from "react-hook-form";
import { useTokenBalance } from "@/hooks/useTokenBalance";
import styles from "./balance.module.scss";
import { useSelectedToken } from "@/hooks/useSelectedToken";

export function Balance() {
  // Context
  const token = useSelectedToken();

  // Wagmi
  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { balance, queryKey } = useTokenBalance(token);

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

  return (
    <div className={styles.balance}>
      <span>{balance && `${formatBalance(balance)} ${token?.symbol}`} available</span>
    </div>
  );
}
