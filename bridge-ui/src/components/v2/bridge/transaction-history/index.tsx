import { useAccount } from "wagmi";
import styles from "./transaction-history.module.scss";
import ArrowLeftIcon from "@/assets/icons/arrow-left.svg";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import Button from "../../ui/button";
import { TransactionStatus } from "@/components/transactions/TransactionItem";
import ListTransaction from "./list-transaction";
import NoTransaction from "./no-transaction";
import TransactionNotConnected from "./transaction-not-connected";

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

export default function TransactionHistory() {
  const { isConnected } = useAccount();
  const setIsTransactionHistoryOpen = useNativeBridgeNavigationStore.useSetIsTransactionHistoryOpen();
  const setIsBridgeOpen = useNativeBridgeNavigationStore.useSetIsBridgeOpen();

  return (
    <div className={styles["transaction-history-wrapper"]}>
      <div className={styles.headline}>
        <div className={styles["action"]}>
          <Button
            variant="link"
            onClick={() => {
              setIsTransactionHistoryOpen(false);
              setIsBridgeOpen(true);
            }}
          >
            <ArrowLeftIcon className={styles["go-back-icon"]} />
          </Button>
        </div>
        <h2 className={styles.title}>Transaction History</h2>
      </div>

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
  );
}
