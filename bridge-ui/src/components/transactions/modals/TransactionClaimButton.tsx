import { useContext, useEffect, useState } from "react";
import { toast } from "react-toastify";
import { useSwitchNetwork } from "@/hooks";
import { Transaction } from "@/models";
import { TransactionHistory } from "@/models/history";
import { useWaitForTransactionReceipt } from "wagmi";
import { useChainStore } from "@/stores/chainStore";
import Button from "@/components/bridge/Button";
import useClaimTransaction, { MessageWithStatus } from "@/hooks/useClaimTransaction";
import { ModalContext } from "@/contexts/modal.context";
import TransactionConfirmationModal from "@/components/bridge/modals/TransactionConfirmationModal";

interface Props {
  message: MessageWithStatus;
  transaction: TransactionHistory;
  handleClose: () => void;
}

export default function TransactionClaimButton({ message, transaction, handleClose }: Props) {
  const [waitingTransaction, setWaitingTransaction] = useState<Transaction | undefined>();
  const [isClaimingLoading, setIsClaimingLoading] = useState<boolean>(false);

  // Context
  const toChain = useChainStore((state) => state.toChain);
  const { handleShow: handleShowConfirmationModal, handleClose: handleCloseConfirmationModal } =
    useContext(ModalContext);
  // Hooks
  const { switchChainById } = useSwitchNetwork(toChain?.id);
  const { writeClaimMessage, isLoading: isTxLoading, transaction: claimTx } = useClaimTransaction();

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
      handleClose();
      handleShowConfirmationModal(<TransactionConfirmationModal handleClose={handleCloseConfirmationModal} />, {
        showCloseButton: true,
      });
      setWaitingTransaction(undefined);
    }
  }, [
    handleClose,
    handleCloseConfirmationModal,
    handleShowConfirmationModal,
    isWaitingSuccess,
    transaction.toChain.name,
  ]);

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
    <Button
      id={!claimBusy ? "claim-funds-btn" : "claim-funds-btn-disabled"}
      onClick={() => !claimBusy && onClaimMessage()}
      variant="primary"
      loading={claimBusy}
      type="button"
      disabled={claimBusy}
    >
      Claim
    </Button>
  );
}
