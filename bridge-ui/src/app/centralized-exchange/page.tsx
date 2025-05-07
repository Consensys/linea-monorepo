import { Widget } from "@/components/layerswap/widget";
import styles from "./page.module.scss";
import FaqHelp from "@/components/bridge/faq-help";

export default function Page() {
  return (
    <>
      <section className={styles["content-wrapper"]}>
        <Widget />
      </section>
      <FaqHelp />
    </>
  );
}
