import Image from "next/image";
import styles from "./to-chain.module.scss";
import SelectNetwork from "@/components/v2/bridge/modal/select-network";
import { useState } from "react";

export default function ToChain() {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const openModal = () => setIsModalOpen(true);
  const closeModal = () => setIsModalOpen(false);

  return (
    <>
      <button onClick={openModal} className={styles["to"]} type="button">
        <div className={styles["name"]}>To</div>
        <div className={styles["info"]}>
          <Image src="/images/logo/linea-rounded.svg" width="40" height="40" alt="eth" />
          <div className={styles["info-value"]}>Ethereum</div>
        </div>
      </button>
      <SelectNetwork isModalOpen={isModalOpen} onCloseModal={closeModal} />
    </>
  );
}
