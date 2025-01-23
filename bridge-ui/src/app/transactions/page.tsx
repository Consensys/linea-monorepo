"use client";

import NoTransaction from "@/components/v2/transaction/no-transaction";
import styles from "./page.module.scss";
import ListTransaction from "@/components/v2/transaction/list-transaction";
import InternalNav from "@/components/v2/internal-nav";
import TopBanner from "@/components/v2/top-banner";
import { useAccount } from "wagmi";
import TransactionNotConnected from "@/components/v2/transaction/transaction-not-connected";
import { TransactionStatus } from "@/components/transactions/TransactionItem";

const listTransaction = [
  {
    code: "0xcBf603bF17dyhxzkD23dE93a2B44",
    value: "1",
    date: "Dec, 18, 2024",
    status: TransactionStatus.COMPLETED,
    unit: "eth",
  },
  {
    code: "0xcBf603bF17dyhxzkD23dE93a2B44",
    value: "1",
    date: "Dec, 18, 2024",
    status: TransactionStatus.READY_TO_CLAIM,
    unit: "eth",
  },
  {
    code: "0xcBf603bF17dyhxzkD23dE93a2B44",
    value: "1",
    date: "Dec, 18, 2024",
    status: TransactionStatus.PENDING,
    unit: "eth",
    estimatedTime: "20 mins",
  },
];

export default function TransactionsPage() {
  const { isConnected } = useAccount();

  return (
    <>
      <TopBanner
        href="/"
        text="⭐️ Stay ahead of the curve with the latest trending tokens - Discover trending tokens"
      />
      <div className={styles["content-wrapper"]}>
        <InternalNav />
        {isConnected ? (
          listTransaction?.length ? (
            <ListTransaction transactions={listTransaction} />
          ) : (
            <NoTransaction />
          )
        ) : (
          <TransactionNotConnected />
        )}
      </div>
    </>
  );
}
