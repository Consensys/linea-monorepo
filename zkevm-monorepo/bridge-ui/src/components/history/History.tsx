'use client';

import { useContext, useEffect, useRef } from 'react';
import { useBlockNumber } from 'wagmi';
import { toast } from 'react-toastify';
import { Variants } from 'framer-motion';

import { HistoryContext } from '@/contexts/history.context';
import HistoryItem from './HistoryItem';

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
  const historyContext = useContext(HistoryContext);
  const { transactions, fetchHistory, isLoading, clearHistory } = historyContext;

  useEffect(() => {
    if (blockNumber && blockNumber % BigInt(2) === BigInt(0)) {
      fetchHistory();
    }
  }, [blockNumber]);

  return (
    <div className="card w-full md:w-[500px] bg-base-100 shadow-xl">
      <div className="card-body">
        <h2 className="justify-between mb-5 font-medium text-white card-title">
          <span>Recent transactions</span>
          <span className="text-sm font-normal">({transactions.length})</span>
        </h2>
        <div
          className="max-h-[545px]  scrollbar-thin scrollbar-thumb-gray-500 scrollbar-track-gray-700 overflow-y-auto pr-4"
          id="transactions-list"
        >
          {transactions.map((transaction) => (
            <HistoryItem key={transaction.transactionHash} transaction={transaction} variants={variants} />
          ))}
        </div>

        {!transactions.length && (
          <div className="flex flex-row items-center justify-center w-full h-16 mb-5 text-sm italic text-center">
            <span>No recent transactions</span>
            {isLoading ? <div className="mt-6 mb-3 loading loading-dots loading-xs"></div> : <span>.</span>}
          </div>
        )}

        <div className="flex justify-end">
          <button
            id="reload-history-btn"
            className="font-light text-gray-200 no-underline normal-case btn-link btn-sm hover:text-primary opacity-60 hover:opacity-100"
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
            <div className="justify-between modal-action">
              <button
                id="reload-history-confirm-btn"
                className="rounded-full btn btn-warning btn-sm"
                onClick={() => {
                  clearHistory();
                  clearHistoryModalRef.current?.close();
                  toast.success('History cleared');
                }}
              >
                Reload history
              </button>
              <button
                id="reload-history-cancel-btn"
                className="rounded-full btn btn-sm"
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
