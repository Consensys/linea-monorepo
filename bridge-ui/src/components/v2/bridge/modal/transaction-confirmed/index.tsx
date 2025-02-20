import Modal from "@/components/v2/modal";
import styles from "./transaction-confirmed.module.scss";
import Link from "next/link";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
};

export default function TransactionConfirmed({ isModalOpen, onCloseModal }: Props) {
  return (
    <Modal title="Transaction confirmed!" isOpen={isModalOpen} onClose={onCloseModal} size="lg">
      <div className={styles["modal-inner"]}>
        <p className={styles["text"]}>
          You may now bridge another transaction, check your transaction history, or stay ahead of the curve with the
          latest trending tokens.
        </p>
        <div className={styles["list-button"]}>
          <Link
            className={styles["primary-btn"]}
            href="https://linea.build/ecosystem"
            target="_blank"
            rel="noopenner noreferrer"
          >
            See What&apos;s Trending
          </Link>

          <Link className={styles["secondary-btn"]} href="/transactions">
            View transactions
          </Link>
        </div>
      </div>
    </Modal>
  );
}
