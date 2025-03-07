import { useEffect, useMemo } from "react";
import Link from "next/link";
import { useAccount, useTransactionReceipt } from "wagmi";
import { formatEther, zeroAddress } from "viem";
import { formatDate, fromUnixTime } from "date-fns";
import Modal from "@/components/v2/modal";
import styles from "./transaction-details.module.scss";
import Button from "@/components/v2/ui/button";
import { formatBalance, formatHex } from "@/utils/format";
import ArrowRightIcon from "@/assets/icons/arrow-right.svg";
import { BridgeTransaction } from "@/utils/history";
import { TransactionStatus } from "@/components/transactions/TransactionItem";
import useTokenPrices from "@/hooks/useTokenPrices";
import { useConfigStore } from "@/stores/configStore";
import useClaim from "@/hooks/useClaim";
import { useQueryClient } from "@tanstack/react-query";
import { useChainStore } from "@/stores/chainStore";

type Props = {
  transaction: BridgeTransaction | undefined;
  isModalOpen: boolean;
  onCloseModal: () => void;
};

export default function TransactionDetails({ transaction, isModalOpen, onCloseModal }: Props) {
  const { address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const currency = useConfigStore((state) => state.currency);
  const formattedTimestamp = transaction?.timestamp ? fromUnixTime(Number(transaction.timestamp)) : "";
  const formattedDate = transaction?.timestamp ? formatDate(formattedTimestamp, "MMM, dd, yyyy") : "";
  const formattedTime = transaction?.timestamp ? formatDate(formattedTimestamp, "ppp") : "";

  const { data: tokensPrices } = useTokenPrices([zeroAddress], transaction?.fromChain.id);

  const queryClient = useQueryClient();
  const { claim, isConfirming, isPending, isConfirmed } = useClaim({
    status: transaction?.status,
    fromChain: transaction?.fromChain,
    toChain: transaction?.toChain,
    args: {
      from: transaction?.message?.from,
      to: transaction?.message?.to,
      fee: transaction?.message?.fee,
      value: transaction?.message?.value,
      nonce: transaction?.message?.nonce,
      calldata: transaction?.message?.calldata,
      messageHash: transaction?.message?.messageHash,
      proof: transaction?.message?.proof,
    },
  });

  useEffect(() => {
    if (isConfirmed) {
      queryClient.invalidateQueries({ queryKey: ["transactionHistory", address, fromChain.id, toChain.id] });
      onCloseModal();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isConfirmed]);

  const { data: initialTransactionReceipt } = useTransactionReceipt({
    hash: transaction?.bridgingTx as `0x${string}`,
    chainId: transaction?.fromChain.id,
    query: {
      enabled: !!transaction?.bridgingTx && transaction?.status === TransactionStatus.COMPLETED,
    },
  });

  const { data: claimingTransactionReceipt } = useTransactionReceipt({
    hash: transaction?.claimingTx as `0x${string}`,
    chainId: transaction?.toChain.id,
    query: {
      enabled: !!transaction?.claimingTx && transaction?.status === TransactionStatus.COMPLETED,
    },
  });

  const gasFees = useMemo(() => {
    const initialTransactionFee =
      initialTransactionReceipt?.gasUsed && initialTransactionReceipt?.effectiveGasPrice
        ? initialTransactionReceipt.gasUsed * initialTransactionReceipt.effectiveGasPrice
        : 0n;

    const claimingTransactionFee =
      claimingTransactionReceipt?.gasUsed && claimingTransactionReceipt?.effectiveGasPrice
        ? claimingTransactionReceipt.gasUsed * claimingTransactionReceipt.effectiveGasPrice
        : 0n;

    return initialTransactionFee + claimingTransactionFee;
  }, [initialTransactionReceipt, claimingTransactionReceipt]);

  const handleClaim = () => {
    if (claim) {
      claim();
    }
  };
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
                href={`${transaction?.fromChain.blockExplorers?.default.url}/tx/${transaction?.bridgingTx}` || ""}
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
              {transaction?.claimingTx ? (
                <Link
                  href={`${transaction?.toChain.blockExplorers?.default.url}/tx/${transaction.claimingTx}` || ""}
                  target="_blank"
                  rel="noopenner noreferrer"
                >
                  {formatHex(transaction.claimingTx)}
                </Link>
              ) : (
                <span>N/A</span>
              )}

              <ArrowRightIcon />
            </div>
          </li>
          {transaction?.status === TransactionStatus.COMPLETED && (
            <li>
              <span>Gas fee</span>
              <span className={styles.price}>
                {tokensPrices[zeroAddress] ? (
                  <span>
                    {(tokensPrices[zeroAddress] * Number(gasFees)).toLocaleString("en-US", {
                      style: "currency",
                      currency: currency.label,
                      maximumFractionDigits: 4,
                    })}
                  </span>
                ) : (
                  `${formatBalance(formatEther(gasFees), 8)} ETH`
                )}
              </span>
            </li>
          )}
        </ul>
        {transaction?.status === TransactionStatus.READY_TO_CLAIM && (
          <Button disabled={isPending || isConfirming} onClick={handleClaim} fullWidth>
            {isPending || isConfirming ? "Waiting for confirmation..." : "Claim"}
          </Button>
        )}
      </div>
    </Modal>
  );
}
