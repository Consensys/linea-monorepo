import Link from "next/link";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import { formatHex, formatTimestamp } from "@/utils/format";
import { NETWORK_ID_TO_NAME } from "@/utils/constants";
import { MessageWithStatus } from "@/hooks";
import { TransactionHistory } from "@/models/history";
import TransactionClaimButton from "./TransactionClaimButton";

type TransactionDetailsModalProps = {
  transaction: TransactionHistory;
  message: MessageWithStatus;
};

const BlockExplorerLink: React.FC<{
  transactionHash?: string;
  blockExplorer?: string;
}> = ({ transactionHash, blockExplorer }) => {
  if (!transactionHash || !blockExplorer) {
    return <span>N/A</span>;
  }
  return (
    <Link
      href={`${blockExplorer}/tx/${transactionHash}`}
      passHref
      target="_blank"
      rel="noopener noreferrer"
      className="link text-primary"
    >
      {formatHex(transactionHash)}
    </Link>
  );
};

const TransactionDetailsModal: React.FC<TransactionDetailsModalProps> = ({ transaction, message }) => {
  return (
    <div className="flex flex-col gap-8">
      <h2 className="text-xl">Transaction details</h2>
      <div className="space-y-2">
        <div className="flex">
          <label className="w-44 text-[#C0C0C0]">Date & Time</label>
          <span className="text-[#C0C0C0]">{formatTimestamp(Number(transaction.timestamp), "h:mma d MMMM yyyy")}</span>
        </div>

        <div className="flex">
          <label className="w-44 text-[#C0C0C0]">{NETWORK_ID_TO_NAME[transaction.fromChain.id]} Tx Hash</label>
          <BlockExplorerLink
            blockExplorer={transaction.fromChain.blockExplorers?.default.url}
            transactionHash={transaction.transactionHash}
          />
        </div>

        <div className="flex">
          <label className="w-44 text-[#C0C0C0]">{NETWORK_ID_TO_NAME[transaction.toChain.id]} Tx Hash</label>
          <BlockExplorerLink
            blockExplorer={transaction.toChain.blockExplorers?.default.url}
            transactionHash={message.claimingTransactionHash}
          />
        </div>
        <div className="flex">
          <label className="w-44 text-[#C0C0C0]">Fee</label>
          <span className="text-[#C0C0C0]">5 ETH ~$15000</span>
        </div>
      </div>
      {message.status === OnChainMessageStatus.CLAIMABLE && (
        <TransactionClaimButton key={message.messageHash} message={message} transaction={transaction} />
      )}
    </div>
  );
};

export default TransactionDetailsModal;
