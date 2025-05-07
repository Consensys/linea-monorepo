"use client";

import LayerSwapWidget from "@/components/layerswap";
import styles from "./page.module.scss";
import FaqHelp from "@/components/bridge/faq-help";

export default function Page() {
  return (
    <>
      <section className={styles["content-wrapper"]}>
        <LayerSwapWidget />
      </section>
      <FaqHelp />
    </>
  );
}
