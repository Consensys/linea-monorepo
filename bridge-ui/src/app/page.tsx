"use client";

import styles from "./page.module.scss";
import dynamic from "next/dynamic";

const Widget = dynamic(() => import("@/components/lifi/widget").then((mod) => mod.Widget), {
  ssr: false,
  loading: () => <p>Loading...</p>,
});

export default function Page() {
  return (
    <section className={styles["content-wrapper"]}>
      <Widget />
    </section>
  );
}
