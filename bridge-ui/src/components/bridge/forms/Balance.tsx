"use client";

import { useEffect, useState } from "react";
import { useFormContext } from "react-hook-form";
import { useAccount, useBalance, UseBalanceReturnType, useBlockNumber } from "wagmi";
import classNames from "classnames";
import { useQueryClient } from "@tanstack/react-query";
import { formatBalance } from "@/utils/format";
import { useChainStore } from "@/stores/chainStore";

export default function Balance() {
  const [currentBalance, setCurrentBalance] = useState<UseBalanceReturnType["data"] | undefined>();
  // Context
  const { token, networkLayer, fromChain } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));

  const tokenAddress = token && token[networkLayer] ? token[networkLayer] : undefined;

  // Wagmi
  const { address, isConnected } = useAccount();
  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { data: balance, queryKey } = useBalance({
    address,
    token: tokenAddress ?? undefined,
    chainId: fromChain?.id,
  });

  // Form
  const { setValue } = useFormContext();

  useEffect(() => {
    if (balance) {
      setValue("balance", balance.formatted);
      setCurrentBalance(balance);
    } else {
      setValue("balance", "");
      setCurrentBalance(undefined);
    }
  }, [balance, setValue]);

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber, queryClient]);

  return (
    <div
      className={classNames("", {
        "text-neutral-600": !isConnected,
      })}
    >
      Balance: {formatBalance(currentBalance?.formatted)} {currentBalance?.symbol}
    </div>
  );
}
