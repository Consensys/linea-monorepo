import styles from "./no-transaction.module.scss";
import Link from "next/link";
import TransactionCircleIcon from "@/assets/icons/transaction-circle.svg";

export default function NoTransaction() {
  return (
    <div className={styles["content"]}>
      <span className={styles["icon"]}>
        <TransactionCircleIcon />
      </span>
      <p className={styles["title"]}>No transactions yet</p>
      <p className={styles["desc"]}>Use the bridge and view your transactions and their status here.</p>
      <Link className={styles["link"]} href="/">
        Bridge assets
      </Link>
    </div>
  );
}
