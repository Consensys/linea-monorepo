import { TransactionStatus } from "@/components/transactions/TransactionItem";
import { useAccount } from "wagmi";
import NoTransaction from "@/components/v2/transaction/no-transaction";
import ListTransaction from "@/components/v2/transaction/list-transaction";
import TransactionNotConnected from "@/components/v2/transaction/transaction-not-connected";
import SkeletonLoader from "@/components/v2/transaction/skeleton-loader";

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

export default function Transactions() {
  const { isConnected } = useAccount();
  const isLoading = false;

  if (isLoading) {
    return <SkeletonLoader />;
  }

  return (
    <>
      {isConnected ? (
        listTransaction?.length ? (
          <ListTransaction transactions={listTransaction} />
        ) : (
          <NoTransaction />
        )
      ) : (
        <TransactionNotConnected />
      )}
    </>
  );
}
