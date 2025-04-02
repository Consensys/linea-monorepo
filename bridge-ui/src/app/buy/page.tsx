"use client";

import OnRamperWidget from "@/components/onramper";
import styles from "./page.module.scss";

export default function Page() {
  return (
    <section className={styles["content-wrapper"]}>
      <OnRamperWidget />
    </section>
  );
}
