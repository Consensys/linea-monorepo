"use client";

import { useEffect, useRef } from "react";
import { useBlockNumber } from "wagmi";
import { toast } from "react-toastify";
import { Variants } from "framer-motion";

import HistoryItem from "./HistoryItem";
import useFetchHistory from "@/hooks/useFetchHistory";

export type BridgeForm = {
  amount: string;
};

const variants: Variants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      duration: 0.5,
    },
  },
};

export default function History() {
  const clearHistoryModalRef = useRef<HTMLDialogElement>(null);

  // Wagmi
  const { data: blockNumber } = useBlockNumber({
    watch: true,
  });

  // Context
  const { transactions, fetchHistory, isLoading, clearHistory } = useFetchHistory();

  useEffect(() => {
    if (blockNumber && blockNumber % BigInt(2) === BigInt(0)) {
      fetchHistory();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber]);

  return (
    <div className="card w-full bg-base-100 shadow-xl md:w-[500px]">
      <div className="card-body">
        <h2 className="card-title mb-5 justify-between font-medium text-white">
          <span>Recent transactions</span>
          <span className="text-sm font-normal">({transactions.length})</span>
        </h2>
        <div
          className="max-h-[545px]  overflow-y-auto pr-4 scrollbar-thin scrollbar-track-gray-700 scrollbar-thumb-gray-500"
          id="transactions-list"
        >
          {transactions.map((transaction) => (
            <HistoryItem key={transaction.transactionHash} transaction={transaction} variants={variants} />
          ))}
        </div>

        {!transactions.length && (
          <div className="mb-5 flex h-16 w-full flex-row items-center justify-center text-center text-sm italic">
            <span>No recent transactions</span>
            {isLoading ? <div className="loading loading-dots loading-xs mb-3 mt-6"></div> : <span>.</span>}
          </div>
        )}

        <div className="flex justify-end">
          <button
            id="reload-history-btn"
            className="btn-link btn-sm font-light normal-case text-gray-200 no-underline opacity-60 hover:text-primary hover:opacity-100"
            onClick={() => clearHistoryModalRef.current?.showModal()}
          >
            Reload history
          </button>
        </div>
        <dialog ref={clearHistoryModalRef} id="clear_history_modal" className="modal">
          <form method="dialog" className="modal-box">
            <h3 className="text-lg font-bold">Are you sure?</h3>
            <p className="py-4">
              This might be necessary if you&apos;re encountering issues and need to reset the synchronization process.
              However, rebuilding your history could take a considerable amount of time.
            </p>
            <p className="py-4">Proceed with caution and only if absolutely necessary.</p>
            <div className="modal-action justify-between">
              <button
                id="reload-history-confirm-btn"
                className="btn btn-warning btn-sm rounded-full"
                onClick={() => {
                  clearHistory();
                  clearHistoryModalRef.current?.close();
                  toast.success("History cleared");
                }}
              >
                Reload history
              </button>
              <button
                id="reload-history-cancel-btn"
                className="btn btn-sm rounded-full"
                onClick={() => clearHistoryModalRef.current?.close()}
              >
                Cancel
              </button>
            </div>
          </form>
          <form method="dialog" className="modal-backdrop">
            <button id="close-history-reload-btn">close</button>
          </form>
        </dialog>
      </div>
    </div>
  );
}
