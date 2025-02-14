"use client";

import styles from "./page.module.scss";
import InternalNav from "@/components/v2/internal-nav";
import TopBanner from "@/components/v2/top-banner";
import Transactions from "@/components/v2/transaction/transactions";

export default function TransactionsPage() {
  return (
    <>
      <TopBanner
        href="https://linea.build/ecosystem"
        text="⭐️ Stay ahead of the curve with the latest trending tokens - Discover trending tokens"
      />
      <div className={styles["content-wrapper"]}>
        <InternalNav />
        <Transactions />
      </div>
    </>
  );
}
