import styles from "./list-transaction.module.scss";
import Transaction from "./item";
import TransactionDetails from "@/components/v2/transaction/modal/transaction-details";
import { useState } from "react";
import { TransactionStatus } from "@/components/transactions/TransactionItem";

export type Transaction = {
  code: string;
  value: string;
  date: string;
  unit: string;
  estimatedTime?: string;
  status: TransactionStatus;
};

type Props = {
  transactions: Transaction[];
};

export default function ListTransaction({ transactions }: Props) {
  const [currentTransaction, setCurrentTransaction] = useState<Transaction | undefined>(false || undefined);
  const handleCloseModal = () => {
    setCurrentTransaction(undefined);
  };
  const handleClickTransaction = (code: string) => {
    const transaction = transactions.find((t) => t.code === code);
    if (transaction) {
      setCurrentTransaction(transaction);
    }
  };
  return (
    <>
      <ul className={styles["list"]}>
        {transactions.map((item) => (
          <Transaction key={item.code} onClick={handleClickTransaction} {...item} />
        ))}
      </ul>
      <TransactionDetails
        transaction={currentTransaction}
        isModalOpen={!!currentTransaction}
        onCloseModal={handleCloseModal}
      />
    </>
  );
}
