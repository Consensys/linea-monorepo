"use client";

import { useContext } from "react";
import { formatEther } from "viem";
import { ModalContext } from "@/contexts/modal.context";
import StatusText from "./StatusText";
import TransactionProgressBar from "./TransactionProgressBar";
import TransactionDetailsModal from "./modals/TransactionDetailsModal";
import { NETWORK_ID_TO_NAME } from "@/utils/constants";
import { getChainNetworkLayerByChainId } from "@/utils/chainsUtil";
import { TransactionHistory } from "@/models/history";
import { MessageWithStatus } from "@/hooks";

export enum TransactionStatus {
  READY_TO_CLAIM = "READY_TO_CLAIM",
  PENDING = "PENDING",
  COMPLETED = "COMPLETED",
}

type TransactionItemProps = {
  transaction: TransactionHistory;
  message: MessageWithStatus;
};

export default function TransactionItem({ transaction, message }: TransactionItemProps) {
  const { handleShow } = useContext(ModalContext);

  return (
    <div
      className="grid grid-cols-1 items-center gap-0 rounded-lg bg-[#2D2D2D] p-4 text-[#C0C0C0] hover:cursor-pointer hover:outline hover:outline-1 hover:outline-primary sm:grid-cols-1 md:grid-cols-6 md:gap-4"
      onClick={() => {
        handleShow(<TransactionDetailsModal transaction={transaction} message={message} />);
      }}
    >
      <div className="grid grid-cols-2 gap-4 border-b border-card py-4 md:col-span-2 md:border-none md:p-0">
        <div className="px-6 md:px-0">
          <div className="text-xs uppercase">Status</div>
          <StatusText status={message.status} />
        </div>

        <div className="px-6 md:px-0">
          <div className="text-xs uppercase">From</div>
          <span>{NETWORK_ID_TO_NAME[transaction.fromChain.id]}</span>
        </div>
      </div>

      <div className="hidden px-6 md:col-span-2 md:block md:border-x md:border-card">
        <TransactionProgressBar
          status={message.status}
          transactionTimestamp={transaction.timestamp}
          fromChain={getChainNetworkLayerByChainId(transaction.fromChain.id)}
        />
      </div>

      <div className="grid grid-cols-2 items-center gap-4 border-b border-card py-4 md:col-span-2 md:border-none md:p-0">
        <div className="px-6 md:px-0">
          <div className="text-xs uppercase">To</div>
          <span>{NETWORK_ID_TO_NAME[transaction.toChain.id]}</span>
        </div>

        <div className="px-6 md:px-0">
          <div className="text-xs uppercase">Amount</div>
          <span className="font-bold text-white">
            {formatEther(transaction.amount)} {transaction.token.symbol}
          </span>
        </div>
      </div>

      <div className="px-6 pt-4 md:hidden md:pt-0">
        <TransactionProgressBar
          status={message.status}
          transactionTimestamp={transaction.timestamp}
          fromChain={getChainNetworkLayerByChainId(transaction.fromChain.id)}
        />
      </div>
    </div>
  );
}
