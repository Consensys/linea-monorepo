import InternalNav from "@/components/internal-nav";
import BridgeForm from "@/components/bridge/bridge-layout";
import styles from "./page.module.scss";

export default function Home() {
  return (
    <section className={styles["content-wrapper"]}>
      <InternalNav />
      <BridgeForm />
    </section>
  );
}
