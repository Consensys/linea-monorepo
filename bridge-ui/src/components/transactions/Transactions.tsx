"use client";

import { useEffect, useMemo } from "react";
import { useBlockNumber } from "wagmi";
import TransactionItem from "./TransactionItem";
import { TransactionHistory } from "@/models/history";
import { formatDate, fromUnixTime } from "date-fns";
import { NoTransactions } from "./NoTransaction";
import { useFetchHistory } from "@/hooks";
import { Button } from "../ui";

const groupByDay = (transactions: TransactionHistory[]): Record<string, TransactionHistory[]> => {
  return transactions.reduce(
    (acc, transaction) => {
      const date = formatDate(fromUnixTime(Number(transaction.timestamp)), "yyyy-MM-dd");
      if (!acc[date]) {
        acc[date] = [];
      }
      acc[date].push(transaction);
      return acc;
    },
    {} as Record<string, TransactionHistory[]>,
  );
};

function SkeletonLoader() {
  return (
    <div className="flex flex-col gap-8 border-2 border-card bg-cardBg p-4">
      {Array.from({ length: 3 }).map((_, groupIndex) => (
        <div key={groupIndex} className="flex flex-col gap-4">
          <div className="skeleton h-6 w-1/3 bg-card"></div>
          {Array.from({ length: 2 }).map((_, itemIndex) => (
            <div
              key={itemIndex}
              className="grid grid-cols-1 items-center gap-0 rounded-lg bg-[#2D2D2D] p-4 text-[#C0C0C0] sm:grid-cols-1 md:grid-cols-6 md:gap-4"
            >
              <div className="grid grid-cols-2 gap-4 border-b border-card py-4 md:col-span-2 md:border-none md:p-0">
                <div className="skeleton h-4 w-1/2 bg-card"></div>
                <div className="skeleton h-4 w-1/2 bg-card"></div>
              </div>
              <div className="hidden px-6 md:col-span-2 md:block md:border-x md:border-card">
                <div className="skeleton h-4 w-full bg-card"></div>
              </div>
              <div className="grid grid-cols-2 items-center gap-4 border-b border-card py-4 md:col-span-2 md:border-none md:p-0">
                <div className="skeleton h-4 w-1/2 bg-card"></div>
                <div className="skeleton h-4 w-1/2 bg-card"></div>
              </div>
              <div className="px-6 pt-4 md:hidden md:pt-0">
                <div className="skeleton h-4 w-full bg-card"></div>
              </div>
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

function ReloadHistoryButton({ clearHistory }: { clearHistory: () => void }) {
  return (
    <div className="flex justify-end">
      <Button
        id="reload-history-btn"
        variant="link"
        size="sm"
        className="font-light normal-case text-gray-200 no-underline opacity-60 hover:text-primary hover:opacity-100"
        onClick={() => {
          clearHistory();
        }}
      >
        Reload history
      </Button>
    </div>
  );
}

function TransactionGroup({ date, transactions }: { date: string; transactions: TransactionHistory[] }) {
  return (
    <div className="flex flex-col gap-2">
      <span className="block text-base-content">{formatDate(date, "PPP")}</span>
      {transactions.map((transaction, transactionIndex) => {
        if (transaction.messages && transaction.messages.length > 0 && transaction.messages[0].status) {
          const { messages, ...bridgingTransaction } = transaction;
          return (
            <TransactionItem
              key={`transaction-group-${date}-item-${transactionIndex}`}
              transaction={bridgingTransaction}
              message={messages[0]}
            />
          );
        }
      })}
    </div>
  );
}

export function Transactions() {
  const { data: blockNumber } = useBlockNumber({
    watch: true,
  });

  // Context
  const { transactions, fetchHistory, isLoading, clearHistory } = useFetchHistory();

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      fetchHistory();
    }
  }, [blockNumber, fetchHistory]);

  const groupedTransactions = useMemo(() => groupByDay(transactions), [transactions]);

  if (isLoading && transactions.length === 0) {
    return <SkeletonLoader />;
  }

  if (transactions.length === 0) {
    return <NoTransactions />;
  }

  return (
    <div className="flex flex-col gap-8 rounded-lg border-2 border-card bg-cardBg p-4">
      <ReloadHistoryButton clearHistory={clearHistory} />
      {Object.keys(groupedTransactions).map((date) => (
        <TransactionGroup key={`transaction-group-${date}`} date={date} transactions={groupedTransactions[date]} />
      ))}
    </div>
  );
}
