import { useMemo } from "react";

import { useConnection } from "wagmi";

import { type TransactionRequest, getAdapterById } from "@/adapters";
import { BridgeMessage, Chain, TransactionStatus } from "@/types";
import { isUndefined, isUndefinedOrEmptyString } from "@/utils/misc";

import useTransactionSender from "./useTransactionSender";

type UseClaimProps = {
  status?: TransactionStatus;
  adapterId?: string;
  fromChain?: Chain;
  toChain?: Chain;
  args?: BridgeMessage;
  lstSimulationPassed?: boolean;
};

const useClaim = ({ status, adapterId, fromChain, toChain, args, lstSimulationPassed }: UseClaimProps) => {
  const { address } = useConnection();

  const claimTx = useMemo((): TransactionRequest | undefined => {
    if (
      isUndefinedOrEmptyString(address) ||
      isUndefined(status) ||
      isUndefined(adapterId) ||
      isUndefined(fromChain) ||
      isUndefined(toChain) ||
      isUndefined(args)
    )
      return;

    if (status !== TransactionStatus.READY_TO_CLAIM) return;

    const adapter = getAdapterById(adapterId);
    if (!adapter) return;

    return adapter.buildClaimTx?.({ message: args, fromChain, toChain, lstSimulationPassed }) ?? undefined;
  }, [address, status, adapterId, fromChain, toChain, args, lstSimulationPassed]);

  const { send, ...txState } = useTransactionSender(claimTx);

  return {
    transactionType: "claim" as const,
    claim: send,
    ...txState,
  };
};

export default useClaim;
