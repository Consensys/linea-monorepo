import { useMemo } from "react";

import Link from "next/link";
import { formatEther } from "viem";
import { useTransactionReceipt } from "wagmi";

import ArrowRightIcon from "@/assets/icons/arrow-right.svg";
import Modal from "@/components/modal";
import { useClaimingTx, useBridgeTransactionMessage } from "@/hooks";
import { BridgeTransaction, TransactionStatus } from "@/types";
import { formatBalance, formatHex, formatTimestamp } from "@/utils/format";

import ClaimActions from "./claim-actions";
import styles from "./transaction-details.module.scss";

type Props = {
  transaction: BridgeTransaction | undefined;
  isModalOpen: boolean;
  onCloseModal: () => void;
};

export default function TransactionDetails({ transaction, isModalOpen, onCloseModal }: Props) {
  const formattedDate = transaction?.timestamp ? formatTimestamp(Number(transaction.timestamp), "MMM, dd, yyyy") : "";
  const formattedTime = transaction?.timestamp ? formatTimestamp(Number(transaction.timestamp), "ppp") : "";

  const { message, isLoading: isLoadingClaimTxParams } = useBridgeTransactionMessage(transaction);
  const claimingTx = useClaimingTx(transaction);

  const hydratedTransaction = useMemo(() => {
    if (!transaction) return undefined;
    return {
      ...transaction,
      ...(message ? { message } : {}),
      ...(claimingTx && !transaction.claimingTx ? { claimingTx } : {}),
    };
  }, [transaction, message, claimingTx]);

  const displayClaimingTx = hydratedTransaction?.claimingTx;

  const { data: initialTransactionReceipt } = useTransactionReceipt({
    hash: transaction?.bridgingTx as `0x${string}`,
    chainId: transaction?.fromChain.id,
    query: {
      enabled: !!transaction?.bridgingTx && transaction?.status === TransactionStatus.COMPLETED,
      staleTime: 1000 * 60 * 5,
      refetchOnMount: false,
    },
  });

  const { data: claimingTransactionReceipt } = useTransactionReceipt({
    hash: displayClaimingTx as `0x${string}`,
    chainId: transaction?.toChain.id,
    query: {
      enabled: !!displayClaimingTx && transaction?.status === TransactionStatus.COMPLETED,
      staleTime: 1000 * 60 * 5,
      refetchOnMount: false,
    },
  });

  const gasFees = useMemo(() => {
    const initialTransactionFee =
      initialTransactionReceipt &&
      "gasUsed" in initialTransactionReceipt &&
      "effectiveGasPrice" in initialTransactionReceipt &&
      initialTransactionReceipt.gasUsed &&
      initialTransactionReceipt.effectiveGasPrice
        ? (initialTransactionReceipt.gasUsed as bigint) * (initialTransactionReceipt.effectiveGasPrice as bigint)
        : 0n;

    const claimingTransactionFee =
      claimingTransactionReceipt &&
      "gasUsed" in claimingTransactionReceipt &&
      "effectiveGasPrice" in claimingTransactionReceipt &&
      claimingTransactionReceipt.gasUsed &&
      claimingTransactionReceipt.effectiveGasPrice
        ? (claimingTransactionReceipt.gasUsed as bigint) * (claimingTransactionReceipt.effectiveGasPrice as bigint)
        : 0n;

    return initialTransactionFee + claimingTransactionFee;
  }, [initialTransactionReceipt, claimingTransactionReceipt]);

  return (
    <Modal title="Transaction details" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={styles["modal-inner"]}>
        <ul className={styles.list}>
          <li>
            <span>Timestamp</span>
            <div className={styles["date-time"]}>
              <span>{formattedDate}</span>
              <span>{formattedTime}</span>
            </div>
          </li>
          <li>
            <span>{transaction?.fromChain.name} Tx hash</span>
            <div className={styles.hash}>
              <Link
                href={`${transaction?.fromChain.blockExplorers?.default.url}/tx/${transaction?.bridgingTx}`}
                target="_blank"
                rel="noopenner noreferrer"
              >
                {formatHex(transaction?.bridgingTx)}
              </Link>
              <ArrowRightIcon />
            </div>
          </li>
          <li>
            <span>{transaction?.toChain.name} Tx hash</span>
            <div className={styles.hash}>
              {displayClaimingTx ? (
                <Link
                  href={`${transaction?.toChain.blockExplorers?.default.url}/tx/${displayClaimingTx}`}
                  target="_blank"
                  rel="noopenner noreferrer"
                >
                  {formatHex(displayClaimingTx)}
                </Link>
              ) : (
                <span>Pending</span>
              )}

              <ArrowRightIcon />
            </div>
          </li>
          {transaction?.status === TransactionStatus.COMPLETED && (
            <li>
              <span>Gas fee</span>
              <span className={styles.price}>{formatBalance(formatEther(gasFees), 8)} ETH</span>
            </li>
          )}
        </ul>
        {hydratedTransaction?.status === TransactionStatus.READY_TO_CLAIM && (
          <ClaimActions
            transaction={hydratedTransaction}
            isLoadingClaimTxParams={isLoadingClaimTxParams}
            onCloseModal={onCloseModal}
          />
        )}
      </div>
    </Modal>
  );
}
