"use client";

import { useEffect, useMemo } from "react";
import { useBlockNumber } from "wagmi";
import { toast } from "react-toastify";
import TransactionItem from "./TransactionItem";
import { TransactionHistory } from "@/models/history";
import { formatDate, fromUnixTime } from "date-fns";
import { NoTransactions } from "./NoTransaction";
import useFetchHistory from "@/hooks/useFetchHistory";

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
    return (
      <div className="flex flex-col gap-4 border-2 border-card bg-cardBg p-4">
        <div className="skeleton h-4 w-full bg-card"></div>
        <div className="skeleton h-4 w-full bg-card"></div>
        <div className="skeleton h-4 w-full bg-card"></div>
        <div className="skeleton h-4 w-full bg-card"></div>
        <div className="skeleton h-4 w-full bg-card"></div>
        <div className="skeleton h-4 w-full bg-card"></div>
        <div className="skeleton h-4 w-full bg-card"></div>
      </div>
    );
  }

  if (transactions.length === 0) {
    return <NoTransactions />;
  }

  return (
    <div className="flex flex-col gap-8 rounded-lg border-2 border-card bg-cardBg p-4">
      <div className="flex justify-end">
        <button
          id="reload-history-btn"
          className="btn-link btn-sm font-light normal-case text-gray-200 no-underline opacity-60 hover:text-primary hover:opacity-100"
          onClick={() => {
            clearHistory();
            toast.success("History cleared");
          }}
        >
          Reload history
        </button>
      </div>
      {Object.keys(groupedTransactions).map((date, groupIndex) => (
        <div key={`transaction-group-${groupIndex}`} className="flex flex-col gap-2">
          <span className="block text-base-content">{formatDate(date, "PPP")}</span>
          {groupedTransactions[date].map((transaction, transactionIndex) => {
            if (transaction.messages && transaction.messages.length > 0 && transaction.messages[0].status) {
              const { messages, ...bridgingTransaction } = transaction;
              return (
                <TransactionItem
                  key={`transaction-group-${groupIndex}-item-${transactionIndex}`}
                  transaction={bridgingTransaction}
                  message={messages[0]}
                />
              );
            }
          })}
        </div>
      ))}
    </div>
  );
}
