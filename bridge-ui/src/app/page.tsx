"use client";

import InternalNav from "@/components/internal-nav";
import BridgeLayout from "@/components/bridge/bridge-layout";
import styles from "./page.module.scss";
import TopBanner from "@/components/top-banner";

export default function Home() {
  return (
    <>
      <TopBanner
        text="The Linea Mainnet USDC bridge has been paused on Sunday 16th of March 20:00 UTC for an upgrade and will
              remain paused until CCTP V2 integration is complete. All pending messages were automatically claimed.
              To bridge USDC between Linea and Ethereum, you can use alternative bridge providers.
              Linea Sepolia (Testnet) is currently being upgraded to support CCTP V2."
        href="https://www.circle.com/blog/linea-to-become-the-first-bridged-usdc-standard-blockchain-to-upgrade-to-native-usdc"
      />
      <section className={styles["content-wrapper"]}>
        <InternalNav />
        <BridgeLayout />
      </section>
    </>
  );
}
