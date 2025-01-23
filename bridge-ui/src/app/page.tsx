import InternalNav from "@/components/v2/internal-nav";
import BridgeForm from "@/components/v2/bridge/bridge-layout";
import styles from "./page.module.scss";

export default function Home() {
  return (
    <>
      <section className={styles["content-wrapper"]}>
        <InternalNav />
        <BridgeForm />
      </section>
    </>
  );
}
