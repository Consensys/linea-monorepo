import { useEffect, useState } from "react";
import { MdCheck } from "react-icons/md";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import classNames from "classnames";
import { toast } from "react-toastify";
import { MessageWithStatus } from "@/hooks/useMessageService";
import { useMessageService, useSwitchNetwork } from "@/hooks";
import { Transaction } from "@/models";
import { TransactionHistory } from "@/models/history";
import { useWaitForTransactionReceipt } from "wagmi";
import { useChainStore } from "@/stores/chainStore";

interface Props {
  message: MessageWithStatus;
  transaction: TransactionHistory;
}

export default function HistoryClaim({ message, transaction }: Props) {
  const [waitingTransaction, setWaitingTransaction] = useState<Transaction | undefined>();
  const [isClaimingLoading, setIsClaimingLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);

  // Is automatic or manual bridging
  const manualBridging = message.fee === "0";

  // Context
  const toChain = useChainStore((state) => state.toChain);

  // Hooks
  const { switchChainById } = useSwitchNetwork(toChain?.id);
  const { writeClaimMessage, isLoading: isTxLoading, transaction: claimTx } = useMessageService();

  // Wagmi
  const {
    isLoading: isWaitingLoading,
    isSuccess: isWaitingSuccess,
    isError: isWaitingError,
  } = useWaitForTransactionReceipt({
    hash: waitingTransaction?.txHash,
    chainId: waitingTransaction?.chainId,
  });

  const BridgingClaimable = ({ enabled = false }) => {
    const claimBusy = isClaimingLoading || isTxLoading || isWaitingLoading || !enabled;
    return (
      <div className="flex flex-row items-center space-x-2">
        <button
          id={enabled ? "claim-funds-btn" : "claim-funds-btn-disabled"}
          onClick={() => !claimBusy && onClaimMessage()}
          className={classNames("btn btn-primary w-38 rounded-full btn-sm mt-1 no-animation mr-2 uppercase", {
            "cursor-wait": claimBusy,
          })}
          type="button"
          disabled={!enabled}
        >
          {claimBusy && <span className="loading loading-spinner loading-xs"></span>}
          Claim funds
        </button>
      </div>
    );
  };

  const BridgingComplete = () => (
    <div className="flex flex-row items-center space-x-1">
      <MdCheck className="text-2xl text-success" />
      <span className="text-xs">Bridging complete</span>
    </div>
  );

  const WaitingForTransaction = ({ loading }: { loading: boolean }) => (
    <div className="flex flex-row items-center space-x-1">
      {loading && <span className="loading loading-spinner mr-1"></span>}
      <span className="text-xs">Please wait, your funds are being bridged</span>
    </div>
  );

  const getClaimStatus = () => {
    if (message.status === OnChainMessageStatus.CLAIMED || isSuccess) {
      return <BridgingComplete />;
    } else if (message.status === OnChainMessageStatus.CLAIMABLE) {
      return <BridgingClaimable enabled={true} />;
    } else {
      if (manualBridging) {
        return (
          <div className="flex flex-row">
            <BridgingClaimable enabled={false} />
            <WaitingForTransaction loading={false} />
          </div>
        );
      }
      return <WaitingForTransaction loading={true} />;
    }
  };

  const onClaimMessage = async () => {
    if (isClaimingLoading) {
      return;
    }

    try {
      setIsClaimingLoading(true);
      await switchChainById(transaction.toChain.id);
      await writeClaimMessage(message, transaction);
      // eslint-disable-next-line no-empty
    } catch (error) {
    } finally {
      setIsClaimingLoading(false);
    }
  };

  useEffect(() => {
    if (claimTx) {
      setWaitingTransaction({
        txHash: claimTx.txHash,
        chainId: transaction.toChain.id,
        name: transaction.toChain.name,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [claimTx]);

  useEffect(() => {
    if (isWaitingSuccess) {
      toast.success(`Funds claimed on ${transaction.toChain.name}.`);
      setIsSuccess(true);
      setWaitingTransaction(undefined);
    }
  }, [isWaitingSuccess, transaction]);

  useEffect(() => {
    if (isWaitingError) {
      toast.error("Funds claiming failed.");
      setWaitingTransaction(undefined);
    }
  }, [isWaitingError]);

  return <div>{getClaimStatus()}</div>;
}
