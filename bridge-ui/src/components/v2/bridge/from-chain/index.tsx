import Image from "next/image";
import styles from "./from-chain.module.scss";
import SelectNetwork from "@/components/v2/bridge/modal/select-network";
import { useState } from "react";

export default function FromChain() {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const openModal = () => setIsModalOpen(true);
  const closeModal = () => setIsModalOpen(false);

  return (
    <>
      <button onClick={openModal} className={styles["from"]} type="button">
        <Image src="/images/logo/ethereum-rounded.svg" width="40" height="40" alt="eth" />
        <div className={styles["info"]}>
          <div className={styles["info-name"]}>From</div>
          <div className={styles["info-value"]}>Ethereum</div>
        </div>
      </button>
      <SelectNetwork isModalOpen={isModalOpen} onCloseModal={closeModal} />
    </>
  );
}
