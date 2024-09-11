import { useEffect, useState } from "react";
import classNames from "classnames";
import { toast } from "react-toastify";
import { useSwitchNetwork } from "@/hooks";
import { Transaction } from "@/models";
import { TransactionHistory } from "@/models/history";
import { useWaitForTransactionReceipt } from "wagmi";
import { useChainStore } from "@/stores/chainStore";
import useTransactionManagement, { MessageWithStatus } from "@/hooks/useTransactionManagement";

interface Props {
  message: MessageWithStatus;
  transaction: TransactionHistory;
}

export default function TransactionClaimButton({ message, transaction }: Props) {
  const [waitingTransaction, setWaitingTransaction] = useState<Transaction | undefined>();
  const [isClaimingLoading, setIsClaimingLoading] = useState<boolean>(false);

  // Context
  const toChain = useChainStore((state) => state.toChain);

  // Hooks
  const { switchChainById } = useSwitchNetwork(toChain?.id);
  const { writeClaimMessage, isLoading: isTxLoading, transaction: claimTx } = useTransactionManagement();

  // Wagmi
  const {
    isLoading: isWaitingLoading,
    isSuccess: isWaitingSuccess,
    isError: isWaitingError,
  } = useWaitForTransactionReceipt({
    hash: waitingTransaction?.txHash,
    chainId: waitingTransaction?.chainId,
  });

  const claimBusy = isClaimingLoading || isTxLoading || isWaitingLoading;

  useEffect(() => {
    if (claimTx) {
      setWaitingTransaction({
        txHash: claimTx.txHash,
        chainId: transaction.toChain.id,
        name: transaction.toChain.name,
      });
    }
  }, [claimTx, transaction.toChain.id, transaction.toChain.name]);

  useEffect(() => {
    if (isWaitingSuccess) {
      toast.success(`Funds claimed on ${transaction.toChain.name}.`);
      setWaitingTransaction(undefined);
    }
  }, [isWaitingSuccess, transaction.toChain.name]);

  useEffect(() => {
    if (isWaitingError) {
      toast.error("Funds claiming failed.");
      setWaitingTransaction(undefined);
    }
  }, [isWaitingError]);

  const onClaimMessage = async () => {
    if (isClaimingLoading) {
      return;
    }

    try {
      setIsClaimingLoading(true);
      await switchChainById(transaction.toChain.id);
      await writeClaimMessage(message, transaction);
    } catch (error) {
      toast.error("Failed to claim funds. Please try again.");
    } finally {
      setIsClaimingLoading(false);
    }
  };

  return (
    <button
      id={!claimBusy ? "claim-funds-btn" : "claim-funds-btn-disabled"}
      onClick={() => !claimBusy && onClaimMessage()}
      className={classNames("btn btn-primary w-full rounded-full uppercase", {
        "cursor-wait": claimBusy,
      })}
      type="button"
      disabled={claimBusy}
    >
      {claimBusy && <span className="loading loading-spinner loading-xs"></span>}
      Claim
    </button>
  );
}
